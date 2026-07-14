package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type GoToCFixedSizeArrayConvertibleConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCFixedSizeArrayConvertibleConverter) Convert(w file.File) {
	cname := c.Param.CName

	array := c.Param.Type.Type.(*typesystem.Array)

	if array.FixedSize == 0 || array.Inner.Type == nil || array.InnerPointers != 0 {
		panic("GoToCFixedSizeArrayConvertibleConverter can only be used for fixed-size arrays with a known inner type and no inner pointers")
	}

	w.GoImport("unsafe")

	fmt.Fprintln(w.Go(), "{")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "var carr [%d]%s\n", array.FixedSize, array.Inner.Type.CGoType(0))
	fmt.Fprintf(w.Go(), "for i := range %d {\n", array.FixedSize)
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "carr[i] = %s(%s[i])\n", array.Inner.Type.CGoType(0), c.Param.GoName)
	fmt.Fprintf(w.Go(), "%s = unsafe.SliceData(carr[:])\n", cname)
	w.Go().Unindent()
	fmt.Fprintln(w.Go(), "}")
	w.Go().Unindent()
	fmt.Fprintln(w.Go(), "}")
}

// Metadata implements Converter.
func (c *GoToCFixedSizeArrayConvertibleConverter) Metadata() string {
	array := c.Param.Type.Type.(*typesystem.Array)
	return fmt.Sprintf("%s, %s, array fixed size (inner: %s, size: %d)", c.Param.Direction, c.Param.TransferOwnership, array.Inner.Type.CType(0), array.FixedSize)
}

var _ Converter = (*GoToCFixedSizeArrayConvertibleConverter)(nil)
