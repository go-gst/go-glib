package typesystem

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

type Record struct {
	Doc
	BaseType

	gir *gir.Record

	Fields []*Field

	// PrivateGoType is the inner struct that contains the C pointer and gets the finalizer attached
	PrivateGoType string

	BaseConversions
	Marshaler

	GoUnsafeRefFunction      string
	CgoRefFunction           string
	CgoRefNeedsUnsafePointer bool

	// GoCopyMethod is the go method that copies the record. It takes the struct pointer as an argument and
	// returns a new struct pointer. This is needed for transfer:none
	GoCopyMethod *CallableSignature

	GoUnsafeUnrefFunction   string
	CgoUnrefFunction        string
	CgoUnrefNeedsUnsafeCast bool

	// IsTypeStructFor contains the (foreign) Class that this struct is a type struct for.
	// This is used to figure out where in the type hierarchy the struct is. That way we can
	// generate a cast-to-parent method for the struct.
	IsTypeStructFor *Class

	Functions    []*CallableSignature
	Methods      []*CallableSignature
	Constructors []*CallableSignature

	// TODO:
	Unions     []*Union
	Properties []*struct{}
}

func DeclareRecord(e *env, v *gir.Record) *Record {
	e = e.sub("record", v.CType)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	if strings.HasSuffix(v.Name, "Private") {
		return nil
	}

	gotyp := e.identifierToGo(v.CType)

	return &Record{
		Doc:           NewDoc(&v.InfoAttrs, &v.InfoElements),
		PrivateGoType: strcases.Unexport(gotyp),

		BaseConversions: newDefaultBaseConversions(v.Name),

		GoUnsafeUnrefFunction:   fmt.Sprintf("Unsafe%sFree", gotyp),
		CgoUnrefFunction:        "C.free", // replaced below if an unref or custom free method is found
		CgoUnrefNeedsUnsafeCast: true,

		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   gotyp,
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		Marshaler: e.newDefaultMarshaler(v.GLibGetType, gotyp),

		gir: v,
	}
}

// markAsTypestructFor marks the record as a type struct for the given class in the current namespace.
func (r *Record) markAsTypestructFor(e *env, c *Class) {
	if r.IsTypeStructFor != nil {
		e.logger.Warn("record is already marked as a type struct for another class", "record", r.GirName, "type-struct-for", r.IsTypeStructFor.CType(0))
		return
	}

	// we don't want any methods using typestructs as parameters where
	// we drop the ownership
	r.BaseConversions.FromGlibFullFunction = ""
	r.BaseConversions.FromGlibNoneFunction = ""
	r.BaseConversions.ToGlibFullFunction = ""
	// keep the borrow function, we need it to wrap the type struct
	// keep the to none functions, because we need them for instance params

	r.IsTypeStructFor = c
}

func (r *Record) declareNested(e *env) {
	e = e.sub("record", r.CType(0))

	for _, v := range r.gir.Functions {
		if t := DeclarePrefixedFunction(e, r, v.CallableAttrs); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range r.gir.Methods {
		if v.Name == "weak_ref" || v.Name == "weak_unref" {
			// we don't want the user to be able to weakly reference the object
			// as there are better tools for this and this will only cause problems
			continue
		}

		// see https://github.com/gtk-rs/gir/blob/87cddb70c739f25edd8047e6780e3934af8ff474/src/library.rs#L459-L481

		if v.Name == "ref" {
			r.GoUnsafeRefFunction = fmt.Sprintf("Unsafe%sRef", r.GoType(0))
			r.CgoRefFunction = "C." + v.CIdentifier
			continue
		}

		if v.Name == "unref" {
			r.GoUnsafeUnrefFunction = fmt.Sprintf("Unsafe%sUnref", r.GoType(0))
			r.CgoUnrefFunction = "C." + v.CIdentifier
			r.CgoUnrefNeedsUnsafeCast = false
			continue
		}

		if v.Name == "free" {
			r.GoUnsafeUnrefFunction = fmt.Sprintf("Unsafe%sFree", r.GoType(0))
			r.CgoUnrefFunction = "C." + v.CIdentifier
			r.CgoUnrefNeedsUnsafeCast = false
			continue
		}

		if v.Name == "destroy" {
			r.GoUnsafeUnrefFunction = fmt.Sprintf("Unsafe%sDestroy", r.GoType(0))
			r.CgoUnrefFunction = "C." + v.CIdentifier
			r.CgoUnrefNeedsUnsafeCast = false
			continue
		}

		if t := DeclareMethod(e, r, v); t != nil {
			r.Methods = append(r.Methods, t)

			if t.GoIndentifier() == "Copy" &&
				len(t.GoParameters) == 0 &&
				t.InstanceParam != nil && t.InstanceParam.Type.Type == r &&
				len(t.GoReturns) == 1 && t.GoReturns[0].TransferOwnership == TransferFull && t.GoReturns[0].Type.Type == r {

				r.GoCopyMethod = t
			}

			// TODO: copy_into method
		}
	}

	for _, v := range r.gir.Constructors {
		if t := DeclarePrefixedFunction(e, r, v.CallableAttrs); t != nil {
			r.Constructors = append(r.Constructors, t)
		}
	}

	// Disguised means opaque, so we're not supposed to access these fields.
	if !r.gir.Disguised {
		for _, v := range r.gir.Fields {
			if t := NewField(e, r, v); t != nil {
				r.Fields = append(r.Fields, t)
			}
		}
	}
}

// ParentTypeStruct resolves the parent classes type struct. This panics if the
// record is not a type struct.
func (r *Record) ParentTypeStruct() *CouldBeForeign[*Record] {
	if r.IsTypeStructFor == nil {
		return nil
	}

	if r.IsTypeStructFor.Parent.Type == nil {
		return nil
	}

	parentTs := r.IsTypeStructFor.Parent.Type.TypeStruct

	if parentTs == nil {
		return nil
	}

	return &CouldBeForeign[*Record]{
		// if the parent is foreign, so is the type struct:
		Namespace: r.IsTypeStructFor.Parent.Namespace,
		Type:      parentTs,
	}
}

// minPointersRequired implements Type.
func (a *Record) minPointersRequired() int {
	return 1
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Record) maxPointersAllowed() int {
	return 1
}
