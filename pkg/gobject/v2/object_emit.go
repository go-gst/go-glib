package gobject

import (
	"fmt"
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
// static GValue *_alloc_gvalue_list(int n) {
//   GValue *valv;
//   valv = g_new0(GValue, n);
//   return (valv);
// }
// static void _val_list_insert(GValue *valv, int i, GValue *val) {
//   valv[i] = *val;
// }
import "C"

// Emit is a wrapper around g_signal_emitv() and emits the signal
// specified by the string s to an Object.  Arguments to callback
// functions connected to this signal must be specified in args.  Emit()
// returns an interface{} which contains the go equivalent of the C return value.
func (obj *ObjectInstance) Emit(s string, args ...any) any {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	t := obj.typeFromInstance()
	id := C.g_signal_lookup((*C.gchar)(cstr), C.GType(t))

	if id == 0 {
		panic(fmt.Sprintf("signal %s not found for type %s", s, t.Name()))
	}

	// query the signal info to determine the number of arguments and the return type
	var q C.GSignalQuery
	C.g_signal_query(id, &q)

	if len(args) != int(q.n_params) {
		panic(fmt.Sprintf("signal %s has %d parameters, but %d were passed", s, q.n_params, len(args)))
	}

	// FIXME: signal arg typ checking would be nice here, but currently this breaks passing nil objects, as their
	// value type is coming from the embedded GObject, which creates a nil dereference when the main pointer is nil
	// signalArgTypes := unsafe.Slice(q.param_types, q.n_params)

	// get the return type, remove the static scope flag first
	return_type := Type(q.return_type &^ C.G_SIGNAL_TYPE_STATIC_SCOPE)

	// Create array of this instance and arguments
	instanceAndParams := C._alloc_gvalue_list(C.int(len(args)) + 1)

	// Add args and valv
	instanceValue := NewValue(obj)

	C._val_list_insert(instanceAndParams, C.int(0), instanceValue.native())
	defer runtime.KeepAlive(instanceValue) // keep the value alive until the signal has been emitted

	for i := range args {
		// check the value type
		// argType := valueType(args[i])
		// requestedType := Type(signalArgTypes[i])

		// if argType != requestedType && !argType.IsA(requestedType) {
		// 	panic(fmt.Sprintf("signal emit argument %d has wrong type, expected %s (%s), got %s (%s)", i, requestedType.Name(), FundamentalType(requestedType).Name(), argType.Name(), FundamentalType(argType).Name()))
		// }

		valueArg := NewValue(args[i])
		C._val_list_insert(instanceAndParams, C.int(i+1), valueArg.native())
		defer runtime.KeepAlive(valueArg) // keep the value alive until the signal has been emitted
	}

	// free the valv array after the values have been freed
	defer C.g_free(C.gpointer(instanceAndParams))

	if return_type != TypeInvalid && return_type != TypeNone {
		// the return value must have the correct type set
		ret := InitValue(return_type)
		defer runtime.KeepAlive(ret) // keep the value alive until the signal has been emitted

		C.g_signal_emitv(instanceAndParams, id, C.GQuark(0), ret.native())

		return ret.GoValue()
	}

	// signal has no return value
	C.g_signal_emitv(instanceAndParams, id, C.GQuark(0), nil)

	return nil
}
