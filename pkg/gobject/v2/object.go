package gobject

import (
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
// static GObjectClass *_g_object_get_class(GObject *object) {
//   return (G_OBJECT_GET_CLASS(object));
// }
// static GType _g_type_from_instance(gpointer instance) {
//   return (G_TYPE_FROM_INSTANCE(instance));
// }
import "C"

// The base object type.
//
// This is an interface because the actual type will almost always extend
type Object interface {
	GoValueInitializer

	Emit(detailedSignal string, args ...any) any

	Connect(detailedSignal string, f interface{}) SignalHandle
	ConnectAfter(detailedSignal string, f interface{}) SignalHandle

	HandlerBlock(SignalHandle)
	HandlerUnblock(SignalHandle)
	HandlerDisconnect(SignalHandle)

	NotifyProperty(string, func(Object, *ParamSpec)) SignalHandle
	ObjectProperty(string) interface{}
	SetObjectProperty(string, interface{})
	SetObjectProperties(props map[string]any)
	FreezeNotify()
	ThawNotify()
	StopEmission(string)

	// Parent virtual methods:

	ParentConstructed()
	ParentFinalize()

	// internal methods:

	isFloating() bool

	// UnsafeLoadInstanceFromPrivateData is used internally by the subclassing code to load the instance from the private data
	// of the object. This is used to implement the virtual methods in the generated code.
	// This should not be called by user code.
	UnsafeLoadInstanceFromPrivateData() Object

	baseObject() *ObjectInstance
}

var _ Object = (*ObjectInstance)(nil)

const (
	TypeObject Type = C.G_TYPE_OBJECT
)

func init() {
	RegisterGValueMarshaler(TypeObject, marshalObject)

	RegisterObjectCasting(TypeObject, func(inst *ObjectInstance) Object {
		// this is the base type, so we can just return the instance
		return inst
	})
}

// marshalObject returns a concrete ObjectInstance, because this is only called when we do not know the actual
// extending type
func marshalObject(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(p))
	return UnsafeObjectFromGlibNone(unsafe.Pointer(c)), nil
}

// UnsafeObjectFromGlibNone is used to convert raw C object pointers to go while taking a reference.
// the returned Object is casted correctly and needs a manual cast by the user to the correct extending interface
//
// This is used by the bindings internally.
func UnsafeObjectFromGlibNone(p unsafe.Pointer) Object {
	obj := wrapObjectFinalized(p)

	if obj == nil {
		return nil
	}

	// if the object was floating this removes the floating ref.
	// if not, then this is equivalent to g_object_ref
	C.g_object_ref_sink(C.gpointer(obj.unsafe()))

	return obj.cast()
}

// UnsafeObjectFromGlibBorrow is used to convert raw C object pointers to go without taking a reference or touching the
// floating reference. The returned Object is casted correctly and needs a manual cast by the user to the correct extending interface
//
// This is used by the bindings internally.
func UnsafeObjectFromGlibBorrow(p unsafe.Pointer) Object {
	obj := wrapObject(p)

	if obj == nil {
		return nil
	}

	return obj.cast()
}

// UnsafeObjectFromGlibFull is used to convert raw C object pointers to go.
// the returned Object is casted correctly and needs a manual cast by the user to the correct extending interface
func UnsafeObjectFromGlibFull(p unsafe.Pointer) Object {
	obj := wrapObjectFinalized(p)

	if obj == nil {
		return nil
	}

	return obj.cast()
}

// UnsafeObjectRef increases the reference count of the given Object.
func UnsafeObjectRef(obj Object) {
	if obj == nil {
		return
	}

	C.g_object_ref(C.gpointer(obj.baseObject().unsafe()))
}

// UnsafeObjectUnref decreases the reference count of the given Object.
func UnsafeObjectUnref(obj Object) {
	if obj == nil {
		return
	}

	C.g_object_unref(C.gpointer(obj.baseObject().unsafe()))
	runtime.SetFinalizer(
		obj.baseObject().objectInstance,
		nil,
	)
}

// UnsafeObjectToGlibNone is used to convert the Object to C.
func UnsafeObjectToGlibNone(obj Object) unsafe.Pointer {
	if obj == nil {
		return nil
	}

	base := obj.baseObject()

	return base.unsafe()
}

// UnsafeObjectToGlibFull is used to convert the Object to C. This is used for cases where the receiver will not call ref but
// will call unref when the object is no longer needed. This means that we need to take another reference on the object
// to prevent memory corruption.
func UnsafeObjectToGlibFull(obj Object) unsafe.Pointer {
	if obj == nil {
		return nil
	}

	C.g_object_ref(C.gpointer(obj.baseObject().unsafe()))

	return UnsafeObjectToGlibNone(obj)
}

func wrapObject(p unsafe.Pointer) *ObjectInstance {
	if p == nil {
		return nil
	}
	return &ObjectInstance{
		objectInstance: &objectInstance{
			native: (*C.GObject)(p),
		},
	}
}

// wrapObjectFinalized returns a new object instance and attaches a cleanup to unref the c pointer on GC. if ref is true
// then a reference will be taken on the object.
func wrapObjectFinalized(p unsafe.Pointer) *ObjectInstance {
	obj := wrapObject(p)
	if obj == nil {
		return nil
	}

	runtime.SetFinalizer(
		obj.objectInstance,
		func(intern *objectInstance) {
			C.g_object_unref(C.gpointer(intern.native))
		},
	)

	return obj
}

type ObjectInstance struct {
	*objectInstance
}

// objectInstance is the object that is finalized
type objectInstance struct {
	native *C.GObject
}

// isFloating implements Object.
func (obj *ObjectInstance) isFloating() bool {
	return C.g_object_is_floating(C.gpointer(obj.unsafe())) != 0
}

// GoValueType implements GoValueInitializer.
func (obj *ObjectInstance) GoValueType() Type {
	// always use the type from the object instance, instead of the base type,
	// since this is inherited by all extending types
	return obj.typeFromInstance()
}

// SetGoValue implements GoValueInitializer.
func (obj *ObjectInstance) SetGoValue(v *Value) {
	v.SetObject(obj)
}

func (obj *ObjectInstance) unsafe() unsafe.Pointer {
	if obj == nil {
		return nil
	}
	if obj.objectInstance == nil {
		panic("this Object is invalid")
	}

	return unsafe.Pointer(obj.native)
}

func (obj *ObjectInstance) baseObject() *ObjectInstance {
	return obj
}

// ObjectProperty is a wrapper around g_object_get_property(). If the property's
// type cannot be resolved to a Go type, then InvalidValue is returned.
func (obj *ObjectInstance) ObjectProperty(name string) interface{} {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	t := obj.propertyType((*C.gchar)(cstr))
	if t == TypeInvalid {
		return InvalidValue
	}

	p := InitValue(t)

	C.g_object_get_property(obj.native, (*C.gchar)(cstr), p.native())
	runtime.KeepAlive(obj)

	return p.GoValue()
}

// SetObjectProperty is a wrapper around g_object_set_property().
func (obj *ObjectInstance) SetObjectProperty(name string, value interface{}) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	p := NewValue(value)

	C.g_object_set_property(obj.native, (*C.gchar)(cstr), p.native())
	runtime.KeepAlive(obj)
	runtime.KeepAlive(p)
}

// SetObjectProperties is a convinience function to set multiple properties at once.
func (obj *ObjectInstance) SetObjectProperties(props map[string]any) {
	for p, v := range props {
		obj.SetObjectProperty(p, v)
		runtime.KeepAlive(obj)
		runtime.KeepAlive(v)
	}
	runtime.KeepAlive(obj)
}

// NotifyProperty adds a handler that's called when the object's property is
// updated.
func (obj *ObjectInstance) NotifyProperty(property string, f func(Object, *ParamSpec)) SignalHandle {
	return obj.Connect("notify::"+property, f)
}

// FreezeNotify increases the freeze count on object. If the freeze count is
// non-zero, the emission of “notify” signals on object is stopped. The signals
// are queued until the freeze count is decreased to zero. Duplicate
// notifications are squashed so that at most one GObject::notify signal is
// emitted for each property modified while the object is frozen.
//
// This is necessary for accessors that modify multiple properties to prevent
// premature notification while the object is still being modified.
func (obj *ObjectInstance) FreezeNotify() {
	C.g_object_freeze_notify(obj.native)
	runtime.KeepAlive(obj)
}

// ThawNotify reverts the effect of a previous call to g_object_freeze_notify().
// The freeze count is decreased on object and when it reaches zero, queued
// “notify” signals are emitted.
//
// Duplicate notifications for each property are squashed so that at most one
// GObject::notify signal is emitted for each property, in the reverse order in
// which they have been queued.
//
// It is an error to call this function when the freeze count is zero.
func (obj *ObjectInstance) ThawNotify() {
	C.g_object_thaw_notify(obj.native)
	runtime.KeepAlive(obj)
}

// HandlerBlock is a wrapper around g_signal_handler_block().
func (obj *ObjectInstance) HandlerBlock(handle SignalHandle) {
	C.g_signal_handler_block(C.gpointer(obj.unsafe()), C.gulong(handle))
	runtime.KeepAlive(obj)
}

// HandlerUnblock is a wrapper around g_signal_handler_unblock().
func (obj *ObjectInstance) HandlerUnblock(handle SignalHandle) {
	C.g_signal_handler_unblock(C.gpointer(obj.unsafe()), C.gulong(handle))
	runtime.KeepAlive(obj)
}

// HandlerDisconnect is a wrapper around g_signal_handler_disconnect().
func (obj *ObjectInstance) HandlerDisconnect(handle SignalHandle) {
	C.g_signal_handler_disconnect(C.gpointer(obj.unsafe()), C.gulong(handle))
	runtime.KeepAlive(obj)
}

// StopEmission stops a signal’s current emission. It is a wrapper around
// g_signal_stop_emission_by_name().
func (obj *ObjectInstance) StopEmission(s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	C.g_signal_stop_emission_by_name(C.gpointer(obj.unsafe()), (*C.gchar)(cstr))
	runtime.KeepAlive(obj)
}

// PropertyType returns the Type of a property of the underlying GObject. If the
// property is missing, it will return TypeInvalid.
func (obj *ObjectInstance) PropertyType(name string) Type {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	return obj.propertyType(cstr)
}

func (obj *ObjectInstance) propertyType(cstr *C.gchar) Type {
	paramSpec := C.g_object_class_find_property(C._g_object_get_class(obj.native), (*C.gchar)(cstr))
	runtime.KeepAlive(obj)

	if paramSpec == nil {
		return TypeInvalid
	}

	return Type(paramSpec.value_type)
}

func (obj *ObjectInstance) unsafePrivateData() unsafe.Pointer {
	private := C.g_type_instance_get_private((*C.GTypeInstance)(obj.unsafe()), C.GType(obj.typeFromInstance()))

	return unsafe.Pointer(private)
}

// TypeFromInstance is a wrapper around g_type_from_instance().
func (obj *ObjectInstance) typeFromInstance() Type {
	return unsafeTypeFromObject(obj.unsafe())
}

func unsafeTypeFromObject(instance unsafe.Pointer) Type {
	c := C._g_type_from_instance(C.gpointer(instance))
	return Type(c)
}

// cast casts v to the concrete Go type (e.g. *Object to *gtk.Entry).
func (obj *ObjectInstance) cast() Object {
	if obj.unsafe() == nil {
		// nil-typed interface != non-nil-typed nil-value interface
		return nil
	}

	// re-implement the gvalue marshaling here that takes the type from the instance
	// and walks up the inheritance chain to find the correct casting function
	//
	// we MUST NOT use gvalue here, because that would take a reference on the object,
	// and that is something the caller should decide
	//
	// we also can't ref and unref, because a floating reference would be cleaned up
	//
	// we KISS here: just don't touch the reference at all!

	typeFromInstance := obj.typeFromInstance()

	for {
		objectCastingsLock.RLock()
		castFunc, exists := objectCastings[typeFromInstance]
		objectCastingsLock.RUnlock()

		if exists {
			return castFunc(obj)
		}

		if typeFromInstance == TypeObject {
			// panic here to never block or return nil, Object should always be handled
			panic("type object must have a casting function registered")
		}

		typeFromInstance = typeFromInstance.Parent()
	}
}
