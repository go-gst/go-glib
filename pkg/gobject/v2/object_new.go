package gobject

import (
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
// static GObjectClass *_g_object_get_class(GObject *object) {
//   return (G_OBJECT_GET_CLASS(object));
// }
// static GType _g_type_from_instance(gpointer instance) {
//   return (G_TYPE_FROM_INSTANCE(instance));
// }
import "C"

// NewObjectWithProperties is a wrapper around `g_object_new_with_properties`
//
// Creates a new instance of a GObject subtype and sets its properties using the provided arrays.
//
// Construction parameters (see G_PARAM_CONSTRUCT, G_PARAM_CONSTRUCT_ONLY) which are not explicitly specified are set to their default values.
func NewObjectWithProperties(_type Type, properties map[string]any) Object {
	props := make([]*C.char, 0)
	values := make([]C.GValue, 0)

	for p, v := range properties {
		cpropName := C.CString(p)
		defer C.free(unsafe.Pointer(cpropName))

		props = append(props, cpropName)

		value := NewValue(v)

		// value goes out of scope, but the finalizer must not run until the cgo call is finished
		defer runtime.KeepAlive(value)

		values = append(values, *(*C.GValue)(UnsafeValueToGlibNone(value)))
	}

	propCount := C.uint(len(properties))
	cProps := unsafe.SliceData(props)
	cPropValues := unsafe.SliceData(values)

	cobj := C.g_object_new_with_properties(C.GType(_type), propCount, cProps, cPropValues)

	var obj Object

	if cobj != nil {
		obj = UnsafeObjectFromGlibFull(unsafe.Pointer(cobj))
	}

	return obj
}
