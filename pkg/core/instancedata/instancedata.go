package instancedata

import (
	"sync"
	"unsafe"
)

// instanceDataKey is the key that is used to store the data for the instance.
// it is safe to pass to C code, if the passed pointers are safe to pass to C code.
//
// it is only exposed to the user as an unsafe.Pointer
type instanceDataKey struct {
	instance unsafe.Pointer
	// key is the key that is used to store the data for the instance. It should uniquely identify the data.
	// A good key to use is the Cgo address of the trampoline function.
	key unsafe.Pointer
}

// instanceDataEntry contains the passed data and some metadata
type instanceDataEntry struct {
	data any
	// once tells the lookup to delete the entry after it is used
	once bool
}

var instanceDataLock sync.Mutex
var instanceDataRegistry map[instanceDataKey]instanceDataEntry = make(map[instanceDataKey]instanceDataEntry)

// register registers the given data for the given instance and key.
// It is safe to call this function from multiple goroutines.
// It is the caller's responsibility to ensure that the instance and key are unique.
// The data will be deleted after it is used if once is true.
//
// the key needed for lookup the data again is returned, but as an unsafe.Pointer.
func register(instance unsafe.Pointer, key unsafe.Pointer, data any, once bool) unsafe.Pointer {
	instanceDataLock.Lock()
	defer instanceDataLock.Unlock()

	k := instanceDataKey{instance: instance, key: key}

	if _, ok := instanceDataRegistry[k]; ok {
		panic("given pointer is already registered")
	}

	instanceDataRegistry[k] = instanceDataEntry{
		data: data,
		once: once,
	}

	return unsafe.Pointer(&k)
}

// Register registers the given data for the given instance and key.
// It is safe to call this function from multiple goroutines.
func Register(instance unsafe.Pointer, key unsafe.Pointer, data any) unsafe.Pointer {
	return register(instance, key, data, false)
}

// RegisterOnce registers the given data for the given instance and key.
// It is safe to call this function from multiple goroutines.
// The data will be deleted after it is used.
func RegisterOnce(instance unsafe.Pointer, key unsafe.Pointer, data any) unsafe.Pointer {
	return register(instance, key, data, true)
}

// Load loads the data for the given unsafe.Pointer.
// It is safe to call this function from multiple goroutines.
// It panics if the given pointer does not exist in the registry.
func Load(k unsafe.Pointer) any {
	instanceDataLock.Lock()
	defer instanceDataLock.Unlock()

	key := (*instanceDataKey)(k)

	fs, ok := instanceDataRegistry[*key]
	if !ok {
		panic("given pointer is not registered as instance data")
	}

	if fs.once {
		delete(instanceDataRegistry, *key)
	}

	return fs.data
}

func Delete(k unsafe.Pointer) {
	instanceDataLock.Lock()
	defer instanceDataLock.Unlock()

	key := (*instanceDataKey)(k)

	delete(instanceDataRegistry, *key)
}
