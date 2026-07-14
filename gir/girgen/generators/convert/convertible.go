package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CToGoConvertibleConverter struct {
	Param       *typesystem.Param
	ConvertFunc string
}

// Convert implements Converter.
func (c *CToGoConvertibleConverter) Convert(w file.File) {
	w.GoImport("unsafe")

	fmt.Fprintf(w.Go(), "%s = %s(unsafe.Pointer(%s))\n", c.Param.GoName, c.ConvertFunc, c.Param.CName)

	if c.Param.TransferOwnership == typesystem.TransferBorrow {
		if c.Param.BorrowFrom != nil {
			w.GoImport("runtime")

			fmt.Fprintf(w.Go(), "runtime.AddCleanup(%s, func(_ *%s) {}, %s)\n", c.Param.GoName, c.Param.BorrowFrom.Type.NamespacedGoType(0), c.Param.BorrowFrom.GoName)
		} else {
			fmt.Fprintf(w.Go(), "// borrow not bound to another value, this requires correct handling by the user\n")
		}
	}
}

// Metadata implements Converter.
func (c *CToGoConvertibleConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, converted", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*CToGoConvertibleConverter)(nil)

type GoToCConvertibleConverter struct {
	Param       *typesystem.Param
	ConvertFunc string
}

// Convert implements Converter.
func (c *GoToCConvertibleConverter) Convert(w file.File) {
	cname := c.Param.CName

	if c.Param.Direction == "out" {
		cname = "*" + cname
	}

	fmt.Fprintf(w.Go(), "%s = (%s)(%s(%s))\n", cname, c.Param.CGoType(), c.ConvertFunc, c.Param.GoName)
}

// Metadata implements Converter.
func (c *GoToCConvertibleConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, converted", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*GoToCConvertibleConverter)(nil)
