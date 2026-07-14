package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CToGoNullableStringConverter struct {
	Param        *typesystem.Param
	SubConverter *CToGoStringConverter
}

// Convert implements Converter.
func (c *CToGoNullableStringConverter) Convert(w file.File) {
	fmt.Fprintf(w.Go(), "if %s != nil {\n", c.Param.CName)
	w.Go().Indent()
	c.SubConverter.Convert(w)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
}

// Metadata implements Converter.
func (c *CToGoNullableStringConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable-string", c.SubConverter.Metadata())
}

var _ Converter = (*CToGoNullableStringConverter)(nil)

type GoToCNullableStringConverter struct {
	Param        *typesystem.Param
	SubConverter *GoToCStringConverter
}

// Convert implements Converter.
func (c *GoToCNullableStringConverter) Convert(w file.File) {
	fmt.Fprintf(w.Go(), "if %s != \"\" {\n", c.Param.GoName)
	w.Go().Indent()
	c.SubConverter.Convert(w)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
}

// Metadata implements Converter.
func (c *GoToCNullableStringConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable-string", c.SubConverter.Metadata())
}

var _ Converter = (*GoToCNullableStringConverter)(nil)
