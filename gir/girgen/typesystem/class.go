package typesystem

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/go-gst/go-glib/gir"
)

// Class is a type that extends GObject
type Class struct {
	BaseType

	Doc

	// BaseClass is GObject. It is needed to identify needed foreign references to functions and types
	// if we are in a different namespace (which mostly we are)
	BaseClass CouldBeForeign[*Class]

	// GoInterfaceName is the name of the interface that describes the class. Every extending class will implement this
	// interface. Constructors and methods on the class will use the interface (e.g. MyClasser) name instead of the pointer type
	// (e.g. *MyClass) to allow easy passing of child class types.
	GoInterfaceName string

	GoWrapBaseClassFunction string

	GoPrivateUpcastMethod string

	// GoExtendOverrideStructName is the name of the struct that will be used to to override virtual class methods
	// when extending the class.
	GoExtendOverrideStructName string
	// GoUnsafeApplyOverridesName is the name of the function that will be used to apply the overrides
	// to the gclass.
	GoUnsafeApplyOverridesName string

	// GType is the (foreign) Type that represents GType. It is needed because we need this type for the return value of the
	// GoRegisterSubClassName function.
	GType CouldBeForeign[Type]
	// GoRegisterSubClassName is the name of the function that will be used to register a subclass for this class.
	// it will wrap the gobject base register function (see gobject manual subclassing code)
	GoRegisterSubClassName string

	BaseConversions
	Marshaler

	Abstract bool
	// Final is true if the class is final. This means that it can't be extended by other classes.
	Final bool

	// gir is used to resolve the class and it's nested definitions after it has been declared
	gir *gir.Class

	TypeStruct *Record
	Parent     CouldBeForeign[*Class]
	// Implements contains the implemented (foreign) interfaces
	Implements []CouldBeForeign[*Interface]

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Constructors   []*CallableSignature
	Fields         []*Field
	Signals        []*Signal

	// ManuallyExtended is true if the class is manually extended by the user
	// this will embed an extra (not generated) interface with the naming scheme `<GoInterfaceName>ExtManual` in the classes interface.
	//
	// This must be set by a post-processing step.
	ManuallyExtended bool
}

// GoType implements Type. Use the interface type if a pointer is needed
func (in *Class) GoType(pointers int) string {
	if pointers == 0 {
		return in.BaseType.GoType(0)
	}

	return in.GoInterfaceName
}

// minPointersRequired implements minPointerConstrainedType.
func (a *Class) minPointersRequired() int {
	return 1
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Class) maxPointersAllowed() int {
	return 1
}

func DeclareClass(e *env, v *gir.Class) *Class {
	e = e.sub("class", v.CType)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	if v.Parent == "" {
		// we can't handle non GObject child classes for now
		// FIXME: this should instead register a fundamental type in the namespace
		e.logger.Warn("skipping because it's a fundamental type")
		return nil
	}

	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	c := &Class{
		Doc:             NewDoc(&v.InfoAttrs, &v.InfoElements),
		Abstract:        v.Abstract,
		Final:           false, // overridden after the type struct is resolved
		GoInterfaceName: v.Name,

		GoWrapBaseClassFunction:    fmt.Sprintf("unsafeWrap%s", v.Name),
		GoPrivateUpcastMethod:      fmt.Sprintf("upcastTo%s", v.CType), // use cidentifier to not shadow parent methods
		GoExtendOverrideStructName: fmt.Sprintf("%sOverrides", v.Name),
		GoUnsafeApplyOverridesName: fmt.Sprintf("UnsafeApply%sOverrides", v.Name),
		GoRegisterSubClassName:     fmt.Sprintf("Register%sSubClass", v.Name),

		BaseConversions: BaseConversions{
			FromGlibBorrowFunction: fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name), // borrow function is used for subclassing
			FromGlibNoneFunction:   fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
			FromGlibFullFunction:   fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

			ToGlibNoneFunction: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
			ToGlibFullFunction: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),
		},

		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name + "Instance",
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,
		},
		Marshaler: e.newDefaultMarshaler(v.GLibGetType, v.Name),
		gir:       v,
	}

	return c
}

func (c *Class) resolve(e *env) bool {
	e = e.sub("class", c.gir.CType)

	ns, baseClass := e.findTypeByGIRName("GObject.Object")
	if baseClass == nil {
		panic("Gobject.Object not found")
	}

	if _, ok := baseClass.(*Class); !ok {
		panic("gobject is not a class")
	}

	c.BaseClass = CouldBeForeign[*Class]{
		Namespace: ns,
		Type:      baseClass.(*Class),
	}

	ns, gtype := e.findTypeByGIRName("GType")

	if gtype == nil {
		panic("GType not found")
	}

	c.GType = CouldBeForeign[Type]{
		Namespace: ns,
		Type:      gtype,
	}

	if c.gir.GLibTypeStruct != "" {
		ns, typeStructType := e.findTypeByGIRName(c.gir.GLibTypeStruct)

		if ns != nil {
			e.logger.Warn("type struct is foreign", "namespace", ns.Name)
			return false
		}

		if typeStructType == nil {
			return false
		}

		// FIXME: the typestruct can also be an alias for a record
		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			e.logger.Warn("type struct is not a record", "actual", reflect.TypeOf(typeStructType).String())
			return false
		}

		c.TypeStruct = typeStruct

		typeStruct.markAsTypestructFor(e, c)
	}

	// Final check as in https://gitlab.gnome.org/GNOME/gi-docgen/-/blob/9bec04ee3a294111121c45be77f9b7803206f7f9/gidocgen/gdgenerate.py?page=2#L1548-1564
	c.Final = c.gir.GLibTypeStruct == "" || c.TypeStruct.gir.Disguised

	ns, parent := e.findTypeByGIRName(c.gir.Parent)

	if parent == nil {
		return false
	}

	if _, ok := parent.(*Class); !ok {
		e.logger.Warn("parent is not a class", "parent", parent.GIRName(), "actual", reflect.TypeOf(parent).String())
		return false
	}

	c.Parent = CouldBeForeign[*Class]{
		Namespace: ns,
		Type:      parent.(*Class),
	}

	for _, impl := range c.gir.Implements {
		ns, inter := e.findTypeByGIRName(impl.Name)

		if inter == nil {
			e.logger.Info("implemented interface not found", "interface", impl.Name)
			continue
		}

		if _, ok := inter.(*Interface); !ok {
			return false
		}

		if c.redundantImplements(inter) {
			e.logger.Debug("skipping redundant interface, as it is already implemented by parents", "interface", impl.Name)
			continue
		}

		c.Implements = append(c.Implements, CouldBeForeign[*Interface]{
			Namespace: ns,
			Type:      inter.(*Interface),
		})
	}

	return true
}

func (c *Class) redundantImplements(inter Type) bool {
	for _, impl := range c.Implements {
		if impl.Type == inter {
			return true
		}
	}

	if c.Parent.Type != nil {
		return c.Parent.Type.redundantImplements(inter)
	}

	return false
}

func (c *Class) declareNested(e *env) {
	e = e.sub("class", c.gir.CType)

	for _, v := range c.gir.Functions {
		if t := DeclarePrefixedFunction(e, c, v.CallableAttrs); t != nil {
			c.Functions = append(c.Functions, t)
		}
	}

	for _, v := range c.gir.Methods {
		if v.Name == "weak_ref" || v.Name == "weak_unref" {
			// we don't want the user to be able to weakly reference the object
			// as there are better tools for this and this will only cause problems
			continue
		}

		if v.Name == "ref" || v.Name == "unref" {
			// reffing will be done on the base GObject, and we don't want such methods generated
			// because they will cause mem leaks
			continue
		}

		if t := DeclareMethod(e, c, v); t != nil {
			c.Methods = append(c.Methods, t)
		}
	}

	for _, v := range c.gir.Constructors {
		if t := DeclarePrefixedFunction(e, c, v.CallableAttrs); t != nil {
			c.Constructors = append(c.Constructors, t)
		}
	}

	for _, v := range c.gir.Signals {
		if t := NewSignal(e, c, v); t != nil {
			c.Signals = append(c.Signals, t)
		}
	}

	for _, v := range c.gir.Fields {
		if t := NewField(e, c, v); t != nil {
			c.Fields = append(c.Fields, t)
		}
	}

	if c.TypeStruct != nil {
		for _, v := range c.gir.VirtualMethods {
			if t := NewVirtualMethod(e, c, c.TypeStruct, v); t != nil {
				c.VirtualMethods = append(c.VirtualMethods, t)
			}
		}
	}
}

func (c *Class) ImplementedGoInterfaceNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, inter := range c.Implements {
			if !yield(inter.WithForeignNamespace(inter.Type.GoInterfaceName)) {
				return
			}
		}
	}
}

func (c *Class) ParentGoInterfaceName() string {
	return c.Parent.WithForeignNamespace(c.Parent.Type.GoInterfaceName)
}

func (c *Class) BaseClassGoUnsafeFromGlibFullFunction() string {
	base := c.BaseClass
	return base.WithForeignNamespace(base.Type.GoUnsafeFromGlibFullFunction())
}

func (c *Class) BaseClassGoUnsafeFromGlibNoneFunction() string {
	base := c.BaseClass
	return base.WithForeignNamespace(base.Type.GoUnsafeFromGlibNoneFunction())
}

func (c *Class) BaseClassGoUnsafeToGlibFullFunction() string {
	base := c.BaseClass
	return base.WithForeignNamespace(base.Type.GoUnsafeToGlibFullFunction())
}

func (c *Class) BaseClassGoUnsafeToGlibNoneFunction() string {
	base := c.BaseClass
	return base.WithForeignNamespace(base.Type.GoUnsafeToGlibNoneFunction())
}

// FindVirtualMethod returns the virtual method with the given C identifier. It is primarily useful for user
// code post processors that need to modify the virtual method names.
func (c *Class) FindVirtualMethod(cidentifier string) *VirtualMethod {
	for _, v := range c.VirtualMethods {
		if v.Invoker.CIndentifier() == cidentifier {
			return v
		}
	}

	return nil
}

// // AllParents returns a list of parents. Note that the list is relative to the current namespace,
// // meaning that foreign types will be labeled as such.
// //
// // This also requires that the following chain can happen:
// //
// //	C1 -> ForeignA[C2] -> C3 -> ForeignB[C4] -> C5
// //
// // From the view of C1 the classes C3 and C5 are also foreign in there respective namespaces. C5 is the base
// // class so it will be used as a pointer. They will get returned as:
// //
// //	[ForeignA[C2], ForeignA[C3], ForeignB[C4], ForeignB[C5]]
// func (c *Class) AllParents() []CouldBeForeign[*Class] {
// 	parents := make([]CouldBeForeign[*Class], 0, 10) // abitrary cap

// 	currentNs := c.Parent.Namespace
// 	currentParent := c.Parent.Type

// 	for {
// 		parents = append(parents, CouldBeForeign[*Class]{
// 			Namespace: currentNs,
// 			Type:      currentParent,
// 		})

// 		if currentParent.Parent.Type == nil {
// 			break
// 		}

// 		nextParent := currentParent.Parent.Type

// 		if currentParent.Parent.Namespace != nil {
// 			// only change the current namespace if the type is not local
// 			// to the current type. If current type was foreign to c and next type
// 			// is local to current, then next will still be foreign to c
// 			currentNs = currentParent.Parent.Namespace
// 		}

// 		currentParent = nextParent
// 	}

// 	return parents
// }
