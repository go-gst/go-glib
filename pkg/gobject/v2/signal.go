package gobject

import (
	"unsafe"

	"github.com/go-gst/go-glib/pkg/core/closure"
	"github.com/go-gst/go-glib/pkg/core/userdata"
	"github.com/go-gst/go-glib/pkg/glib/v2"
)

// #include <glib-object.h>
// extern void _goglib_signalAccumulator(GSignalInvocationHint*, GValue*, GValue*, gpointer);
import "C"

func (s *SignalInvocationHint) SignalID() uint {
	return uint(s.native.signal_id)
}

func (s *SignalInvocationHint) Detail() glib.Quark {
	return glib.Quark(s.native.detail)
}

func (s *SignalInvocationHint) RunType() SignalFlags {
	return SignalFlags(s.native.run_type)
}

// SignalAccumulator is a special callback function that can be used to collect return values of the various callbacks that are called during a signal emission.
type SignalAccumulator func(ihint *SignalInvocationHint, return_accu *Value, handler_return *Value) bool

type Signal struct {
	name     string
	signalId C.guint
}

// NewSignal creates a new signal. This can either be directly called in class initializer of a new subclass, or by the bindings when registering a new signal
// through the signal map while adding a new subclass. this is a wrapper around g_signal_newv
func NewSignal(
	name string,
	_type Type,
	flags SignalFlags,
	handler any,
	accumulator SignalAccumulator,
	param_types []Type,
	return_type Type,
) *Signal {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cparams := make([]C.GType, 0, len(param_types))

	for _, t := range param_types {
		cparams = append(cparams, C.GType(t))
	}

	var classHandler *C.GClosure

	if handler != nil {
		classHandler := closureNew()
		fs := closure.NewFuncStack(handler, 2)

		closure.Register(unsafe.Pointer(classHandler), fs)

		defer C.g_closure_unref(classHandler)
	}

	var accudata C.gpointer
	var cAccumulator C.GSignalAccumulator

	if accumulator != nil {
		accudata = C.gpointer(userdata.Register(accumulator))

		cAccumulator = (C.GSignalAccumulator)((*[0]byte)(C._goglib_signalAccumulator))
	}

	signalID := C.g_signal_newv(
		cname,
		C.GType(_type),
		C.GSignalFlags(flags),
		classHandler,
		cAccumulator,
		accudata,
		nil, // no marshaller needed
		C.GType(return_type),
		C.uint(len(cparams)),
		unsafe.SliceData(cparams),
	)

	return &Signal{name: name, signalId: signalID}
}
