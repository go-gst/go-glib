package classdata

import "unsafe"

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
// static GObjectClass *_g_object_get_class(GObject *object) {
//   return (G_OBJECT_GET_CLASS(object));
// }
import "C"

// PeekParentClass returns a c pointer to the parent class of the given GObject instance.
func PeekParentClass(instance unsafe.Pointer) unsafe.Pointer {
	class := C._g_object_get_class((*C.GObject)(instance))
	return unsafe.Pointer(C.g_type_class_peek_parent(C.gpointer(class)))
}

// PeekParentInterface returns a c pointer to the parent interface of the given GType for the given GObject instance.
// This is used by the generated bindings to get the interface of the given GType.
//
// The gtype is a uint64 here so that we do not depend on the gobject package.
func PeekParentInterface(instance unsafe.Pointer, gtype uint64) unsafe.Pointer {
	class := C._g_object_get_class((*C.GObject)(instance))
	return unsafe.Pointer(C.g_type_interface_peek(C.gpointer(class), C.GType(gtype)))
}
