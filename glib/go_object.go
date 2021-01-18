package glib

/*
#include "glib.go.h"

extern void   goObjectSetProperty  (GObject * object, guint property_id, const GValue * value, GParamSpec *pspec);
extern void   goObjectGetProperty  (GObject * object, guint property_id, GValue * value, GParamSpec * pspec);
extern void   goObjectConstructed  (GObject * object);
extern void   goObjectFinalize     (GObject * object, gpointer klass);

extern void   goClassInit     (gpointer g_class, gpointer class_data);
extern void   goInstanceInit  (GTypeInstance * instance, gpointer g_class);
extern void   goInterfaceInit (gpointer iface, gpointer iface_data);

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

void  cgoClassInit      (gpointer g_class, gpointer class_data)       { goClassInit(g_class, class_data); }
void  cgoInstanceInit   (GTypeInstance * instance, gpointer g_class)  { goInstanceInit(instance, g_class); }
void  cgoInterfaceInit  (gpointer iface, gpointer iface_data)         { goInterfaceInit(iface, iface_data); }

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

// GoObjectSubclass is an interface that abstracts on the GObjectClass. It is the minimum that should be implemented
// by Go types that get registered as GTypes. For more information see RegisterGoType.
type GoObjectSubclass interface {
	// New should return a new instantiated GoElement ready to be used.
	New() GoObjectSubclass
	// ClassInit is called on the element after registering it with the type system. This is when the element
	// should install its properties and metadata.
	ClassInit(*ObjectClass)
}

// TypeInstance is a loose binding around the glib GTypeInstance. It holds the information required to register
// various capabilities of a GoObjectSubclass.
type TypeInstance struct {
	GType         Type
	GTypeInstance unsafe.Pointer
	GoType        GoObjectSubclass
}

// InterfaceInitFunc is a method to be implemented and returned by Interfaces. It is called
// with the initializing instance.
type InterfaceInitFunc func(*TypeInstance)

// Interface can be implemented by extending packages. They provide the base type for the interface and
// a function to call during interface_init retrieved by InitFunc. The function returned by InitFunc is
// called after the class has already been initialized. It is passed a TypeInstance populated with the GType
// corresponding to the Go object, a pointer to the underlying C object, and a pointer to the initialized
// Go object. When the object is actually used, a pointer to it can be retrieved from the C object with
// FromObjectUnsafePrivate.
//
// The user of the Interface is responsible for implementing the methods required by the interface. The GoType
// provided to the InterfaceInitFunc will be the object that is expected to carry the implementation.
type Interface interface {
	Type() Type
	InitFunc() InterfaceInitFunc
}

type interfaceData struct {
	init     InterfaceInitFunc
	instance *TypeInstance
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
//
// Interfaces are optional and flags additional interfaces as implemented on the class. Similar to the
// extendables, libraries using these bindings can implement the Interface interface to provide support
// for other GInterfaces.
func RegisterGoType(name string, elem GoObjectSubclass, extendable Extendable, interfaces ...Interface) Type {
	registerMutex.Lock()
	defer registerMutex.Unlock()
	if registered, ok := registeredTypes[reflect.TypeOf(elem).String()]; ok {
		return registered
	}
	classData := &classData{
		elem: elem.New(),
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
	for _, iface := range interfaces {
		gofuncPtr := gopointer.Save(&interfaceData{
			init: iface.InitFunc(),
			instance: &TypeInstance{
				GType:  Type(gtype),
				GoType: classData.elem,
				// GTypeInstance populated when the function is called
			},
		})
		ifaceInfo := C.GInterfaceInfo{
			interface_data:     (C.gpointer)(unsafe.Pointer(gofuncPtr)),
			interface_finalize: nil,
			interface_init:     C.GInterfaceInitFunc(C.cgoInterfaceInit),
		}
		C.g_type_add_interface_static(
			(C.GType)(gtype),
			(C.GType)(iface.Type()),
			&ifaceInfo,
		)
	}

	registeredTypes[reflect.TypeOf(elem).String()] = Type(gtype)
	return Type(gtype)
}

// privateFromObj returns the actual value of the address we stored in the object's private data.
func privateFromObj(obj unsafe.Pointer) unsafe.Pointer {
	private := C.g_type_instance_get_private((*C.GTypeInstance)(obj), C.objectGType((*C.GObject)(obj)))
	privAddr := (*unsafe.Pointer)(unsafe.Pointer(private))
	return *privAddr
}

// Extendable is an interface implemented by extendable classes. It provides the methods necessary to setup
// the vmethods on the object it represents. When the object is actually used, a pointer to the Go object
// can be retrieved from the C object with FromObjectUnsafePrivate.
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
// that all other implementations should derive from.
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
