package closure

import (
	"sync"
	"unsafe"
)

var closureRegistry sync.Map // unsafe.Pointer(*C.GClosure) -> *FuncStack

// Register registers the given GClosure callback. This panics if the
// GClosure is already registered.
func Register(gclosure unsafe.Pointer, callback *FuncStack) {
	if _, ok := closureRegistry.Load(gclosure); ok {
		panic("closure already registered")
	}

	closureRegistry.Store(gclosure, callback)

	profile.Add(uintptr(gclosure), 2)
}

func Load(gclosure unsafe.Pointer) *FuncStack {
	fs, ok := closureRegistry.Load(gclosure)
	if !ok {
		return nil
	}
	return fs.(*FuncStack)
}

// Delete deletes the given GClosure callback. This panics if the
// GClosure is not registered.
func Delete(gclosure unsafe.Pointer) {
	if _, ok := closureRegistry.Load(gclosure); !ok {
		panic("closure not registered")
	}
	closureRegistry.Delete(gclosure)
	profile.Remove(uintptr(gclosure))
}
