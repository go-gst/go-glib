package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

// CToGoBooleanConverter is needed because go's true differs from c's true
type CToGoBooleanConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoBooleanConverter) Convert(w file.File) {
	fmt.Fprintf(w.Go(), "if %s != 0 {\n", c.Param.CName)
	fmt.Fprintf(w.Go(), "\t%s = true\n", c.Param.GoName)
	fmt.Fprintf(w.Go(), "}\n")
}

// Metadata implements Converter.
func (c *CToGoBooleanConverter) Metadata() string {
	return c.Param.Direction
}

var _ Converter = (*CToGoBooleanConverter)(nil)

// GoToCBooleanConverter is needed because go's true differs from c's true
type GoToCBooleanConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCBooleanConverter) Convert(w file.File) {
	param := c.Param.CName

	if c.Param.Direction == "out" {
		// this may be needed for other GoToC out conversions as well
		param = "*" + param
	}

	fmt.Fprintf(w.Go(), "if %s {\n", c.Param.GoName)
	fmt.Fprintf(w.Go(), "\t%s = C.TRUE\n", param)
	fmt.Fprintf(w.Go(), "}\n")
}

// Metadata implements Converter.
func (c *GoToCBooleanConverter) Metadata() string {
	return c.Param.Direction
}

var _ Converter = (*GoToCBooleanConverter)(nil)
