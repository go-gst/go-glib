package glib

// currently unused code for hash tables:

// import "unsafe"

// // #cgo pkg-config: glib-2.0
// // #cgo CFLAGS: -Wno-deprecated-declarations
// // #include <glib.h>
// import "C"

// // hashTableSize returns the size of the *GHashTable.
// func hashTableSize(ptr unsafe.Pointer) int {
// 	return int(C.g_hash_table_size((*C.GHashTable)(ptr)))
// }

// // hashTableForeach calls f on every key-value pair of the given *GHashTable
// func hashTableForeach(ptr unsafe.Pointer, f func(k, v unsafe.Pointer)) {
// 	var k, v C.gpointer
// 	var iter C.GHashTableIter
// 	C.g_hash_table_iter_init(&iter, (*C.GHashTable)(ptr))

// 	for C.g_hash_table_iter_next(&iter, &k, &v) != 0 {
// 		f(unsafe.Pointer(k), unsafe.Pointer(v))
// 	}
// }

// func UnsafeHashTableFromGlibFull[K comparable, V any](ptr unsafe.Pointer, convertKey func(unsafe.Pointer) K, convertValue func(unsafe.Pointer) V) map[K]V {
// 	hash := make(map[K]V, hashTableSize(ptr))

// 	hashTableForeach(ptr, func(k, v unsafe.Pointer) {
// 		key := convertKey(k)
// 		value := convertValue(v)
// 		hash[key] = value
// 	})

// 	C.g_hash_table_unref((*C.GHashTable)(ptr))

// 	return hash
// }

// func UnsafeHashTableFromGlibNone[K comparable, V any](ptr unsafe.Pointer, convertKey func(unsafe.Pointer) K, convertValue func(unsafe.Pointer) V) map[K]V {
// 	hash := make(map[K]V, hashTableSize(ptr))

// 	hashTableForeach(ptr, func(k, v unsafe.Pointer) {
// 		key := convertKey(k)
// 		value := convertValue(v)
// 		hash[key] = value
// 	})

// 	return hash
// }
