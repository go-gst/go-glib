package gobject

import "sync"

var objectCastingsLock sync.RWMutex
var objectCastings = make(map[Type]CastObjectFunc)

type CastObjectFunc func(inst *ObjectInstance) Object

// RegisterObjectCasting registers a casting function for the given type. This should normally be called by the generated bindings
// and is not intended to be used by user code.
//
// We need this to be able to "cast" go structs without touching the reference counting
func RegisterObjectCasting(t Type, fn CastObjectFunc) {
	if fn == nil {
		panic("cannot register nil casting function")
	}

	objectCastingsLock.Lock()
	defer objectCastingsLock.Unlock()
	if _, exists := objectCastings[t]; exists {
		panic("casting function already registered for type " + t.String())
	}
	objectCastings[t] = fn
}
