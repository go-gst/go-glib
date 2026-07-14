package instancedata

import "unsafe"

// #cgo CFLAGS: -Wno-deprecated-declarations
import "C"

//export destroyUserdata
func destroyUserdata(ptr unsafe.Pointer) {
	Delete(ptr)
}
