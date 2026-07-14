package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CToGoCastingConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoCastingConverter) Convert(w file.File) {

	fmt.Fprintf(w.Go(), "%s = %s(%s)\n", c.Param.GoName, c.Param.GoType(), c.Param.CName)
}

// Metadata implements Converter.
func (c *CToGoCastingConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, casted", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*CToGoCastingConverter)(nil)

type GoToCCastingConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCCastingConverter) Convert(w file.File) {
	cname := c.Param.CName

	if c.Param.Direction == "out" {
		cname = "*" + cname
	}

	fmt.Fprintf(w.Go(), "%s = %s(%s)\n", cname, c.Param.CGoType(), c.Param.GoName)
}

// Metadata implements Converter.
func (c *GoToCCastingConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, casted", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*GoToCCastingConverter)(nil)
