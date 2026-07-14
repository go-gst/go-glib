package userdata

/*
#include <stdlib.h>
*/
import "C"
import (
	"sync"
	"unsafe"
)

const chunkSize = 1024 // Number of pointers to allocate in one chunk

var (
	mu sync.Mutex
	// unused is a list of unused but valid malloc'ed C pointers. These can be used
	// to store userdata in a map and then be passed to c code without the runtime ever complaining.
	unused []unsafe.Pointer
)

func allocateChunk() {
	// Allocate a single block of memory for `chunkSize` pointers
	block := C.malloc(C.size_t(chunkSize))

	// Divide the block into individual 1-byte pointers
	for i := range chunkSize {
		ptr := unsafe.Pointer(uintptr(block) + uintptr(i))
		unused = append(unused, ptr)
	}
}

// getPointer retrieves an unused C pointer from the free list, allocating a new chunk if necessary
func getPointer() unsafe.Pointer {
	mu.Lock()
	defer mu.Unlock()

	if len(unused) == 0 {
		allocateChunk()
	}

	// Pop a pointer from the free list
	ptr := unused[len(unused)-1]
	unused = unused[:len(unused)-1]
	return ptr
}

// returnPointer adds a pointer back to the free list for reuse
func returnPointer(ptr unsafe.Pointer) {
	mu.Lock()
	defer mu.Unlock()

	unused = append(unused, ptr)
}
