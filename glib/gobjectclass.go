package glib

/*
#include "glib.go.h"

GObjectClass *  toGObjectClass  (void *p)  { return (G_OBJECT_CLASS(p)); }
*/
import "C"

import "unsafe"

// ObjectClass is a binding around the glib GObjectClass. It exposes methods
// to be used during the construction of objects backed by the go runtime.
type ObjectClass struct {
	ptr *C.GObjectClass
}

// Unsafe is a convenience wrapper to return the unsafe.Pointer of the underlying C instance.
func (o *ObjectClass) Unsafe() unsafe.Pointer { return unsafe.Pointer(o.ptr) }

// Instance returns the underlying C GObjectClass pointer
func (o *ObjectClass) Instance() *C.GObjectClass { return o.ptr }

// InstallProperties will install the given ParameterSpecs to the object class.
// They will be IDed in the order they are provided.
func (o *ObjectClass) InstallProperties(params []*ParamSpec) {
	for idx, prop := range params {
		C.g_object_class_install_property(
			o.Instance(),
			C.guint(idx+1),
			prop.paramSpec,
		)
	}
}

func wrapObjectClass(klass C.gpointer) *ObjectClass {
	return &ObjectClass{ptr: C.toGObjectClass(unsafe.Pointer(klass))}
}
