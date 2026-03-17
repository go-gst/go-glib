package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CToGoStringConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoStringConverter) Convert(w file.File) {
	w.GoImport("unsafe")

	// C.GoString always requires the *C.char type, so we cast it always
	fmt.Fprintf(w.Go(), "%s = C.GoString((*C.char)(unsafe.Pointer(%s)))\n", c.Param.GoName, c.Param.CName)

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		// GoString copies the param, so free it immediately
		fmt.Fprintf(w.Go(), "defer C.g_free(C.gpointer(%s))\n", c.Param.CName)
	case typesystem.TransferNone:
		// C will free it
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}
}

// Metadata implements Converter.
func (c *CToGoStringConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, string", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*CToGoStringConverter)(nil)

type GoToCStringConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCStringConverter) Convert(w file.File) {
	w.GoImport("unsafe")

	param := c.Param.CName

	if c.Param.Direction == "out" {
		// this may be needed for other GoToC out conversions as well
		param = "*" + param
	}

	fmt.Fprintf(w.Go(), "%s = (%s)(unsafe.Pointer(C.CString(%s)))\n", param, c.Param.CGoType(), c.Param.GoName)

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		panic("TransferFull is not supported for GoToCStringConverter, because the free function may be different than C.free")
	case typesystem.TransferNone:
		fmt.Fprintf(w.Go(), "defer C.free(unsafe.Pointer(%s))\n", c.Param.CName)
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}
}

// Metadata implements Converter.
func (c *GoToCStringConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, string", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*GoToCStringConverter)(nil)
