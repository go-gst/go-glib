package glib

// CGO exports have to be defined in a separate file from where they are used or else
// there will be double linkage issues.

/*
#cgo CFLAGS: -Wno-deprecated-declarations
#include "glib.go.h"
*/
import "C"
import "runtime/cgo"

//export goCopyCgoHandle
func goCopyCgoHandle(handle C.guint) C.guint {
	h1 := cgo.Handle(handle)

	h2 := cgo.NewHandle(h1.Value())

	return C.guint(h2)
}

//export goFreeCgoHandle
func goFreeCgoHandle(handle C.guint) {
	h := cgo.Handle(handle)

	h.Delete()
}
