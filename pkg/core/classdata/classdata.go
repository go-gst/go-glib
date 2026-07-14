// package classdata stores and retrieves virtual functions for overloaded virtual methods.
package classdata

import "unsafe"

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
// static GObjectClass *_g_object_get_class(GObject *object) {
//   return (G_OBJECT_GET_CLASS(object));
// }
import "C"

type virtualMethodIdentifier struct {
	// class is the class that the virtual method belongs to
	class unsafe.Pointer
	// name is a unique identifier for the virtual method on the class. Remember that all classes in the type
	// hierarchie share the same class pointer, so this must be unique for the class every parent class also.
	name string
}

var registry = make(map[virtualMethodIdentifier]any)

// StoreVirtualMethod stores the virtual method for the given class.
func StoreVirtualMethod(
	class unsafe.Pointer,
	name string,
	f any,
) {
	registry[virtualMethodIdentifier{
		class: class,
		name:  name,
	}] = f
}

// LoadVirtualMethodFromInstance loads the virtual method for a given instance of a class.
// This is used by the generated bindings to call the virtual method.
//
// instance is the CGo pointer to the GObject type instance.
// trampoline is the CGo pointer to the trampoline function.
func LoadVirtualMethodFromInstance(
	instance unsafe.Pointer,
	name string,
) any {
	class := C._g_object_get_class((*C.GObject)(instance))

	return registry[virtualMethodIdentifier{
		class: unsafe.Pointer(class),
		name:  name,
	}]
}
