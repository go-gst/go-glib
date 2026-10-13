package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type RecordGenerator struct {
	Doc               SubGenerator
	GenerateMarshaler bool

	*typesystem.Record

	// infos used by sub generators:
	ReceiverName string

	// sub generators:
	SubGenerators GeneratorList
}

func (g *RecordGenerator) Generate(w *file.Package) {
	if g.CgoUnrefFunction == "" {
		panic("cannot generate record without an unref method")
	}

	w.GoImport("unsafe")
	w.GoImport("runtime")

	g.Doc.Generate(w.Go())

	if g.IsTypeStructFor != nil {
		fmt.Fprintf(w.Go(), "//\n")
		fmt.Fprintf(w.Go(), "// %s is the type struct for [%s]\n", g.GoType(0), g.IsTypeStructFor.GoType(1))
	}

	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "\t*%s\n", g.PrivateGoType)
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	fmt.Fprintf(w.Go(), "// %s is the struct that's finalized\n", g.PrivateGoType)
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.PrivateGoType)
	if g.Opaque {
		// Pointers to cgo.Incomplete are not GC-scanned. A pointer-sized owner
		// would use the tiny allocator, where a live neighbor can indefinitely
		// prevent its finalizer from running. unsafe.Pointer keeps the same
		// size and native address while making the owner GC-scanned.
		fmt.Fprintf(w.Go(), "\t// Use a GC-scanned pointer so live tiny allocations cannot pin the finalizer.\n")
		fmt.Fprintf(w.Go(), "\tnative unsafe.Pointer\n")
	} else {
		fmt.Fprintf(w.Go(), "\tnative *%s\n", g.CGoType(0))
	}
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	// instance method for returning the c pointer
	fmt.Fprintf(w.Go(), "// %s returns the underlying C pointer. This is used by the bindings internally.\n", g.ToGlibNoneFunction)
	fmt.Fprintf(w.Go(), "func (%s *%s) instance() %s {\n", g.ReceiverName, g.GoType(0), g.CGoType(1))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "if %s == nil {\n", g.ReceiverName)
	fmt.Fprintf(w.Go(), "\treturn nil\n")
	fmt.Fprintf(w.Go(), "}\n")
	if g.Opaque {
		fmt.Fprintf(w.Go(), "return (*%s)(%s.native)\n", g.CGoType(0), g.ReceiverName)
	} else {
		fmt.Fprintf(w.Go(), "return %s.native\n", g.ReceiverName)
	}
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	if g.GenerateMarshaler {
		// GoValueInitializer assertion:
		fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n", g.Value().WithForeignNamespace("GoValueInitializer"), g.GoType(0))
		w.Go().NewSection()

		w.RegisterGType(g.Record)
		fmt.Fprintf(w.Go(), "func marshal%s(p unsafe.Pointer) (interface{}, error) {\n", g.GoType(0))
		fmt.Fprintf(w.Go(), "\tb := %s(p).Boxed()\n", g.Value().WithForeignNamespace(g.Value().Type.FromGlibBorrowFunction))
		fmt.Fprintf(w.Go(), "\treturn %s(b), nil\n", g.GoUnsafeFromGlibNoneFunction())
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()

		fmt.Fprintf(w.Go(), "func (r *%s) GoValueType() %s {\n", g.GoType(0), g.Type().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "return %s\n", g.GoTypeName())
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()

		fmt.Fprintf(w.Go(), "func (r *%s) SetGoValue(v *%s) {\n", g.GoType(0), g.Value().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "v.SetBoxed(unsafe.Pointer(r.instance()))\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go. This is used by the bindings internally.\n", g.GoUnsafeFromGlibBorrowFunction(), g.CGoType(0))
	fmt.Fprintf(w.Go(), "func %s(p unsafe.Pointer) *%s {\n", g.GoUnsafeFromGlibBorrowFunction(), g.GoType(0))
	fmt.Fprintf(w.Go(), "\tif p == nil {\n")
	fmt.Fprintf(w.Go(), "\t\treturn nil\n")
	fmt.Fprintf(w.Go(), "\t}\n")
	if g.Opaque {
		fmt.Fprintf(w.Go(), "\treturn &%s{&%s{p}}\n", g.GoType(0), g.PrivateGoType)
	} else {
		fmt.Fprintf(w.Go(), "\treturn &%s{&%s{(*%s)(p)}}\n", g.GoType(0), g.PrivateGoType, g.CGoType(0))
	}
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	if g.GoUnsafeFromGlibNoneFunction() != "" {
		g.transferNoneFunction(w)
	}

	if g.GoUnsafeFromGlibFullFunction() != "" {
		fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking ownership. This is used by the bindings internally.\n", g.GoUnsafeFromGlibFullFunction(), g.CGoType(0))
		fmt.Fprintf(w.Go(), "func %s(p unsafe.Pointer) *%s {\n", g.GoUnsafeFromGlibFullFunction(), g.GoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "wrapped := %s(p)\n", g.GoUnsafeFromGlibBorrowFunction())
		fmt.Fprintf(w.Go(), "if wrapped == nil {\n")
		fmt.Fprintf(w.Go(), "\treturn nil\n")
		fmt.Fprintf(w.Go(), "}\n")
		g.mkFinalizer(w)
		fmt.Fprintf(w.Go(), "return wrapped\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}

	if g.CgoRefFunction != "" {
		fmt.Fprintf(w.Go(), "// %s increases the refcount on the underlying resource.\n", g.GoUnsafeRefFunction)
		fmt.Fprintf(w.Go(), "//\n")
		fmt.Fprintf(w.Go(), "// When this is called without an associated call to [%s.%s], then [%s] will leak memory.\n", g.GoType(0), g.GoUnsafeUnrefFunction, g.GoType(0))
		fmt.Fprintf(w.Go(), "func %s(%s *%s) {\n", g.GoUnsafeRefFunction, g.ReceiverName, g.GoType(0))
		w.Go().Indent()
		g.refCall(w.Go(), fmt.Sprintf("%s.native", g.ReceiverName))
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}

	fmt.Fprintf(w.Go(), "// %s unrefs/frees the underlying resource. This can be used to remove the instance before the GC decides to do so.\n", g.GoUnsafeUnrefFunction)
	fmt.Fprintf(w.Go(), "//\n")
	fmt.Fprintf(w.Go(), "// After this is called, no other method on [%s] is expected to work anymore.\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "func %s(%s *%s) {\n", g.GoUnsafeUnrefFunction, g.ReceiverName, g.GoType(0))
	w.Go().Indent()
	g.unrefCall(w.Go(), g.ReceiverName)
	fmt.Fprintf(w.Go(), "runtime.SetFinalizer(%s.%s, nil)\n", g.ReceiverName, g.PrivateGoType)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	fmt.Fprintf(w.Go(), "// %s returns the underlying C pointer. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction())
	fmt.Fprintf(w.Go(), "func %s(%s *%s) unsafe.Pointer {\n", g.GoUnsafeToGlibNoneFunction(), g.ReceiverName, g.GoType(0))
	fmt.Fprintf(w.Go(), "\tif %s == nil {\n", g.ReceiverName)
	fmt.Fprintf(w.Go(), "\t\treturn nil\n")
	fmt.Fprintf(w.Go(), "\t}\n")
	fmt.Fprintf(w.Go(), "\treturn unsafe.Pointer(%s.native)\n", g.ReceiverName)
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	if g.GoUnsafeToGlibFullFunction() != "" {
		fmt.Fprintf(w.Go(), "// %s returns the underlying C pointer and gives up ownership.\n", g.GoUnsafeToGlibFullFunction())
		fmt.Fprintf(w.Go(), "// This is used by the bindings internally.\n")
		fmt.Fprintf(w.Go(), "func %s(%s *%s) unsafe.Pointer {\n", g.GoUnsafeToGlibFullFunction(), g.ReceiverName, g.GoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "if %s == nil {\n", g.ReceiverName)
		fmt.Fprintf(w.Go(), "\treturn nil\n")
		fmt.Fprintf(w.Go(), "}\n")
		fmt.Fprintf(w.Go(), "runtime.SetFinalizer(%s.%s, nil)\n", g.ReceiverName, g.PrivateGoType)
		fmt.Fprintf(w.Go(), "_p := unsafe.Pointer(%s.native)\n", g.ReceiverName)
		fmt.Fprintf(w.Go(), "%s.native = nil // %s is invalid from here on\n", g.ReceiverName, g.GoType(0))
		fmt.Fprintf(w.Go(), "return _p\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}

	// for typestructs we need a type cast to the parent type struct
	parentTypeStruct := g.ParentTypeStruct()
	if parentTypeStruct != nil {
		fmt.Fprintf(w.Go(), "// ParentClass returns the type struct of the parent class of this type struct.\n")
		fmt.Fprintf(w.Go(), "// This essentially casts the underlying c pointer.\n")
		fmt.Fprintf(w.Go(), "func (%s *%s) ParentClass() %s {\n", g.ReceiverName, g.GoType(0), parentTypeStruct.NamespacedGoType(1))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "parent := %s(%s(%s))\n", parentTypeStruct.WithForeignNamespace(parentTypeStruct.Type.FromGlibBorrowFunction), g.ToGlibNoneFunction, g.ReceiverName)
		fmt.Fprintf(w.Go(), "// attach a cleanup to keep the instance alive as long as the parent is referenced\n")
		fmt.Fprintf(w.Go(), "runtime.AddCleanup(parent, func(_ %s) {}, %s)\n", g.GoType(1), g.ReceiverName)
		fmt.Fprintf(w.Go(), "return parent\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}

	GenerateAll(
		w,
		g.SubGenerators,
	)
}

// mkFinalizer prints the finalizer code for the record generator. The finalized variable must be called "wrapped".
func (g *RecordGenerator) mkFinalizer(w *file.Package) {
	fmt.Fprintf(w.Go(), "runtime.SetFinalizer(\n")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "wrapped.%s,\n", g.PrivateGoType)
	fmt.Fprintf(w.Go(), "func(intern *%s) {\n", g.PrivateGoType)
	w.Go().Indent()
	g.unrefCall(w.Go(), "intern")
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "},\n")
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), ")\n")
}

func (g *RecordGenerator) transferNoneFunction(w *file.Package) {
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go without transferring ownership. This is used by the bindings internally.\n", g.GoUnsafeFromGlibNoneFunction(), g.CGoType(0))
	fmt.Fprintf(w.Go(), "func %s(p unsafe.Pointer) *%s {\n", g.GoUnsafeFromGlibNoneFunction(), g.GoType(0))
	w.Go().Indent()
	if g.CgoRefFunction != "" {
		// from none only refs if reffing is possible: TODO: this can produce bugs because we are borrowing otherwise
		g.refCall(w.Go(), "p")
	}

	fmt.Fprintf(w.Go(), "wrapped := %s(p)\n", g.GoUnsafeFromGlibBorrowFunction())
	fmt.Fprintf(w.Go(), "if wrapped == nil {\n")
	fmt.Fprintf(w.Go(), "\treturn nil\n")
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	if g.CgoRefFunction == "" && g.GoCopyMethod == nil {
		w.GoImport("log")
		fmt.Fprintf(w.Go(), "log.Println(\"WARNING: not attaching a finalizer to %s because no cgo ref function or copy method is available. This may leak memory. Please file an issue\")\n", g.GoType(0))
		fmt.Fprintf(w.Go(), "return wrapped\n")
	} else if g.CgoRefFunction == "" && g.GoCopyMethod != nil {
		// the copy method already attaches a finalizer
		fmt.Fprintf(w.Go(), "return wrapped.%s() // create an owned copy\n", g.GoCopyMethod.GoIndentifier())
		w.Go().NewSection()
	} else {
		g.mkFinalizer(w)
		fmt.Fprintf(w.Go(), "return wrapped\n")
	}
	w.Go().Unindent()

	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()
}

// refCall writes the ref call, variable is the unsafe pointer of the incoming record
func (g *RecordGenerator) refCall(w file.CodeWriter, variable string) {
	if g.CgoRefNeedsUnsafePointer {
		fmt.Fprintf(w, "%s(unsafe.Pointer(%s))\n", g.CgoRefFunction, variable)
	} else {
		fmt.Fprintf(w, "%s((*%s)(%s))\n", g.CgoRefFunction, g.CGoType(0), variable)
	}
}

func (g *RecordGenerator) unrefCall(w file.CodeWriter, variable string) {
	if g.CgoUnrefNeedsUnsafeCast {
		// unsafe already imported above
		fmt.Fprintf(w, "%s(unsafe.Pointer(%s.native))\n", g.CgoUnrefFunction, variable)
	} else if g.Opaque {
		fmt.Fprintf(w, "%s((*%s)(%s.native))\n", g.CgoUnrefFunction, g.CGoType(0), variable)
	} else {
		fmt.Fprintf(w, "%s(%s.native)\n", g.CgoUnrefFunction, variable)
	}
}

func NewRecordGenerator(cfg *Config, r *typesystem.Record) *RecordGenerator {
	g := &RecordGenerator{
		Doc:               cfg.DocGenerator(r),
		Record:            r,
		GenerateMarshaler: r.GLibGetType() != "",

		ReceiverName: strcases.ReceiverName(r.GoType(0)),
	}

	for _, constructor := range r.Constructors {
		if constGen := NewCallableGenerator(cfg, constructor); constGen != nil {
			g.SubGenerators = append(g.SubGenerators, constGen)
		}
	}

	for _, method := range r.Functions {
		if methGen := NewCallableGenerator(cfg, method); methGen != nil {
			g.SubGenerators = append(g.SubGenerators, methGen)
		}
	}

	for _, method := range r.Methods {
		if methGen := NewCallableGenerator(cfg, method); methGen != nil {
			g.SubGenerators = append(g.SubGenerators, methGen)
		}
	}

	return g
}
