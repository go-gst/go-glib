package convert

import (
	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

// ImplicitConverter does a no-op conversion, because it will be set and read by another
// converter
type ImplicitConverter struct {
	*typesystem.Param
}

// Metadata implements Converter.
func (i *ImplicitConverter) Metadata() string {
	return "implicit"
}

// PostCallConvert implements Converter.
func (i *ImplicitConverter) Convert(file.File) {}

var _ Converter = (*ImplicitConverter)(nil)

// SkippedConverter does a no-op conversion, the value is only useful in C, so we pass a zero type
type SkippedConverter struct {
	*typesystem.Param
}

// Metadata implements Converter.
func (i *SkippedConverter) Metadata() string {
	return "skipped"
}

// PostCallConvert implements Converter.
func (i *SkippedConverter) Convert(file.File) {}

var _ Converter = (*SkippedConverter)(nil)
