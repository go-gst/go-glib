package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CToGoNullableConverter struct {
	Param        *typesystem.Param
	SubConverter Converter
}

// Convert implements Converter.
func (c *CToGoNullableConverter) Convert(w file.File) {
	fmt.Fprintf(w.Go(), "if %s != nil {\n", c.Param.CName)
	w.Go().Indent()
	c.SubConverter.Convert(w)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
}

// Metadata implements Converter.
func (c *CToGoNullableConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable", c.SubConverter.Metadata())
}

var _ Converter = (*CToGoNullableConverter)(nil)

// GoToCNullableConverter is needed because go's true differs from c's true
type GoToCNullableConverter struct {
	Param        *typesystem.Param
	SubConverter Converter
}

// Convert implements Converter.
func (c *GoToCNullableConverter) Convert(w file.File) {
	fmt.Fprintf(w.Go(), "if %s != nil {\n", c.Param.GoName)
	w.Go().Indent()
	c.SubConverter.Convert(w)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
}

// Metadata implements Converter.
func (c *GoToCNullableConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable", c.SubConverter.Metadata())
}

var _ Converter = (*GoToCNullableConverter)(nil)
