package glib

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib.h>
import "C"
import "unsafe"

// NewBytes creates a new [Bytes] from a copy of the given slice. A nil or empty
// slice yields an empty Bytes, which carries no data pointer.
func NewBytes(data []byte) *Bytes {
	var cdata C.gconstpointer

	// unsafe.SliceData returns an unspecified address for a slice without capacity,
	// so only take a pointer when there is something to copy
	if len(data) > 0 {
		cdata = C.gconstpointer(unsafe.SliceData(data))
	}

	cbytes := C.g_bytes_new(cdata, C.gsize(len(data)))

	return UnsafeBytesFromGlibFull(unsafe.Pointer(cbytes))
}
