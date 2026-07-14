package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
)

// AliasConverter is used to output additional metadata for aliased types. It naivly calls the subconverter for conversion
type AliasConverter struct {
	SubConverter Converter
}

// Convert implements Converter.
func (a *AliasConverter) Convert(w file.File) {
	a.SubConverter.Convert(w)
}

// Metadata implements Converter.
func (a *AliasConverter) Metadata() string {
	return fmt.Sprintf("%s, alias", a.SubConverter.Metadata())
}

var _ Converter = (*AliasConverter)(nil)
