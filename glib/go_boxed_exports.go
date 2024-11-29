package glib

// CGO exports have to be defined in a separate file from where they are used or else
// there will be double linkage issues.

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#include "glib.go.h"
*/
import "C"
import (
	"unsafe"

	gopointer "github.com/go-gst/go-pointer"
)

//export goCopyGoPointer
func goCopyGoPointer(handle C.gpointer) C.gpointer {
	v1 := gopointer.Restore(unsafe.Pointer(handle))

	h2 := gopointer.Save(v1)

	return C.gpointer(h2)
}

//export goFreeGoPointer
func goFreeGoPointer(handle C.gpointer) {
	gopointer.Unref(unsafe.Pointer(handle))
}
