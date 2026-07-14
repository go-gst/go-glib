package file

import (
	"fmt"
	"io"

	"github.com/go-gst/go-glib/gir/girgen/file/internal"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type gTypes []typesystem.Marshalable

// TODO: maybe lookup the RegisterGValueMarshalers function directly in the typesystem
func (ts gTypes) reader() io.Reader {
	if len(ts) == 0 {
		return io.MultiReader()
	}

	var block internal.CodeWriter

	fmt.Fprintln(&block, "// GType values.")
	fmt.Fprintln(&block, "var (")
	block.Indent()

	var decls DeclarationWriter
	for _, t := range ts {
		fmt.Fprintf(&decls, "%s\t= %s(C.%s())\n", t.GoTypeName(), t.Type().NamespacedGoType(0), t.GLibGetType())
	}
	decls.WriteTo(&block)
	block.Unindent()

	fmt.Fprintln(&block, ")")

	// these functions/types are local when generating gobject
	registerFn := ts[0].Type().WithForeignNamespace("RegisterGValueMarshalers")
	marshalerType := ts[0].Type().WithForeignNamespace("TypeMarshaler")

	fmt.Fprintln(&block)
	fmt.Fprintln(&block, "func init() {")
	block.Indent()
	fmt.Fprintf(&block, "%s([]%s{\n", registerFn, marshalerType)
	block.Indent()

	for _, t := range ts {
		fmt.Fprintf(&block, "%s{T: %s, F: marshal%s},\n", marshalerType, t.GoTypeName(), t.GoType(0))
	}

	block.Unindent()

	fmt.Fprintln(&block, "})")
	block.Unindent()
	fmt.Fprintln(&block, "}")

	return &block
}
