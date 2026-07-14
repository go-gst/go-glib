package glib

import "unsafe"

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib.h>
import "C"

// listSize returns the length of the list.
func listSize(ptr unsafe.Pointer) int {
	return int(C.g_list_length((*C.GList)(ptr)))
}

// glistForeach calls f on every value of the given *GList
func glistForeach(ptr unsafe.Pointer, f func(v unsafe.Pointer)) {
	for v := (*C.GList)(ptr); v != nil; v = v.next {
		f(unsafe.Pointer(v.data))
	}
}

func UnsafeListFromGlibFull[T any](ptr unsafe.Pointer, convertChild func(v unsafe.Pointer) T) []T {
	list := UnsafeListFromGlibNone(ptr, convertChild)

	C.g_list_free((*C.GList)(ptr))

	return list
}

func UnsafeListFromGlibNone[T any](ptr unsafe.Pointer, convertChild func(v unsafe.Pointer) T) []T {
	list := make([]T, 0, listSize(ptr))

	glistForeach(ptr, func(v unsafe.Pointer) {
		list = append(list, convertChild(v))
	})

	return list
}
