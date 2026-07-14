package gobject

import (
	"unsafe"

	"github.com/go-gst/go-glib/pkg/core/userdata"
)

// #include <glib-object.h>
import "C"

//export _goglib_signalAccumulator
func _goglib_signalAccumulator(
	ihint *C.GSignalInvocationHint,
	return_accu *C.GValue,
	handler_return *C.GValue,
	data C.gpointer,
) C.gboolean {
	goAccuI := userdata.Load(unsafe.Pointer(data))

	goAccu := goAccuI.(SignalAccumulator)

	return gbool(goAccu(
		UnsafeSignalInvocationHintFromGlibBorrow(unsafe.Pointer(ihint)),
		ValueFromNative(unsafe.Pointer(return_accu)),
		ValueFromNative(unsafe.Pointer(handler_return)),
	))
}
