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
// by Go types that get registered as GObjects. For more information see RegisterGoType.
type GoObjectSubclass interface {
	// New should return a new instantiated GoElement ready to be used.
	New() GoObjectSubclass
	// ClassInit is called on the element after registering it with the type system. This is when the element
	// should install its properties and metadata.
	ClassInit(*ObjectClass)
}

// Initter is an interface that can be implemented on top of a GoObjectSubclass. It provides a method that gets
// called with the GObject instance at instance initialization.
type Initter interface {
	InstanceInit(*Object)
}

// TypeInstance is a loose binding around the glib GTypeInstance. It holds the information required to assign
// various capabilities of a GoObjectSubclass.
type TypeInstance struct {
	// The GType cooresponding to this GoType
	GType Type
	// A pointer to the underlying C instance being instantiated.
	GTypeInstance unsafe.Pointer
	// A representation of the GoType. NOTE: This is the instantiated GoType as passed
	// to RegisterGoType and is not that which is being instantiated. It is safe to use
	// to verify implemented methods, but should not be relied on for executing runtime
	// functionality. It CAN be used in rare cases where methods need to be implemented
	// that don't pass a pointer to the object implementing the method.
	GoType GoObjectSubclass
}

// Interface can be implemented by extending packages. They provide the base type for the interface and
// a function to call during interface_init.
//
// The function is called during class_init and  is passed a TypeInstance populated with the GType
// corresponding to the Go object, a pointer to the underlying C object, and a pointer to a reference
// Go object. When the object is actually used, a pointer to it can be retrieved from the C object with
// FromObjectUnsafePrivate.
//
// The user of the Interface is responsible for implementing the methods required by the interface. The GoType
// provided to the InterfaceInitFunc will be the object that is expected to carry the implementation.
type Interface interface {
	Type() Type
	Init(*TypeInstance)
}

type interfaceData struct {
	iface     Interface
	gtype     Type
	classData *classData
}

// FromObjectUnsafePrivate will return the GoObjectSubclass addressed in the private data of the given GObject.
func FromObjectUnsafePrivate(obj unsafe.Pointer) GoObjectSubclass {
	objPriv := privateFromObj(obj)
	if objPriv == nil {
		return nil
	}
	ptr := gopointer.Restore(objPriv)
	if ptr == nil {
		return nil
	}
	class, ok := ptr.(GoObjectSubclass)
	if !ok {
		return nil
	}
	return class
}

// privateFromObj returns the actual value of the address we stored in the object's private data.
func privateFromObj(obj unsafe.Pointer) unsafe.Pointer {
	private := C.g_type_instance_get_private((*C.GTypeInstance)(obj), C.objectGType((*C.GObject)(obj)))
	if private == nil {
		return nil
	}
	privAddr := (*unsafe.Pointer)(unsafe.Pointer(private))
	if privAddr == nil {
		return nil
	}
	return *privAddr
}

type classData struct {
	elem GoObjectSubclass
	ext  Extendable
}

// RegisterGoType is used to register an interface implemented in the Go runtime with the GType
// system. It takes the name to assign the type, the interface containing the object itself, and
// an Extendable denoting the types it extends. It's the responsibility of packages using these
// bindings to implement Extendables that call up to the ExtendsObject.InitClass included in this
// package during their own implementations. ClassSize and InitClass are ignored in the Extendable
// if the object does not implement a GoObjectSubclass.
//
// Interfaces are optional and flags additional interfaces as implemented on the class. Similar to the
// extendables, libraries using these bindings can implement the Interface interface to provide support
// for other GInterfaces.
func RegisterGoType(name string, goObject interface{}, extends Extendable, interfaces ...Interface) Type {
	registerMutex.Lock()
	defer registerMutex.Unlock()
	if registered, ok := registeredTypes[reflect.TypeOf(goObject).String()]; ok {
		return registered
	}

	typeInfo := (*C.GTypeInfo)(C.malloc(C.sizeof_GTypeInfo))
	defer C.free(unsafe.Pointer(typeInfo))

	typeInfo.base_init = nil
	typeInfo.base_finalize = nil
	typeInfo.class_finalize = nil
	typeInfo.n_preallocs = 0
	typeInfo.value_table = nil

	typeInfo.instance_size = C.gushort(extends.InstanceSize())

	// Register class data if the object implements a GoObjectSubclass
	var cd *classData
	if object, ok := goObject.(GoObjectSubclass); ok {
		cd := &classData{
			elem: object,
			ext:  extends,
		}
		ptr := gopointer.Save(cd)
		typeInfo.class_size = C.gushort(extends.ClassSize())
		typeInfo.class_init = C.GClassInitFunc(C.cgoClassInit)
		typeInfo.instance_init = C.GInstanceInitFunc(C.cgoInstanceInit)
		typeInfo.class_data = (C.gconstpointer)(ptr)
	} else {
		typeInfo.class_init = nil
		typeInfo.instance_init = nil
		typeInfo.class_data = nil
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	gtype := C.g_type_register_static(
		C.GType(extends.Type()),
		(*C.gchar)(cName),
		typeInfo,
		C.GTypeFlags(0),
	)

	// Add interfaces if the go object implements a GoObjectSubclass
	if _, ok := goObject.(GoObjectSubclass); ok {
		for _, iface := range interfaces {
			gofuncPtr := gopointer.Save(&interfaceData{
				iface:     iface,
				gtype:     Type(gtype),
				classData: cd,
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
	}

	registeredTypes[reflect.TypeOf(goObject).String()] = Type(gtype)
	return Type(gtype)
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
