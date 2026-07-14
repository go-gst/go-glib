package glib

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib.h>
import "C"

import (
	"unsafe"
)

const (
	// ErrorCode is an arbitrary code we use for all errors that we map from go to C.
	ErrorCode = 666
)

var ErrorDomain Quark

func init() {
	ErrorDomain = Quark(C.g_quark_from_string(C.CString("go-glib-error-quark")))
}

// New creates a new *C.GError from the given error. The caller is responsible
// for freeing the error with g_error_free().
func UnsafeErrorToGlibFull(err error) unsafe.Pointer {
	if err == nil {
		return nil
	}

	errString := (*C.gchar)(C.CString(err.Error()))
	defer C.free(unsafe.Pointer(errString))

	return unsafe.Pointer(C.g_error_new_literal(C.GQuark(ErrorDomain), C.gint(ErrorCode), errString))
}

// GError is converted from a C.GError to implement Go's error interface.
type GError struct {
	quark uint32
	code  int
	err   string
}

// Quark returns the internal quark for the error. Callers that want this quark
// must manually type assert using their own interface.
func (err *GError) Quark() uint32 {
	return err.quark
}

func (err *GError) ErrorCode() int {
	return err.code
}

func (err *GError) Error() string {
	return err.err
}

// UnsafeErrorFromGlibFull returns a new Go error from a *GError and frees the *GError. this is used by the
// bindings internally when a C function is marked as "throws"
func UnsafeErrorFromGlibFull(gerror unsafe.Pointer) error {
	v := (*C.GError)(gerror)
	defer C.g_error_free(v)

	return newGError(v)
}

func newGError(v *C.GError) *GError {
	return &GError{
		quark: uint32(v.domain),
		code:  int(v.code),
		err:   C.GoString(v.message),
	}
}
