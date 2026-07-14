package gobject

import (
	"sync"
	"unsafe"
)

var registryLock sync.Mutex
var registeredSubclasses = make(map[unsafe.Pointer]*subClassData)

// registerSubclassData registers the subclass data for the given class.
func registerSubclassData(class unsafe.Pointer, data *subClassData) {
	registryLock.Lock()
	defer registryLock.Unlock()

	if _, ok := registeredSubclasses[class]; ok {
		panic("class is already registered")
	}

	registeredSubclasses[class] = data
}

// dataFromClass returns the subclass data for the given class.
func dataFromClass(class unsafe.Pointer) *subClassData {
	registryLock.Lock()
	defer registryLock.Unlock()

	data, ok := registeredSubclasses[class]
	if !ok {
		panic("class is not registered")
	}

	return data
}

var instanceID uint64 = 0
var instancesLock sync.RWMutex
var activeInstances = make(map[uint64]Object)

// saveInstanceInPrivateData saves the instance in the private data of the given object.
func saveInstanceInPrivateData(obj Object) {
	instancesLock.Lock()
	defer instancesLock.Unlock()

	baseObj := obj.baseObject()

	instanceID++

	privatePtr := baseObj.unsafePrivateData()

	if privatePtr == nil {
		panic("private data is nil")
	}

	private := (*uint64)(privatePtr)

	*private = instanceID
	activeInstances[instanceID] = obj
}

// UnsafeLoadInstanceFromPrivateData loads the instance from the private data of the given object.
func (obj *ObjectInstance) UnsafeLoadInstanceFromPrivateData() Object {
	instancesLock.RLock()
	defer instancesLock.RUnlock()

	baseObj := obj.baseObject()

	privatePtr := baseObj.unsafePrivateData()

	if privatePtr == nil {
		panic("private data is nil")
	}

	private := (*uint64)(privatePtr)

	instance, ok := activeInstances[*private]
	if !ok {
		panic("instance not found")
	}

	return instance
}

// removeInstanceFromPrivateData removes the instance from the private data of the given object.
func removeInstanceFromPrivateData(obj Object) {
	instancesLock.Lock()
	defer instancesLock.Unlock()

	baseObj := obj.baseObject()

	privatePtr := baseObj.unsafePrivateData()

	if privatePtr == nil {
		panic("private data is nil")
	}

	private := (*uint64)(privatePtr)

	if _, ok := activeInstances[*private]; !ok {
		panic("tried to remove instance that does not exist")
	}

	delete(activeInstances, *private)
}
