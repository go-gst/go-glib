package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type GoToCCallbackConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCCallbackConverter) Convert(w file.File) {
	if c.Param.Closure == nil {
		panic("no closure param")
	}

	if c.Param.Scope == typesystem.CallbackParamScopeNotified && c.Param.Destroy == nil {
		panic("destroy closure without destroy param")
	}

	cb := c.Param.Type.Type.(*typesystem.Callback)
	closure := c.Param.Closure
	destroy := c.Param.Destroy

	w.GoImportCore("userdata")

	assignFunc := "Register"

	if c.Param.Scope == typesystem.CallbackParamScopeAsync {
		assignFunc = "RegisterOnce"
	}

	fmt.Fprintf(w.Go(), "%s = (*[0]byte)(C.%s)\n", c.Param.CName, cb.TrampolineName)
	fmt.Fprintf(w.Go(), "%s = %s(userdata.%s(%s))\n", closure.CName, closure.CGoType(), assignFunc, c.Param.GoName)

	switch c.Param.Scope {
	case typesystem.CallbackParamScopeAsync, typesystem.CallbackParamScopeForever:
		// nothing
	case typesystem.CallbackParamScopeCall:
		w.GoImport("unsafe")
		fmt.Fprintf(w.Go(), "defer userdata.Delete(unsafe.Pointer(%s))\n", closure.CName)
	case typesystem.CallbackParamScopeNotified:
		destroyTrampoline := destroy.Type.Type.(*typesystem.Callback).TrampolineName
		fmt.Fprintf(w.Go(), "%s = (%s)((*[0]byte)(C.%s))\n", destroy.CName, destroy.CGoType(), destroyTrampoline)
	default:
		panic(fmt.Sprintf("unexpected typesystem.CallbackParamScope: %#v", c.Param.Scope))
	}
}

// Metadata implements Converter.
func (c *GoToCCallbackConverter) Metadata() string {
	if c.Param.Scope == typesystem.CallbackParamScopeNotified {
		return fmt.Sprintf("callback, scope: %s, closure: %s, destroy: %s", c.Param.Scope, c.Param.Closure.CName, c.Param.Destroy.CName)
	}
	return fmt.Sprintf("callback, scope: %s, closure: %s", c.Param.Scope, c.Param.Closure.CName)
}

var _ Converter = (*GoToCCallbackConverter)(nil)
