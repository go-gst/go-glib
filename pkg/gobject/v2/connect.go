package gobject

import (
	"unsafe"

	"github.com/go-gst/go-glib/pkg/core/closure"
)

// #include <glib.h>
// #include <glib-object.h>
// extern void _goglib_removeClosure(GObject*, GClosure*);
// extern void _goglib_goMarshal(GClosure*, GValue*, guint, GValue*, gpointer, gpointer);
import "C"

// SignalHandle is the identifier for a connected glib signal on a specific object. It is
// returned when connecting a signal and can be used to disconnect the signal.
//
// Important: This is only unique per object. Different objects can return the same SignalHandle
// for different signals.
type SignalHandle uint

// Connect is a wrapper around g_signal_connect_closure(). f must be a function
// with at least one parameter matching the type it is connected to.
//
// It is optional to list the rest of the required types from GTK, as values
// that don't fit into the function parameter will simply be ignored; however,
// extraneous types will trigger a runtime panic. Arguments for f must be a
// matching Go equivalent type for the C callback, or an interface type which
// the value may be packed in. If the type is not suitable, a runtime panic will
// occur when the signal is emitted.
func (obj *ObjectInstance) Connect(detailedSignal string, f interface{}) SignalHandle {
	return obj.connectClosure(false, detailedSignal, f)
}

// ConnectAfter is a wrapper around g_signal_connect_closure(). The difference
// between Connect and ConnectAfter is that the latter will be invoked after the
// default handler, not before. For more information, refer to Connect.
func (obj *ObjectInstance) ConnectAfter(detailedSignal string, f interface{}) SignalHandle {
	return obj.connectClosure(true, detailedSignal, f)
}

func (obj *ObjectInstance) connectClosure(after bool, detailedSignal string, f interface{}) SignalHandle {
	// TODO: check if the signal is valid and if the function signature is valid for the signal handler

	fs := closure.NewFuncStack(f, 2)

	cstr := C.CString(detailedSignal)
	defer C.free(unsafe.Pointer(cstr))

	gclosure := closureNew()
	defer C.g_closure_unref(gclosure)

	closure.Register(unsafe.Pointer(gclosure), fs)

	c := C.g_signal_connect_closure(C.gpointer(obj.unsafe()), (*C.gchar)(cstr), gclosure, gbool(after))

	return SignalHandle(c)
}

// closureNew constructs a new GClosure object that gets the correct marshaller
// and finalizer set. The returned GClosure is owned by the caller and must be
// unref'd when no longer needed.
func closureNew() *C.GClosure {
	gclosure := C.g_closure_new_simple(C.sizeof_GClosure, nil)

	C.g_closure_set_meta_marshal(gclosure, nil, (*[0]byte)(C._goglib_goMarshal))
	C.g_closure_add_finalize_notifier(gclosure, nil, (*[0]byte)(C._goglib_removeClosure))

	C.g_closure_ref(gclosure)
	C.g_closure_sink(gclosure)

	return gclosure
}

func gbool(b bool) C.gboolean {
	if b {
		return 1
	}
	return 0
}
