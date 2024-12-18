package glib

/*
#include "glib.go.h"

extern gboolean goSignalAccumulator (GSignalInvocationHint* ihint, GValue* return_accu, const GValue* handler_return, gpointer data);
gboolean  cgoSignalAccumulator (GSignalInvocationHint* ihint, GValue* return_accu, const GValue* handler_return, gpointer data) {
  return goSignalAccumulator(ihint, return_accu, handler_return, data);
}
*/
import "C"
import (
	"unsafe"

	gopointer "github.com/go-gst/go-pointer"
)

// SignalFlags are used to specify a signal's behaviour.
type SignalFlags C.GSignalFlags

const (
	// SignalRunFirst: Invoke the object method handler in the first emission stage.
	SignalRunFirst SignalFlags = C.G_SIGNAL_RUN_FIRST
	// SignalRunLast: Invoke the object method handler in the third emission stage.
	SignalRunLast SignalFlags = C.G_SIGNAL_RUN_LAST
	// SignalRunCleanup: Invoke the object method handler in the last emission stage.
	SignalRunCleanup SignalFlags = C.G_SIGNAL_RUN_CLEANUP
	// SignalNoRecurse: Signals being emitted for an object while currently being in
	// emission for this very object will not be emitted recursively,
	// but instead cause the first emission to be restarted.
	SignalNoRecurse SignalFlags = C.G_SIGNAL_NO_RECURSE
	// SignalDetailed: This signal supports "::detail" appendices to the signal name
	// upon handler connections and emissions.
	SignalDetailed SignalFlags = C.G_SIGNAL_DETAILED
	// SignalAction: Action signals are signals that may freely be emitted on alive
	// objects from user code via g_signal_emit() and friends, without
	// the need of being embedded into extra code that performs pre or
	// post emission adjustments on the object. They can also be thought
	// of as object methods which can be called generically by
	// third-party code.
	SignalAction SignalFlags = C.G_SIGNAL_ACTION
	// SignalNoHooks: No emissions hooks are supported for this signal.
	SignalNoHooks SignalFlags = C.G_SIGNAL_NO_HOOKS
	// SignalMustCollect: Varargs signal emission will always collect the
	// arguments, even if there are no signal handlers connected.  Since 2.30.
	SignalMustCollect SignalFlags = C.G_SIGNAL_MUST_COLLECT
	// SignalDeprecated: The signal is deprecated and will be removed
	// in a future version. A warning will be generated if it is connected while
	// running with G_ENABLE_DIAGNOSTIC=1.  Since 2.32.
	SignalDeprecated SignalFlags = C.G_SIGNAL_DEPRECATED
	// SignalAccumulatorFirstRun: Only used in #GSignalAccumulator accumulator
	// functions for the #GSignalInvocationHint::run_type field to mark the first
	// call to the accumulator function for a signal emission.  Since 2.68.
	SignalAccumulatorFirstRun SignalFlags = C.G_SIGNAL_ACCUMULATOR_FIRST_RUN
)

type SignalInvocationHint struct {
	native *C.GSignalInvocationHint
}

func (s *SignalInvocationHint) SignalID() uint {
	return uint(s.native.signal_id)
}

func (s *SignalInvocationHint) Detail() Quark {
	return Quark(s.native.detail)
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

// NewSignal Creates a new signal. (This is usually done in the class initializer.) this is a wrapper around g_signal_newv
func NewSignal(
	name string,
	_type Type,
	flags SignalFlags,
	handler any,
	accumulator SignalAccumulator,
	param_types []Type,
	return_type Type,
) (*Signal, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	cparams := make([]C.GType, 0, len(param_types))

	for _, t := range param_types {
		cparams = append(cparams, C.GType(t))
	}

	classHandler, err := ClosureNew(handler)

	if err != nil {
		return nil, err
	}

	defer C.g_closure_unref(classHandler)

	var accudata C.gpointer
	var cAccumulator C.GSignalAccumulator

	if accumulator != nil {
		accudata = C.gpointer(gopointer.Save(accumulator))

		cAccumulator = C.GSignalAccumulator(C.cgoSignalAccumulator)
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

	return &Signal{name: name, signalId: signalID}, nil
}
