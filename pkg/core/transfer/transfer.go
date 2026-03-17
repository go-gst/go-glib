package transfer

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib.h>
// gchar* newEmptyString() {
//     gchar* empty = g_new(gchar, 1);
//     empty[0] = '\0';
//     return empty;
// }
import "C"
import "unsafe"

// GLibString converts a Go string to a C string that is allocated with g_malloc and should be freed with g_free.
//
// This is needed for strings passed to C that are transfer: full, because g_free will be called to free them.
func GLibString(s string) *C.gchar {
	if s == "" {
		return C.newEmptyString()
	}

	// the given string must not be empty, because the return value for unsafe.StringData is unspecified for empty strings.
	strdata := (*C.gchar)(unsafe.Pointer(unsafe.StringData(s)))

	cstring := C.g_strndup((*C.char)(strdata), C.gsize(len(s)))

	return cstring
}
