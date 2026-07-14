package typesystem

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/go-gst/go-glib/gir"
)

type ParamCompareFunc func(a, b *Param) int

type env struct {
	cfg       Config
	nsCfg     NamespaceConfig
	namespace *Namespace

	minVersion gir.Version
	maxVersion gir.Version

	// user overridable settings via [Config]:

	ignore IgnoreFunc

	logger *slog.Logger

	symbolPrefixes     []string
	identifierPrefixes []string
}

// sub returns a sub env that is an exact copy but with the given attrs used in the logger
func (e *env) sub(attrs ...any) *env {
	subenv := *e

	logger := subenv.logger.With(attrs...)
	subenv.logger = logger

	return &subenv
}

// sortGoParams moves the params according to go conventions:
// 1. context.Context is always first
// 2. error is always last
// 3. all other params don't get reordered
func (e *env) sortGoParams(ps []*Param) {
	slices.SortFunc(ps, func(a, b *Param) int {
		if a.GoType() == "context.Context" {
			return -1
		}
		if b.GoType() == "context.Context" {
			return 1
		}
		if a.GoType() == "error" {
			return 1
		}
		if b.GoType() == "error" {
			return -1
		}
		return 0
	})
}

// sortGoReturns moves the return values according to go conventions:
// 1. error is always last
// 2. all other return values don't get reordered
func (e *env) sortGoReturns(ps []*Param) {
	slices.SortFunc(ps, func(a, b *Param) int {
		if a.GoType() == "error" {
			return 1
		}
		if b.GoType() == "error" {
			return -1
		}
		return 0
	})
}

func (e *env) trampolinePrefix() string {
	return fmt.Sprintf("_goglib_%s%d", e.namespace.GoName, e.namespace.Version.Major)
}

// skip returns true if the gir type/identifier should be skipped. Any optional parent can be passed
// to handle nested gir identifiers
func (e *env) skip(parent Type, anygir any) bool {
	name, attrs, elements := infoFromAnyGir(anygir)

	if e.ingoreDeprecated(name, attrs) {
		return true
	}

	if e.ignoreTooNew(name, attrs) {
		return true
	}

	for _, m := range e.namespace.Manual {
		if m.GIRName() == name {
			e.logger.Info("skipping manually implemented type", "name", name)
			return true
		}
	}

	var parentName string
	if parent != nil {
		parentName = parent.GIRName()
	}

	return e.ignore(parentName, name, attrs, elements)
}

type girWithInfoAttrs interface {
	GetInfoAttrs() gir.InfoAttrs
}

type girWithInfoElements interface {
	GetInfoElements() gir.InfoElements
}

func (e *env) ingoreDeprecated(name string, attrs gir.InfoAttrs) bool {
	var zeroVersion gir.Version
	if attrs.Deprecated && e.minVersion != zeroVersion && attrs.DeprecatedVersion.LessEqual(e.minVersion) {
		e.logger.Info("skipping deprecated", "name", name, "deprecated-since", attrs.DeprecatedVersion, "min-version", e.minVersion)
		return true
	}

	return false
}

func (e *env) ignoreTooNew(name string, attrs gir.InfoAttrs) bool {
	var zeroVersion gir.Version
	if attrs.Version != zeroVersion && e.maxVersion != zeroVersion && e.maxVersion.Less(attrs.Version) {
		e.logger.Info("skipping too new", "name", name, "introduced-since", attrs.Version, "max-version", e.maxVersion)
		return true
	}

	return false
}

func (e *env) findAnyType(t gir.AnyType) (*Namespace, Type) {
	if t.Type != nil && t.Array != nil {
		panic("received invalid anytype")
	}

	if t.Type != nil {
		ns, typ := e.findType(t.Type)

		if typ == nil {
			return nil, nil
		}
		return ns, typ
	}

	if t.Array != nil {
		arr := e.getArrayType(t.Array)

		if arr == nil {
			return nil, nil
		}
		return nil, arr
	}

	// this happens e.g. on vararg params
	return nil, nil
}

// referencedNamespace returns the namespace of the type
// it has a fallback for the own namespace, and signifies this with
// the foreign boolean
func (e *env) referencedNamespace(t string) (foreign bool, ns *Namespace, localtype string) {
	parts := strings.Split(t, ".")

	if len(parts) > 2 {
		panic("invalid type name received")
	}

	if len(parts) == 1 {
		return false, e.namespace, t
	}

	// fallback because sometimes the type references the own namespace
	if parts[0] == e.namespace.Name {
		return false, e.namespace, parts[1]
	}

	reffedNS := e.namespace.Included[parts[0]]

	if reffedNS == nil {
		e.logger.Warn("type referenced unknown namespace", "type", t, "referenced-ns", parts[0])
		return false, nil, ""
	}

	return true, reffedNS, parts[1]
}

// findType searches for a declared type in the namespace
func (e *env) findType(t *gir.Type) (*Namespace, Type) {
	ns, typ := e.findOuterType(t)

	container, ok := typ.(*Container)

	if len(t.Types) == 0 && !ok {
		// resolve the non container type
		return ns, typ
	}

	// typ has to be a container type
	if !ok {
		e.logger.Warn("type has inner types but is not a container", "type", t.Name, "ctype", t.CType, "inner-types", t.Types)
		return nil, nil
	}

	typ = e.resolveContainerInnerTypes(container, t.Types)

	if typ == nil {
		return nil, nil
	}

	return ns, typ
}

// findOuterType searches for a declared type in the namespace, it ignores any inner types
func (e *env) findOuterType(t *gir.Type) (*Namespace, Type) {
	// Ctype resolving is often more reliable than GIR name resolving, so we try it first

	typename := t.Name
	ctype := cleanCType(t.CType)

	if ctypeIsIncompatible(ctype) {
		return nil, nil
	}

	if replaced, ok := e.cfg.GIRReplacements[typename]; ok {
		e.logger.Info("replacing GIR type name", "type", t.Name, "replaced by", replaced)
		typename = replaced
	}

	isForeign, ns, girName := e.referencedNamespace(typename)

	if ns == nil {
		return nil, nil
	}

	typ := findBuiltinPrimitiveByCType(t.CType)

	if typ != nil {
		return nil, typ
	}

	typ = findBuiltinPrimitiveByGIRName(girName)

	if typ != nil {
		return nil, typ
	}

	// foreign namespace is only non nil if the type is
	// found in another namespace
	var foreignNS *Namespace
	if isForeign {
		foreignNS = ns
	}

	typ = ns.findLocalTypeByCType(ctype)

	if typ != nil {
		return foreignNS, typ
	}

	typ = ns.FindLocalTypeByGIRName(girName)

	if typ != nil {
		return foreignNS, typ
	}

	e.logger.Warn("type not found", "type", t.Name, "ctype", t.CType)

	return nil, nil
}

func (e *env) findTypeByGIRName(t string) (*Namespace, Type) {
	if replaced, ok := e.cfg.GIRReplacements[t]; ok {
		e.logger.Info("replacing GIR type name", "type", t, "replaced by", replaced)
		t = replaced
	}

	if ctypeIsIncompatible(t) {
		return nil, nil
	}

	isForeign, ns, girName := e.referencedNamespace(t)

	if ns == nil {
		return nil, nil
	}
	typ := findBuiltinPrimitiveByGIRName(girName)

	if typ != nil {
		return nil, typ
	}

	// foreign namespace is only non nil if the type is
	// found in another namespace
	var foreignNS *Namespace
	if isForeign {
		foreignNS = ns
	}

	typ = ns.FindLocalTypeByGIRName(girName)

	if typ != nil {
		return foreignNS, typ
	}

	e.logger.Warn("type not found by GIR name", "type", t)

	return nil, nil
}

// identifierToGo converts the C identifier to a Go identifier. the identifier prefix of the namespace is stripped.
func (e *env) identifierToGo(identifier string) string {
	for _, p := range e.identifierPrefixes {
		trimmed, ok := strings.CutPrefix(identifier, p)

		trimmed, _ = strings.CutPrefix(trimmed, "_")

		if ok {
			return trimmed
		}
	}

	slog.Warn("identifier does not have a prefix, using as is", "namespace", e.namespace.v, "identifier", identifier)
	return identifier
}
