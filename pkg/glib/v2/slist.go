package glib

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib.h>
import "C"
import "unsafe"

// sListSize returns the length of the singly-linked list.
func sListSize(ptr unsafe.Pointer) int {
	return int(C.g_slist_length((*C.GSList)(ptr)))
}

// gslistForeach calls f on every value of the given *GSList
func gslistForeach(ptr unsafe.Pointer, f func(v unsafe.Pointer)) {
	for v := (*C.GSList)(ptr); v != nil; v = v.next {
		f(unsafe.Pointer(v.data))
	}

	C.g_slist_free((*C.GSList)(ptr))
}

// UnsafeSListFromGlibFull converts a *GSList to a slice of T, freeing the list
// in the process.
func UnsafeSListFromGlibFull[T any](ptr unsafe.Pointer, convertChild func(v unsafe.Pointer) T) []T {
	slist := UnsafeSListFromGlibNone(ptr, convertChild)

	C.g_slist_free((*C.GSList)(ptr))

	return slist
}

// UnsafeSListFromGlibNone converts a *GSList to a slice of T, keeping the list
// intact.
func UnsafeSListFromGlibNone[T any](ptr unsafe.Pointer, convertChild func(v unsafe.Pointer) T) []T {
	slist := make([]T, 0, sListSize(ptr))

	gslistForeach(ptr, func(v unsafe.Pointer) {
		slist = append(slist, convertChild(v))
	})

	return slist
}
