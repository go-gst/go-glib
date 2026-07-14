package gobject

// #include <glib-object.h>
import "C"
import (
	"unsafe"

	"github.com/go-gst/go-glib/pkg/core/userdata"
)

// _goglibInterfaceInit is the function that is called by the GObject system when the interface is initialized on
// the subclass. This is called only once and applies the interface overrides.
//
//export _goglibInterfaceInit
func _goglibInterfaceInit(instance C.gpointer, ifaceData C.gpointer) {
	ptr := unsafe.Pointer(ifaceData)
	// if the interfaceData is deleted then we can never extend the new class.
	// defer userdata.Delete(ptr)

	// Call the downstream interface init handlers
	data := userdata.Load(ptr).(func(gclass unsafe.Pointer))

	data(unsafe.Pointer(instance))
}

// _goglibClassInit is the function that is called by the GObject system when the class is initialized on
// the subclass. This is called only once and applies the class overrides.
//
//export _goglibClassInit
func _goglibClassInit(gclass C.gpointer, classData C.gpointer) {
	data := userdata.Load(unsafe.Pointer(classData)).(*subClassData)

	if data == nil {
		panic("classData is nil")
	}

	data.classInit(unsafe.Pointer(gclass))
}

// _goglibInstanceInit is the function that is called by the GObject system when the instance is initialized on
// the subclass. This is called for each instance and applies the instance overrides.
//
//export _goglibInstanceInit
func _goglibInstanceInit(instance *C.GTypeInstance, gclass C.gpointer) {
	classData := dataFromClass(unsafe.Pointer(gclass))

	if classData == nil {
		panic("classData is nil")
	}

	classData.instanceInit(unsafe.Pointer(instance))
}
