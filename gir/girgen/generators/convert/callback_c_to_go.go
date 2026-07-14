package convert

import (
	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CToGoCallbackConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoCallbackConverter) Convert(w file.File) {
	if !c.Param.IsUserData {
		panic("not userdata param")
	}
}

// Metadata implements Converter.
func (c *CToGoCallbackConverter) Metadata() string {
	return "userdata containing callback pointer"
}

var _ Converter = (*CToGoCallbackConverter)(nil)
