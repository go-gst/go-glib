package glib

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#include "glib.go.h"
*/
import "C"

import (
	"reflect"
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

//export goObjectSetProperty
func goObjectSetProperty(obj *C.GObject, propID C.guint, val *C.GValue, param *C.GParamSpec) {
	WithPointerTransferOriginal(unsafe.Pointer(obj), func(object *Object, subclass GoObjectSubclass) {
		iface := subclass.(interface{ SetProperty(*Object, uint, *Value) })
		iface.SetProperty(object, uint(propID-1), ValueFromNative(unsafe.Pointer(val)))
	})
}

//export goObjectGetProperty
func goObjectGetProperty(obj *C.GObject, propID C.guint, value *C.GValue, param *C.GParamSpec) {
	WithPointerTransferOriginal(unsafe.Pointer(obj), func(object *Object, subclass GoObjectSubclass) {
		iface := subclass.(interface{ GetProperty(*Object, uint) *Value })
		val := iface.GetProperty(object, uint(propID-1))
		if val == nil {
			return
		}
		C.g_value_copy((*C.GValue)(unsafe.Pointer(val.GValue)), value)
	})
}

//export goObjectConstructed
func goObjectConstructed(obj *C.GObject) {
	WithPointerTransferOriginal(unsafe.Pointer(obj), func(object *Object, subclass GoObjectSubclass) {
		iface := subclass.(interface{ Constructed(*Object) })
		iface.Constructed(object)
	})
}

//export goObjectFinalize
func goObjectFinalize(obj *C.GObject, klass C.gpointer) {
	// registerMutex.Lock()
	// defer registerMutex.Unlock()
	// delete(registeredClasses, klass)
	gopointer.Unref(privateFromObj(unsafe.Pointer(obj)))
}

//export goClassInit
func goClassInit(klass C.gpointer, klassData C.gpointer) {
	registerMutex.Lock()
	defer registerMutex.Unlock()

	// deref the classdata and register this C pointer to the instantiated go type
	ptr := unsafe.Pointer(klassData)
	data := gopointer.Restore(ptr).(*classData)
	defer gopointer.Unref(ptr)

	registeredClasses[klass] = data.elem

	// add private data where we will store the actual pointer to the go object later
	C.g_type_class_add_private(klass, C.gsize(unsafe.Sizeof(uintptr(0))))

	// Run the downstream class extensions
	data.ext.InitClass(unsafe.Pointer(klass), data.elem)
	data.elem.ClassInit(wrapObjectClass(klass))
}

//export goInterfaceInit
func goInterfaceInit(iface C.gpointer, ifaceData C.gpointer) {
	ptr := unsafe.Pointer(ifaceData)
	defer gopointer.Unref(ptr)
	// Call the downstream interface init handlers
	data := gopointer.Restore(ptr).(*interfaceData)
	data.iface.Init(&TypeInstance{
		GoType:        data.classData.elem,
		GType:         data.gtype,
		GTypeInstance: unsafe.Pointer(iface),
	})
}

//export goInstanceInit
func goInstanceInit(obj *C.GTypeInstance, klass C.gpointer) {
	registerMutex.Lock()
	defer registerMutex.Unlock()

	// Save the go object that was registered to this pointer to the private data of the GObject
	goelem := registeredClasses[klass].New()
	typeName := reflect.TypeOf(registeredClasses[klass]).String()
	ptr := gopointer.Save(goelem)
	private := C.g_type_instance_get_private(obj, C.GType(registeredTypes[typeName]))

	C.memcpy(unsafe.Pointer(private), unsafe.Pointer(&ptr), C.gsize(unsafe.Sizeof(uintptr(0))))

	// If the go type implements an Initter call it.
	if initter, ok := goelem.(Initter); ok {
		initter.InstanceInit(newObject(ToGObject(unsafe.Pointer(obj))))
	}
}
