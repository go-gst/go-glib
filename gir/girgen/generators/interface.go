package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type InterfaceGenerator struct {
	Doc SubGenerator

	*typesystem.Interface

	Marshaler Generator

	// infos used by sub generators:
	ReceiverName string

	// sub generators:
	SubGenerators GeneratorList

	// Methods contains all generated methods for the go interface of the interface
	// This includes generated signal connect and emit methods
	Methods MethodGeneratorList

	Overrides *GoOverridesGenerator
}

// InterfaceInstanceStructFieldName is the name of the field where we store the gobject.Object.
// It cannot be embedded, because we are embedding the interface struct in the
// interface instance struct, which creates ambiguous selector compile errors.
const InterfaceInstanceStructFieldName = "Instance"

func (g *InterfaceGenerator) Generate(w *file.Package) {
	w.GoImport("unsafe")

	w.GoImportNamespace(g.Parent.Namespace)

	fmt.Fprintf(w.Go(), "// %s is the instance type used by all types implementing %s. It is used internally by the bindings. Users should use the interface [%s] instead.\n", g.GoType(0), g.CType(0), g.GoInterfaceName)
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "\t_ [0]func() // equal guard\n")

	fmt.Fprintf(w.Go(), "\t%s %s\n", InterfaceInstanceStructFieldName, g.Parent.NamespacedGoType(0))

	// for _, inter := range g.Prerequesite {
	// 	fmt.Fprintf(w.Go(), "\t*%s\n", inter.GoType(0))
	// }
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.GoInterfaceName, g.GoType(0))

	g.Doc.Generate(w.Go())
	fmt.Fprintf(w.Go(), "type %s interface {\n", g.GoInterfaceName)
	w.Go().Indent()
	if g.ManuallyExtended {
		fmt.Fprintf(w.Go(), "%sExtManual // handwritten functions\n", g.GoInterfaceName)
	}

	// fmt.Fprintf(w.Go(), "%s\n", g.Parent.NamespacedGoType(1)) Cannot inherit from base type because we do not embed it

	fmt.Fprintf(w.Go(), "%s() *%s\n", g.GoPrivateUpcastMethod, g.GoType(0))

	// fmt.Fprintln(w.Go(), g.Parent.WithForeignNamespace(g.Parent.Type.GoInterfaceName))
	// for inter := range g.PrerequesitesGoInterfaceNames() {
	// 	fmt.Fprintln(w.Go(), inter)
	// }

	w.Go().NewSection()

	g.Methods.GenerateInterfaceSignatures(w)

	w.Go().NewSection()
	if g.Overrides != nil {
		fmt.Fprintln(w.Go(), "// chain up virtual methods:")
		w.Go().NewSection()

		g.Overrides.VirtualMethods.GenerateInterfaceSignatures(w)
	}

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.GoInterfaceName, g.GoType(0))

	baseClassIdentifier := "base"

	fmt.Fprintf(w.Go(), "func %s(%s *%s) *%s {\n", g.GoWrapBaseClassFunction, baseClassIdentifier, g.Parent.WithForeignNamespace(g.Parent.Type.GoType(0)), g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "return &%s{\n", g.GoType(0))
	w.Go().Indent()

	fmt.Fprintf(w.Go(), "%s: *%s,\n", InterfaceInstanceStructFieldName, baseClassIdentifier)
	w.Go().Unindent()

	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()

	fmt.Fprintf(w.Go(), "}\n\n")

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
		fmt.Fprintln(w.Go())
	}

	mkConstructor := func(constructorName, parentConstructorName string) {
		fmt.Fprintf(w.Go(), "func %s(c unsafe.Pointer) %s {\n", constructorName, g.GoInterfaceName)
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "return %s(c).(%s)\n", parentConstructorName, g.GoInterfaceName)
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "func (%s *%s) %s() *%s {\n", strcases.ReceiverName(g.GoType(0)), g.GoType(0), g.GoPrivateUpcastMethod, g.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn %s\n", strcases.ReceiverName(g.GoType(0)))
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking a reference and attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibNoneFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibNoneFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeFromGlibNoneFunction()))
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibFullFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibFullFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeFromGlibFullFunction()))
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go without touching any references. This is used by the bindings internally.\n", g.GoUnsafeFromGlibBorrowFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibBorrowFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeFromGlibBorrowFunction()))

	mkTransfer := func(transfername, baseTransferName string) {
		fmt.Fprintf(w.Go(), "func %s(c %s) unsafe.Pointer {\n", transfername, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "\ti := c.%s()\n", g.GoPrivateUpcastMethod)
		fmt.Fprintf(w.Go(), "\treturn %s(&i.%s)\n", baseTransferName, InterfaceInstanceStructFieldName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction(), g.CType(0))
	mkTransfer(g.GoUnsafeToGlibNoneFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeToGlibNoneFunction()))

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s, while removeing the finalizer. This is used by the bindings internally.\n", g.GoUnsafeToGlibFullFunction(), g.CType(0))
	mkTransfer(g.GoUnsafeToGlibFullFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeToGlibFullFunction()))

	GenerateAll(
		w,
		g.SubGenerators,
		g.Methods,
		// g.Overrides,
	)
}

func NewInterfaceGenerator(cfg *Config, c *typesystem.Interface) *InterfaceGenerator {
	var marshaler Generator

	if c.GLibGetType() != "" {
		marshaler = NewMarshalObjectGenerator(c)
	}

	g := &InterfaceGenerator{
		Doc:       cfg.DocGenerator(c),
		Interface: c,
		Marshaler: marshaler,

		// Overrides: NewGoOverridesGenerator(c),
	}

	for _, fn := range c.Functions {
		if fGen := NewCallableGenerator(cfg, fn); fGen != nil {
			g.SubGenerators = append(g.SubGenerators, fGen)
		}
	}

	for _, method := range c.Methods {
		if methGen := NewCallableGenerator(cfg, method); methGen != nil {
			g.Methods = append(g.Methods, methGen)
		}
	}

	for _, sig := range c.Signals {
		if sig.Action {
			// action signals are only emitted:
			g.Methods = append(g.Methods, NewSignalEmitGenerator(cfg, sig))
		} else {
			g.Methods = append(g.Methods, NewSignalConnectGenerator(cfg, sig))
		}
	}

	return g
}

func wrapInterface(w file.CodeWriter, t typesystem.CouldBeForeign[*typesystem.Interface], baseClassIdentifier string) {
	fmt.Fprintf(w, "%s: %s{\n", t.Type.GoType(0), t.NamespacedGoType(0))

	w.Indent()

	fmt.Fprintf(w, "%s: *%s,\n", InterfaceInstanceStructFieldName, baseClassIdentifier)

	w.Unindent()

	fmt.Fprintf(w, "},\n")
}
