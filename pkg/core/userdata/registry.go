// userdata manages data that needs to be passed to C code, such as
// callbacks
package userdata

import (
	"sync"
	"unsafe"
)

// userdataEntry contains the passed data and some metadata
type userdataEntry struct {
	data any
	// once tells the lookup to delete the entry after it is used
	once bool
}

var userdataLock sync.Mutex
var userdataRegistry map[unsafe.Pointer]userdataEntry = make(map[unsafe.Pointer]userdataEntry)

func register(cpointer unsafe.Pointer, data any, once bool) {
	userdataLock.Lock()
	defer userdataLock.Unlock()

	if _, ok := userdataRegistry[cpointer]; ok {
		panic("given pointer is already registered")
	}

	userdataRegistry[cpointer] = userdataEntry{
		data: data,
		once: once,
	}
}

// Register registers the given userdata and returns a valid C pointer
// that can be passed to C code.
func Register(data any) unsafe.Pointer {
	ptr := getPointer()

	register(ptr, data, false)

	return ptr
}

// RegisterOnce registers the given userdata and returns a valid C pointer
// that can be passed to C code.
// The userdata will be deleted after it is used.
func RegisterOnce(data any) unsafe.Pointer {
	ptr := getPointer()

	register(ptr, data, true)

	return ptr
}

func Load(cpointer unsafe.Pointer) any {
	userdataLock.Lock()
	defer userdataLock.Unlock()

	fs, ok := userdataRegistry[cpointer]
	if !ok {
		return nil
	}

	if fs.once {
		deleteUnlocked(cpointer)
	}

	return fs.data
}

// Delete deletes the given userdata
func Delete(cpointer unsafe.Pointer) {
	userdataLock.Lock()
	defer userdataLock.Unlock()

	deleteUnlocked(cpointer)
}

func deleteUnlocked(cpointer unsafe.Pointer) {
	_, ok := userdataRegistry[cpointer]

	if !ok {
		panic("no userdata for given pointer")
	}

	delete(userdataRegistry, cpointer)

	returnPointer(cpointer)
}
