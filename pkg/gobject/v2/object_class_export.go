package gobject

import (
	"unsafe"

	"github.com/go-gst/go-glib/pkg/core/classdata"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
import "C"

//export _goglib_gobject2_Object_constructed
func _goglib_gobject2_Object_constructed(carg0 *C.GObject) {
	var fn func(carg0 *C.GObject)
	{
		fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(carg0), "_goglib_gobject2_Object_constructed").(func(carg0 *C.GObject))
		if fn == nil {
			panic("_goglib_gobject2_Object_constructed: no function pointer found")
		}
	}
	fn(carg0)
}

//export _goglib_gobject2_Object_dispose
func _goglib_gobject2_Object_dispose(carg0 *C.GObject) {
	var fn func(carg0 *C.GObject)
	{
		fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(carg0), "_goglib_gobject2_Object_dispose").(func(carg0 *C.GObject))
		if fn == nil {
			panic("_goglib_gobject2_Object_dispose: no function pointer found")
		}
	}
	fn(carg0)
}

//export _goglib_gobject2_Object_get_property
func _goglib_gobject2_Object_get_property(carg0 *C.GObject, id C.guint, value *C.GValue, pspec *C.GParamSpec) {
	var fn func(carg0 *C.GObject, id C.guint, value *C.GValue, pspec *C.GParamSpec)
	{
		fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(carg0), "_goglib_gobject2_Object_get_property").(func(carg0 *C.GObject, id C.guint, value *C.GValue, pspec *C.GParamSpec))
		if fn == nil {
			panic("_goglib_gobject2_Object_get_property: no function pointer found")
		}
	}
	fn(carg0, id, value, pspec)
}

//export _goglib_gobject2_Object_set_property
func _goglib_gobject2_Object_set_property(carg0 *C.GObject, id C.guint, value *C.GValue, pspec *C.GParamSpec) {
	var fn func(carg0 *C.GObject, id C.guint, value *C.GValue, pspec *C.GParamSpec)
	{
		fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(carg0), "_goglib_gobject2_Object_set_property").(func(carg0 *C.GObject, id C.guint, value *C.GValue, pspec *C.GParamSpec))
		if fn == nil {
			panic("_goglib_gobject2_Object_set_property: no function pointer found")
		}
	}
	fn(carg0, id, value, pspec)
}

//export _goglib_gobject2_Object_finalize
func _goglib_gobject2_Object_finalize(carg0 *C.GObject) {
	var fn func(carg0 *C.GObject)
	{
		fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(carg0), "_goglib_gobject2_Object_finalize").(func(carg0 *C.GObject))
		if fn == nil {
			panic("_goglib_gobject2_Object_finalize: no function pointer found")
		}
	}
	fn(carg0)
}
