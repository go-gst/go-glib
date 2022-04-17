package glib

/*
#include "glib.go.h"

GObjectClass *  toGObjectClass  (void *p)  { return (G_OBJECT_CLASS(p)); }
*/
import "C"

import (
	"math"
	"unsafe"
)

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

// ListProperties returns a list of the properties associated with this object.
// The default values assumed in the parameter spec reflect the values currently
// set in this object, or their defaults.
//
// Unref params after usage.
func (o *ObjectClass) ListProperties() []*ParamSpec {
	var size C.guint
	props := C.g_object_class_list_properties((*C.GObjectClass)(o.Instance()), &size)
	if props == nil {
		return nil
	}
	defer C.g_free((C.gpointer)(props))
	out := make([]*ParamSpec, 0)

	for _, prop := range (*[(math.MaxInt32 - 1) / unsafe.Sizeof((*C.GParamSpec)(nil))]*C.GParamSpec)(unsafe.Pointer(props))[:size:size] {
		out = append(out, ToParamSpec(unsafe.Pointer(prop)))
	}
	return out
}

func wrapObjectClass(klass C.gpointer) *ObjectClass {
	return &ObjectClass{ptr: C.toGObjectClass(unsafe.Pointer(klass))}
}
