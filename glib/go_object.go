package glib

/*
#include "glib.go.h"

extern void   goObjectSetProperty  (GObject * object, guint property_id, const GValue * value, GParamSpec *pspec);
extern void   goObjectGetProperty  (GObject * object, guint property_id, GValue * value, GParamSpec * pspec);
extern void   goObjectConstructed  (GObject * object);
extern void   goObjectFinalize     (GObject * object, gpointer klass);

extern void   goClassInit     (gpointer g_class, gpointer class_data);
extern void   goInstanceInit  (GTypeInstance * instance, gpointer g_class);

void objectFinalize (GObject * object)
{
	GObjectClass *parent = g_type_class_peek_parent((G_OBJECT_GET_CLASS(object)));
	goObjectFinalize(object, G_OBJECT_GET_CLASS(object));
	parent->finalize(object);
}

void objectConstructed (GObject * object)
{
	GObjectClass *parent = g_type_class_peek_parent((G_OBJECT_GET_CLASS(object)));
	goObjectConstructed(object);
	parent->constructed(object);
}

void  setGObjectClassSetProperty  (void * klass)  { ((GObjectClass *)klass)->set_property = goObjectSetProperty; }
void  setGObjectClassGetProperty  (void * klass)  { ((GObjectClass *)klass)->get_property = goObjectGetProperty; }
void  setGObjectClassConstructed  (void * klass)  { ((GObjectClass *)klass)->constructed = objectConstructed; }
void  setGObjectClassFinalize     (void * klass)  { ((GObjectClass *)klass)->finalize = objectFinalize; }

void  cgoClassInit     (gpointer g_class, gpointer class_data)       { goClassInit(g_class, class_data); }
void  cgoInstanceInit  (GTypeInstance * instance, gpointer g_class)  { goInstanceInit(instance, g_class); }

GType  objectGType   (GObject *obj) { return G_OBJECT_TYPE(obj); };

*/
import "C"
import (
	"reflect"
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

// GoObject is an interface that abstracts on the GObject. In almost all cases at least SetProperty and GetProperty
// should be implemented by objects built from the go bindings.
type GoObject interface {
	// SetProperty should set the value of the property with the given id. ID is the index+1 of the parameter
	// in the order it was registered.
	SetProperty(obj *Object, id uint, value *Value)
	// GetProperty should retrieve the value of the property with the given id. ID is the index+1 of the parameter
	// in the order it was registered.
	GetProperty(obj *Object, id uint) *Value
	// Constructed is called when the Object has finished setting up.
	Constructed(*Object)
}

// GoObjectSubclass is an interface that abstracts on the GObjectClass. It should be implemented
// by plugins using the go bindings.
type GoObjectSubclass interface {
	// New should return a new instantiated GoElement ready to be used.
	New() GoObjectSubclass
	// TypeInit is called after the GType is registered and right before ClassInit. It is when the
	// element should add any interfaces it plans to implement.
	TypeInit(*TypeInstance)
	// ClassInit is called on the element after registering it with the type system. This is when the element
	// should install its properties and pad templates.
	ClassInit(*ObjectClass)
}

// TypeInstance is a loose binding around the glib GTypeInstance. It holds the information required to register
// various capabilities of a GoObjectSubclass.
type TypeInstance struct {
	GType  Type
	GoType GoObjectSubclass
}

// Interface can be implemented by extending packages and provides a the base type for the interface and
// a pointer to a C function that can be used for the interface_init in a GInterfaceInfo.
type Interface interface {
	Type() Type
	InitFunc(t *TypeInstance) unsafe.Pointer
}

// AddInterface will add an interface implementation for the type referenced by this object.
func (t *TypeInstance) AddInterface(iface Interface) {
	ifaceInfo := C.GInterfaceInfo{
		interface_data:     nil,
		interface_finalize: nil,
	}
	ifaceInfo.interface_init = C.GInterfaceInitFunc(iface.InitFunc(t))
	C.g_type_add_interface_static(
		(C.GType)(t.GType),
		(C.GType)(iface.Type()),
		&ifaceInfo,
	)
}

// FromObjectUnsafePrivate will return the GoObjectSubclass addressed in the private data of the given GObject.
func FromObjectUnsafePrivate(obj unsafe.Pointer) GoObjectSubclass {
	ptr := gopointer.Restore(privateFromObj(obj))
	return ptr.(GoObjectSubclass)
}

type classData struct {
	elem GoObjectSubclass
	ext  Extendable
}

// RegisterGoType is used to register an interface implemented in the Go runtime with the GType
// system. It takes the name to assign the type, the interface itself, and an Extendable denoting
// the subclasses it extends. It's the responsibility of packages using these bindings to implement
// Extendables that call up to the ExtendsObject.InitClass included in this package during their
// own implementation.
func RegisterGoType(name string, elem GoObjectSubclass, extendable Extendable) Type {
	registerMutex.Lock()
	defer registerMutex.Unlock()
	if registered, ok := registeredTypes[reflect.TypeOf(elem).String()]; ok {
		return registered
	}
	classData := &classData{
		elem: elem,
		ext:  extendable,
	}
	ptr := gopointer.Save(classData)

	typeInfo := (*C.GTypeInfo)(C.malloc(C.sizeof_GTypeInfo))
	defer C.free(unsafe.Pointer(typeInfo))

	typeInfo.base_init = nil
	typeInfo.base_finalize = nil
	typeInfo.class_size = C.gushort(extendable.ClassSize())
	typeInfo.class_finalize = nil
	typeInfo.class_init = C.GClassInitFunc(C.cgoClassInit)
	typeInfo.class_data = (C.gconstpointer)(ptr)
	typeInfo.instance_size = C.gushort(extendable.InstanceSize())
	typeInfo.n_preallocs = 0
	typeInfo.instance_init = C.GInstanceInitFunc(C.cgoInstanceInit)
	typeInfo.value_table = nil

	gtype := C.g_type_register_static(
		C.GType(extendable.Type()),
		(*C.gchar)(C.CString(name)),
		typeInfo,
		C.GTypeFlags(0),
	)
	elem.TypeInit(&TypeInstance{GType: Type(gtype), GoType: elem})
	registeredTypes[reflect.TypeOf(elem).String()] = Type(gtype)
	return Type(gtype)
}

// privateFromObj returns the actual value of the address we stored in the object's private data.
func privateFromObj(obj unsafe.Pointer) unsafe.Pointer {
	private := C.g_type_instance_get_private((*C.GTypeInstance)(obj), C.objectGType((*C.GObject)(obj)))
	privAddr := (*unsafe.Pointer)(unsafe.Pointer(private))
	return *privAddr
}

// Extendable is an interface implemented by extendable classes. It provides
// the methods necessary to setup the vmethods on the object it represents.
type Extendable interface {
	// Type should return the type of the extended object
	Type() Type
	// ClasSize should return the size of the extended class
	ClassSize() int64
	// InstanceSize should return the size of the object itself
	InstanceSize() int64
	// InitClass should take a pointer to a new subclass and a GoElement and override any
	// methods implemented by the GoElement in the subclass.
	InitClass(unsafe.Pointer, GoObjectSubclass)
}

// ExtendsObject signifies a GoElement that extends a GObject. It is the base Extendable
// that all other implementations derive from.
var ExtendsObject Extendable = &extendObject{}

type extendObject struct{}

func (e *extendObject) Type() Type          { return Type(C.g_object_get_type()) }
func (e *extendObject) ClassSize() int64    { return int64(C.sizeof_GObjectClass) }
func (e *extendObject) InstanceSize() int64 { return int64(C.sizeof_GObject) }

func (e *extendObject) InitClass(klass unsafe.Pointer, elem GoObjectSubclass) {
	C.setGObjectClassFinalize(klass)

	if _, ok := elem.(interface {
		SetProperty(obj *Object, id uint, value *Value)
	}); ok {
		C.setGObjectClassSetProperty(klass)
	}
	if _, ok := elem.(interface {
		GetProperty(obj *Object, id uint) *Value
	}); ok {
		C.setGObjectClassGetProperty(klass)
	}
	if _, ok := elem.(interface {
		Constructed(*Object)
	}); ok {
		C.setGObjectClassConstructed(klass)
	}
}
