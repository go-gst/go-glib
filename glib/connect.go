package glib

// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import (
	"errors"
	"reflect"
	"runtime/cgo"
	"unsafe"
)

/*
 * Events
 */

// SignalHandle is the identifier for a connected glib signal on a specific object. It is
// returned when connecting a signal and can be used to disconnect the signal.
//
// Important: This is only unique per object. Different objects can return the same SignalHandle
// for different signals.
type SignalHandle uint

func (v *Object) connectClosure(after bool, detailedSignal string, f interface{}, userData ...interface{}) (SignalHandle, error) {
	if len(userData) > 1 {
		return 0, errors.New("userData len must be 0 or 1")
	}

	cstr := C.CString(detailedSignal)
	defer C.free(unsafe.Pointer(cstr))

	closure, err := ClosureNew(f, userData...)
	if err != nil {
		return 0, err
	}

	c := C.g_signal_connect_closure(C.gpointer(v.native()),
		(*C.gchar)(cstr), closure, gbool(after))
	handle := SignalHandle(c)

	return handle, nil
}

// Connect is a wrapper around g_signal_connect_closure().  f must be
// a function with a signaure matching the callback signature for
// detailedSignal.  userData must either 0 or 1 elements which can
// be optionally passed to f.  If f takes less arguments than it is
// passed from the GLib runtime, the extra arguments are ignored.
//
// Arguments for f must be a matching Go equivalent type for the
// C callback, or an interface type which the value may be packed in.
// If the type is not suitable, a runtime panic will occur when the
// signal is emitted.
func (v *Object) Connect(detailedSignal string, f interface{}, userData ...interface{}) (SignalHandle, error) {
	return v.connectClosure(false, detailedSignal, f, userData...)
}

// ConnectAfter is a wrapper around g_signal_connect_closure().  f must be
// a function with a signaure matching the callback signature for
// detailedSignal.  userData must either 0 or 1 elements which can
// be optionally passed to f.  If f takes less arguments than it is
// passed from the GLib runtime, the extra arguments are ignored.
//
// Arguments for f must be a matching Go equivalent type for the
// C callback, or an interface type which the value may be packed in.
// If the type is not suitable, a runtime panic will occur when the
// signal is emitted.
//
// The difference between Connect and ConnectAfter is that the latter
// will be invoked after the default handler, not before.
func (v *Object) ConnectAfter(detailedSignal string, f interface{}, userData ...interface{}) (SignalHandle, error) {
	return v.connectClosure(true, detailedSignal, f, userData...)
}

// ClosureNew creates a new GClosure with the given function f. The returned closure is floating. This
// is useful so that the finalizer of the closure gets automatically called when the signal is disconnected.
//
// It's exported for visibility to go-gst packages and shouldn't be used in application code.
func ClosureNew(f interface{}, marshalData ...interface{}) (*C.GClosure, error) {
	// Create a reflect.Value from f.  This is called when the
	// returned GClosure runs.
	rf := reflect.ValueOf(f)

	// Create closure context which points to the reflected func.
	cc := closureContext{rf: rf}

	// Closures can only be created from funcs.
	if rf.Type().Kind() != reflect.Func {
		return nil, errors.New("value is not a func")
	}

	if len(marshalData) > 0 {
		cc.userData = reflect.ValueOf(marshalData[0])
	}

	// save the closure context in the closure itself
	ccHandle := cgo.NewHandle(&cc)

	c := C._g_closure_new(C.guint(ccHandle))

	C.g_closure_ref(c)
	C.g_closure_sink(c)

	return c, nil
}

// removeClosure removes the go function allowing the GC to collect it.
//
//export removeClosure
func removeClosure(data C.gpointer, closure *C.GClosure) {
	ccHandle := cgo.Handle(*(*C.guint)(data))

	ccHandle.Delete()

	C.free(unsafe.Pointer(data))
	C.g_closure_unref(closure)
}
