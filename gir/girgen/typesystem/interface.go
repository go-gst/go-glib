package typesystem

import (
	"fmt"
	"reflect"

	"github.com/go-gst/go-glib/gir"
)

// Interface declares an interface implemented by a GObject.
type Interface struct {
	Doc
	BaseType
	gir *gir.Interface

	TypeStruct *Record

	// Parent is always (foreign) GObject, because at runtime we will receive a GObject pointer
	// and wrap it. We look it up because we don't know the implementation here.
	Parent CouldBeForeign[*Class]

	GoWrapBaseClassFunction string

	GoPrivateUpcastMethod string

	// GoExtendOverrideStructName is the name of the struct that will be used to to override virtual class methods
	// when extending the class.
	GoExtendOverrideStructName string
	// GoUnsafeApplyOverridesName is the name of the function that will be used to apply the overrides
	// to the gclass.
	GoUnsafeApplyOverridesName string

	GoInterfaceName string

	BaseConversions
	Marshaler

	Prerequesite []CouldBeForeign[Type]

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Signals        []*Signal

	// ManuallyExtended is true if the class is manually extended by the user
	// this will embed an extra (not generated) interface with the naming scheme `<GoInterfaceName>ExtManual` in the classes interface.
	//
	// This must be set by a post-processing step.
	ManuallyExtended bool
}

// GoType implements Type. Use the interface type if a pointer is needed
func (in *Interface) GoType(pointers int) string {
	if pointers == 0 {
		return in.BaseType.GoType(0)
	}

	return in.GoInterfaceName
}

// minPointersRequired implements minPointerConstrainedType.
func (a *Interface) minPointersRequired() int {
	return 1
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Interface) maxPointersAllowed() int {
	return 1
}

func DeclareInterface(e *env, v *gir.Interface) *Interface {
	e = e.sub("interface", v.CType)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	i := &Interface{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name + "Instance",
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,
		},
		Marshaler:       e.newDefaultMarshaler(v.GLibGetType, v.Name),
		GoInterfaceName: v.Name,

		GoWrapBaseClassFunction: fmt.Sprintf("unsafeWrap%s", v.Name),

		GoPrivateUpcastMethod: fmt.Sprintf("upcastTo%s", v.CType), // use cidentifier to not shadow parent methods

		GoExtendOverrideStructName: fmt.Sprintf("%sOverrides", v.Name),
		GoUnsafeApplyOverridesName: fmt.Sprintf("UnsafeApply%sOverrides", v.Name),

		BaseConversions: BaseConversions{
			FromGlibBorrowFunction: fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name), // borrow is needed for subclassing
			FromGlibNoneFunction:   fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
			FromGlibFullFunction:   fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

			ToGlibNoneFunction: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
			ToGlibFullFunction: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),
		},

		gir: v,
	}

	return i
}

func (in *Interface) resolve(e *env) resolvedState {
	e = e.sub("interface", in.gir.CType)

	if in.gir.GLibTypeStruct != "" {
		ns, typeStructType := e.findTypeByGIRName(in.gir.GLibTypeStruct)

		if ns != nil {
			e.logger.Warn("type struct is foreign", "namespace", ns.Name)
			return notResolvable
		}

		if typeStructType == nil {
			return notResolvable
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			e.logger.Warn("type struct is not a record", "actual", reflect.TypeOf(typeStructType).String())
			return notResolvable
		}

		in.TypeStruct = typeStruct
	}

	ns, parent := e.findTypeByGIRName("GObject.Object")

	if parent == nil {
		e.logger.Error("GObject.Object not found")
		return notResolvable
	}

	if _, ok := parent.(*Class); !ok {
		return notResolvable
	}

	in.Parent = CouldBeForeign[*Class]{
		Namespace: ns,
		Type:      parent.(*Class),
	}

	for _, prereq := range in.gir.Prerequisites {
		ns, inter := e.findTypeByGIRName(prereq.Name)

		if inter == nil {
			e.logger.Info("interface prerequesite not found", "interface", prereq.Name)
			return maybeResolvable
		}

		switch inter.(type) {
		case *Class, *Interface:
		default:
			e.logger.Warn("prerequesite is not a class or an interface", "prerequesite", inter.GIRName(), "actual", reflect.TypeOf(parent).String())
			return notResolvable
		}

		in.Prerequesite = append(in.Prerequesite, CouldBeForeign[Type]{
			Namespace: ns,
			Type:      inter,
		})
	}

	return okResolved
}

func (in *Interface) declareNested(e *env) {
	e = e.sub("interface", in.gir.CType)

	for _, v := range in.gir.Functions {
		if t := DeclarePrefixedFunction(e, in, v.CallableAttrs); t != nil {
			in.Functions = append(in.Functions, t)
		}
	}

	for _, v := range in.gir.Methods {
		if t := DeclareMethod(e, in, v); t != nil {
			in.Methods = append(in.Methods, t)
		}
	}

	for _, v := range in.gir.Signals {
		if t := NewSignal(e, in, v); t != nil {
			in.Signals = append(in.Signals, t)
		}
	}

	if in.TypeStruct != nil {
		for _, v := range in.gir.VirtualMethods {
			if t := NewVirtualMethod(e, in, in.TypeStruct, v); t != nil {
				in.VirtualMethods = append(in.VirtualMethods, t)
			}
		}
	}
}
