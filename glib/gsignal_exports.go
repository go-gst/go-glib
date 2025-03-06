package glib

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#include "glib.go.h"
*/
import "C"
import (
	"unsafe"

	gopointer "github.com/go-gst/go-pointer"
)

//export goSignalAccumulator
func goSignalAccumulator(
	ihint *C.GSignalInvocationHint,
	return_accu *C.GValue,
	handler_return *C.GValue,
	data C.gpointer,
) C.gboolean {
	goAccuI := gopointer.Restore(unsafe.Pointer(data))

	goAccu := goAccuI.(SignalAccumulator)

	return gbool(goAccu(
		&SignalInvocationHint{ihint},
		ValueFromNative(unsafe.Pointer(return_accu)),
		ValueFromNative(unsafe.Pointer(handler_return)),
	))
}
