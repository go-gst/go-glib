package glib

/*
#include "glib.go.h"

GObjectClass *  getGObjectClass (void * p)  { return (G_OBJECT_GET_CLASS(p)); }
*/
import "C"
import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"unsafe"
)

// Object is a representation of GLib's GObject.
type Object struct {
	GObject *C.GObject
}

// NewObjectWithProperties is a wrapper around `g_object_new_with_properties`
//
// see https://docs.gtk.org/gobject/ctor.Object.new_with_properties.html for more details
func NewObjectWithProperties(_type Type, properties map[string]interface{}) (*Object, error) {
	props := make([]*C.char, 0)
	values := make([]C.GValue, 0)

	for p, v := range properties {
		cpropName := C.CString(p)
		defer C.free(unsafe.Pointer(cpropName))

		props = append(props, cpropName)

		value, err := GValue(v)

		if err != nil {
			return nil, err
		}

		// value goes out of scope, but the finalizer must not run until the cgo call is finished
		defer runtime.KeepAlive(value)

		values = append(values, *(*C.GValue)(value.Unsafe()))
	}

	propCount := C.uint(len(properties))
	cProps := unsafe.SliceData(props)
	cPropValues := unsafe.SliceData(values)

	obj := C.g_object_new_with_properties(C.GType(_type), propCount, cProps, cPropValues)

	if obj == nil {
		return nil, fmt.Errorf("could not create object")
	}
	return TransferFull(unsafe.Pointer(obj)), nil
}

func (v *Object) toGObject() *C.GObject {
	if v == nil {
		return nil
	}
	return v.native()
}

func (v *Object) toObject() *Object {
	return v
}

// newObject creates a new Object from a GObject pointer.
func newObject(p *C.GObject) *Object {
	return &Object{GObject: p}
}

// native returns a pointer to the underlying GObject.
func (v *Object) native() *C.GObject {
	if v == nil || v.GObject == nil {
		return nil
	}
	p := unsafe.Pointer(v.GObject)
	return C.toGObject(p)
}

// goValue converts a *Object to a Go type (e.g. *Object => *gtk.Entry).
// It is used in goMarshal to convert generic GObject parameters to
// signal handlers to the actual types expected by the signal handler.
func (v *Object) goValue() (interface{}, error) {
	objType := Type(C._g_type_from_instance(C.gpointer(v.native())))
	f, err := gValueMarshalers.lookupType(objType)
	if err != nil {
		return nil, err
	}

	// The marshalers expect Values, not Objects
	val, err := ValueInit(objType)
	if err != nil {
		return nil, err
	}
	val.SetInstance(unsafe.Pointer(v.GObject))
	rv, err := f(unsafe.Pointer(val.native()))
	return rv, err
}

// Take wraps a unsafe.Pointer as a glib.Object, taking ownership of it.
// If the object is a floating reference a RefSink is taken, otherwise a
// Ref. A runtime finalizer is placed on the object to clear the ref
// when the object leaves scope.
func Take(ptr unsafe.Pointer) *Object {
	obj := newObject(ToGObject(ptr))

	if obj.IsFloating() {
		obj.RefSink()
	} else {
		obj.Ref()
	}

	runtime.SetFinalizer(obj, (*Object).Unref)
	return obj
}

// TransferNone is an alias to Take.
func TransferNone(ptr unsafe.Pointer) *Object { return Take(ptr) }

// TransferFull wraps a unsafe.Pointer as a glib.Object, taking ownership of it.
// it does not increase the ref count on the object. A finalizer is placed on the object
// to clear the transferred ref.
func TransferFull(ptr unsafe.Pointer) *Object {
	obj := newObject(ToGObject(ptr))
	runtime.SetFinalizer(obj, (*Object).Unref)
	return obj
}

// Native returns a pointer to the underlying GObject.
func (v *Object) Native() unsafe.Pointer {
	return unsafe.Pointer(v.native())
}

// Unsafe returns the unsafe pointer to the underlying object. This method is primarily
// for internal usage and is exposed for visibility in other packages.
func (v *Object) Unsafe() unsafe.Pointer {
	if v == nil || v.GObject == nil {
		return nil
	}
	return unsafe.Pointer(v.GObject)
}

// Class returns the GObjectClass of this instance.
func (v *Object) Class() *ObjectClass {
	return &ObjectClass{ptr: C.getGObjectClass(v.Unsafe())}
}

// IsA is a wrapper around g_type_is_a().
func (v *Object) IsA(typ Type) bool {
	return gobool(C.g_type_is_a(C.GType(v.TypeFromInstance()), C.GType(typ)))
}

// TypeFromInstance is a wrapper around g_type_from_instance().
func (v *Object) TypeFromInstance() Type {
	c := C._g_type_from_instance(C.gpointer(unsafe.Pointer(v.native())))
	return Type(c)
}

// ToGObject type converts an unsafe.Pointer as a native C GObject.
// This function is exported for visibility in other go-gst packages and
// is not meant to be used by applications.
func ToGObject(p unsafe.Pointer) *C.GObject {
	return C.toGObject(p)
}

// Ref is a wrapper around g_object_ref().
func (v *Object) Ref() *Object {
	gObjectProfile.Add(v, 1)
	C.g_object_ref(C.gpointer(v.GObject))
	return v
}

// Unref is a wrapper around g_object_unref().
func (v *Object) Unref() {
	gObjectProfile.Remove(v)
	C.g_object_unref(C.gpointer(v.GObject))
}

// RefSink is a wrapper around g_object_ref_sink().
func (v *Object) RefSink() {
	gObjectProfile.Add(v, 1)
	C.g_object_ref_sink(C.gpointer(v.GObject))
}

// IsFloating is a wrapper around g_object_is_floating().
func (v *Object) IsFloating() bool {
	c := C.g_object_is_floating(C.gpointer(v.GObject))
	return gobool(c)
}

// ForceFloating is a wrapper around g_object_force_floating().
func (v *Object) ForceFloating() {
	gObjectProfile.Remove(v)
	C.g_object_force_floating(v.GObject)
}

// Notify is a wrapper around g_object_notify().
func (v *Object) Notify(paramName string) {
	cstr := C.CString(paramName)
	defer C.free(unsafe.Pointer(cstr))
	C.g_object_notify(v.GObject, cstr)
}

// NotifyByPspec is a wrapper around g_object_notify_by_pspec().
func (v *Object) NotifyByPspec(pspec *ParamSpec) {
	C.g_object_notify_by_pspec(v.GObject, pspec.paramSpec)
}

// StopEmission is a wrapper around g_signal_stop_emission_by_name().
func (v *Object) StopEmission(s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	C.g_signal_stop_emission_by_name((C.gpointer)(v.GObject),
		(*C.gchar)(cstr))
}

// Set is a wrapper around g_object_set().  However, unlike
// g_object_set(), this function only sets one name value pair.  Make
// multiple calls to this function to set multiple properties.
func (v *Object) Set(name string, value interface{}) error {
	return v.SetProperty(name, value)
}

// GetPropertyType returns the Type of a property of the underlying GObject.
// If the property is missing it will return TYPE_INVALID and an error.
func (v *Object) GetPropertyType(name string) (Type, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	paramSpec := C.g_object_class_find_property(C._g_object_get_class(v.native()), (*C.gchar)(cstr))
	if paramSpec == nil {
		return TYPE_INVALID, errors.New("couldn't find Property")
	}
	return Type(paramSpec.value_type), nil
}

// GetProperty is a wrapper around g_object_get_property().
func (v *Object) GetProperty(name string) (interface{}, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	t, err := v.GetPropertyType(name)
	if err != nil {
		return nil, err
	}

	p, err := ValueInit(t)
	if err != nil {
		return nil, errors.New("unable to allocate value")
	}
	C.g_object_get_property(v.GObject, (*C.gchar)(cstr), p.native())
	return p.GoValue()
}

// SetProperty is a wrapper around g_object_set_property(). It attempts to convert
// the given Go value to a GValue before setting the property.
func (v *Object) SetProperty(name string, value interface{}) error {
	if _, ok := value.(Object); ok {
		value = value.(Object).GObject
	}
	p, err := gValue(value)
	if err != nil {
		return fmt.Errorf("unable to perform type conversion: %s", err.Error())
	}
	return v.SetPropertyValue(name, p)
}

// SetPropertyValue is like SetProperty except it operates on native
// GValues instead of first trying to convert from a Go value.
func (v *Object) SetPropertyValue(name string, value *Value) error {
	propType, err := v.GetPropertyType(name)
	if err != nil {
		return err
	}
	valType, _, err := value.Type()
	if err != nil {
		return err
	}
	if valType != propType {
		return fmt.Errorf("invalid type %s for property %s", value.TypeName(), name)
	}
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))
	C.g_object_set_property(v.GObject, (*C.gchar)(cstr), value.native())

	// the value object must be alive until after g_object_set_property finished, or else we risk a SIGSEGV
	// when the value is extracted into a golang defined custom element
	runtime.KeepAlive(value)

	return nil
}

// ListInterfaces returns the interfaces associated with this object.
func (v *Object) ListInterfaces() []string {
	var size C.guint
	ifaces := C.g_type_interfaces(C.gsize(v.TypeFromInstance()), &size)
	if int(size) == 0 {
		return nil
	}
	defer C.g_free((C.gpointer)(ifaces))
	out := make([]string, int(size))

	for _, t := range (*[(math.MaxInt32 - 1) / unsafe.Sizeof(int(0))]int)(unsafe.Pointer(ifaces))[:size:size] {
		out = append(out, Type(t).Name())
	}
	return out
}

/*
* GObject Signals
 */
var ErrSignalNotFound = errors.New("signal not found")
var ErrSignalWrongNumberOfArgs = errors.New("wrong number of arguments")

// Emit is a wrapper around g_signal_emitv() and emits the signal
// specified by the string s to an Object.  Arguments to callback
// functions connected to this signal must be specified in args.  Emit()
// returns an interface{} which contains the go equivalent of the C return value.
//
// Make sure that the Types are known to go-glib. Special types need to be registered with
// RegisterGValueMarshalers before calling Emit.
func (v *Object) Emit(s string, args ...interface{}) (interface{}, error) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	t := v.TypeFromInstance()
	id := C.g_signal_lookup((*C.gchar)(cstr), C.GType(t))

	if id == 0 {
		return nil, ErrSignalNotFound
	}

	// query the signal info to determine the number of arguments and the return type
	var q C.GSignalQuery
	C.g_signal_query(id, &q)

	if len(args) != int(q.n_params) {
		return nil, fmt.Errorf("%w for signal %s: expected %d, got %d", ErrSignalWrongNumberOfArgs, s, q.n_params, len(args))
	}

	// get the return type, remove the static scope flag first
	return_type := Type(q.return_type &^ C.G_SIGNAL_TYPE_STATIC_SCOPE)

	// Create array of this instance and arguments
	instanceAndParams := C._alloc_gvalue_list(C.int(len(args)) + 1)

	// Add args and valv
	instanceValue, err := GValue(v)
	if err != nil {
		return nil, errors.New("error converting Object to GValue: " + err.Error())
	}
	C._val_list_insert(instanceAndParams, C.int(0), instanceValue.native())
	defer runtime.KeepAlive(instanceValue) // keep the value alive until the signal has been emitted

	for i := range args {
		valueArg, err := GValue(args[i])
		if err != nil {
			return nil, fmt.Errorf("error converting arg %d to GValue: %s", i, err.Error())
		}
		C._val_list_insert(instanceAndParams, C.int(i+1), valueArg.native())
		defer runtime.KeepAlive(valueArg) // keep the value alive until the signal has been emitted
	}

	// free the valv array after the values have been freed
	defer C.g_free(C.gpointer(instanceAndParams))

	// check the values types against the signals types
	values := unsafe.Slice(instanceAndParams, len(args)+1)
	signalArgTypes := unsafe.Slice(q.param_types, q.n_params)
	for i := range len(args) {
		v := ValueFromNative(unsafe.Pointer(&values[i+1]))

		actual, fundamental, _ := v.Type()
		requestedType := Type(signalArgTypes[i])

		if actual != requestedType && fundamental != requestedType {
			return nil, fmt.Errorf("signal emit argument %d has wrong type, expected %s, got %s (%s)", i, requestedType.Name(), actual.Name(), fundamental.Name())
		}
	}

	if return_type != TYPE_INVALID && return_type != TYPE_NONE {
		// the return value must have the correct type set
		ret, err := ValueInit(return_type)
		if err != nil {
			return nil, errors.New("error creating Value for return value")
		}
		C.g_signal_emitv(instanceAndParams, id, C.GQuark(0), ret.native())

		return ret.GoValue()
	}

	// signal has no return value
	C.g_signal_emitv(instanceAndParams, id, C.GQuark(0), nil)

	return nil, nil
}

// HandlerBlock is a wrapper around g_signal_handler_block().
func (v *Object) HandlerBlock(handle SignalHandle) {
	C.g_signal_handler_block(C.gpointer(v.GObject), C.gulong(handle))
}

// HandlerUnblock is a wrapper around g_signal_handler_unblock().
func (v *Object) HandlerUnblock(handle SignalHandle) {
	C.g_signal_handler_unblock(C.gpointer(v.GObject), C.gulong(handle))
}

// HandlerDisconnect is a wrapper around g_signal_handler_disconnect().
func (v *Object) HandlerDisconnect(handle SignalHandle) {
	C.g_signal_handler_disconnect(C.gpointer(v.GObject), C.gulong(handle))
}

// WithTransferOriginal can be used to capture an object from transfer-none
// with a RefSink, and restore the original floating state of the ref after
// the given function's execution. Strictly speaking this is not thread safe,
// since additional references can be taken on the object elsewhere while the
// closure is executing. But for the lack of a better method for handling
// virtual methods this will suffice for now.
func (v *Object) WithTransferOriginal(f func()) {
	wasFloating := v.IsFloating()
	v.RefSink()
	defer func() {
		if wasFloating {
			v.ForceFloating()
		} else {
			v.Unref()
		}
	}()
	f()
}

// Keep will call runtime.KeepAlive on this or the extending object. It is useful for blocking
// a pending finalizer on this instance from firing and freeing the underlying C object.
// This is needed in the bindings where the Object goes out of scope but the C pointer is still needed.
func (v *Object) Keep() { runtime.KeepAlive(v) }

// GetPrivate returns a pointer to the private data stored inside this object.
func (v *Object) GetPrivate() unsafe.Pointer {
	private := C.g_type_instance_get_private((*C.GTypeInstance)(v.Unsafe()), C.objectGType(v.GObject))
	if private == nil {
		return nil
	}
	return unsafe.Pointer(private)
}

// WithPointerTransferOriginal is a convenience wrapper for wrapping the given pointer
// in an object, capturing the ref state, executing the given function with that object,
// and then restoring the original state. It is intended to be used with objects that were
// extended via the bindings. If the Object has an instantiated Go counterpart, it will be
// sent to the function as well, otherwise GoObjectSubclass will be nil.
//
// See WithTransferOriginal for more information.
func WithPointerTransferOriginal(o unsafe.Pointer, f func(*Object, GoObjectSubclass)) {
	obj := wrapObjectClean(o)
	obj.WithTransferOriginal(func() {
		f(obj, FromObjectUnsafePrivate(o))
	})
}

// WithPointerTransferNone will take a pointer to an object retrieved with transfer-none and call
// the corresponding function with it wrapped in an Object. If the object has an instantiated
// Go counterpart, it will be sent to the function as well. It is an alternative to using finalizers
// around bindings calls.
func WithPointerTransferNone(o unsafe.Pointer, f func(*Object, GoObjectSubclass)) {
	obj := wrapObjectClean(o)
	if obj.IsFloating() {
		obj.RefSink()
	} else {
		obj.Ref()
	}
	defer obj.Unref()
	f(obj, FromObjectUnsafePrivate(o))
}

// WithPointerTransferFull will take a pointer to an object retrieved with transfer-full and call
// the corresponding function with it wrapped in an Object. If the object has an instantiated
// Go counterpart, it will be sent to the function as well. It is an alternative to using finalizers
// around binding calls.
func WithPointerTransferFull(o unsafe.Pointer, f func(*Object, GoObjectSubclass)) {
	obj := wrapObjectClean(o)
	defer obj.Unref()
	f(obj, FromObjectUnsafePrivate(o))
}

func wrapObjectClean(ptr unsafe.Pointer) *Object {
	obj := &Object{ToGObject(ptr)}
	return obj
}

// Wrapper function for new objects with reference management.
func wrapObject(ptr unsafe.Pointer) *Object {
	obj := &Object{ToGObject(ptr)}

	if obj.IsFloating() {
		obj.RefSink()
	} else {
		obj.Ref()
	}

	runtime.SetFinalizer(obj, (*Object).Unref)
	return obj
}
