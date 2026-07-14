package gobject

// #include <glib-object.h>
// extern void _goglibInterfaceInit(gpointer instance, gpointer ifaceData);
// extern void _goglibClassInit(gpointer gclass, gpointer classData);
// extern void _goglibInstanceInit(GTypeInstance* instance, gpointer g_class);
// GType typeFromGObjectClass (GObjectClass *c) { return (G_OBJECT_CLASS_TYPE(c)); };
import "C"
import (
	"log"
	"reflect"
	"unsafe"

	"github.com/go-gst/go-glib/pkg/core/userdata"
)

type subClassData struct {
	classInit    func(gclass unsafe.Pointer)
	instanceInit func(instance unsafe.Pointer)
}

// UnsafeRegisterSubClass registers a new subclass of the given parentGtype. This is wrapped by the generated bindings
// for ease of use.
// This aligns with https://gjs.guide/guides/gobject/subclassing.html
func UnsafeRegisterSubClass[InstanceT Object, ClassT any, OverridesT ObjectOverrider[InstanceT]](
	// user supplied arguments:
	name string,
	classInit func(class ClassT),
	constructor func() InstanceT,
	overrides OverridesT,
	signals map[string]SignalDefinition,
	// parent class depending arguments, these can be autofilled by the generated bindings:
	parentGtype Type,
	parentClassFromUnsafePointer func(unsafe.Pointer) ClassT,
	parentApplyOverridesFunc func(gclass unsafe.Pointer, overrides OverridesT),
	parentWrapObject func(*ObjectInstance) Object,

	// user supplied interfaces:
	interfaceInits ...SubClassInterfaceInit[InstanceT],
) Type {
	instanceType := getInstanceType[InstanceT]()

	if constructor == nil {
		constructor = func() InstanceT {
			return reflect.New(instanceType).Interface().(InstanceT)
		}
	}
	if classInit == nil {
		classInit = func(class ClassT) {}
	}

	var typeQuery C.GTypeQuery
	C.g_type_query(C.GType(parentGtype), &typeQuery)
	if typeQuery._type == 0 {
		log.Panicln("GType", parentGtype, "is is unknown")
	}

	baseOverrides := overrides.getObjectOverrides()

	var data *subClassData
	data = &subClassData{
		classInit: func(gclass unsafe.Pointer) {
			// first override the virtual methods on the class
			parentApplyOverridesFunc(gclass, overrides)

			class := parentClassFromUnsafePointer(gclass)

			// gtype is the type of the instance
			gtype := Type(C.typeFromGObjectClass((*C.GObjectClass)(gclass)))

			// register the signals
			for name, signal := range signals {
				signal.registerFor(name, gtype)
			}

			// save the subClass data with the class pointer
			registerSubclassData(gclass, data)

			// register the private data needed for the instance in instance init
			baseClass := UnsafeObjectClassFromGlibBorrow(gclass)
			baseClass.UnsafeAddPrivateData(unsafe.Sizeof(instanceID))

			// then allow the user to call some methods on the class, e.g. to supply metadata
			classInit(class)
		},
		instanceInit: func(cInstance unsafe.Pointer) {
			obj := wrapObject(cInstance)
			// parent is the pointer to the parent instance
			parent := parentWrapObject(obj)

			instance := constructor()

			// set the parent instance, aka the first embedded field.
			// the embedded field is not the parent interface though, but instead the instance struct
			// so we need to deref the pointer and set the first field

			parentInstance := reflect.ValueOf(parent)

			if parentInstance.Kind() != reflect.Ptr {
				panic("parent instance is not a pointer to a struct")
			}

			parentInstance = parentInstance.Elem()

			if parentInstance.Kind() != reflect.Struct {
				log.Panicln("parent instance is not pointer to a struct")
			}

			instanceValue := reflect.ValueOf(instance).Elem()

			if instanceValue.Kind() == reflect.Ptr {
				instanceValue = instanceValue.Elem()
			}

			if instanceValue.Kind() != reflect.Struct {
				log.Panicln("instance is not a struct")
			}

			// panic on type mismatch of the first field
			if instanceValue.Field(0).Type() != parentInstance.Type() {
				log.Panicf("instance first field is not of the same type as parents instance type. expected %s, got %s\n", parentInstance.Type(), instanceValue.Field(0).Type())
			}

			// panic on initialized first field
			if !instanceValue.Field(0).IsZero() {
				log.Panicln("instance first field is already set")
			}

			instanceValue.Field(0).Set(parentInstance)

			// store the instance in the private data of the instance, so we can retrieve it later
			saveInstanceInPrivateData(instance)

			if baseOverrides.InstanceInit != nil {
				baseOverrides.InstanceInit(instance)
			}
		},
	}

	dataKey := userdata.Register(data)

	// this can be allocated by go because the parameter is owned by the caller
	typeInfo := &C.GTypeInfo{
		// not needed:
		value_table:    nil,
		base_init:      nil,
		base_finalize:  nil,
		n_preallocs:    0,
		class_finalize: nil,

		instance_size: C.guint16(typeQuery.instance_size),
		class_size:    C.guint16(typeQuery.class_size),
		class_init:    C.GClassInitFunc(C._goglibClassInit),
		instance_init: C.GInstanceInitFunc(C._goglibInstanceInit),
		class_data:    C.gconstpointer(dataKey),
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	gtype := C.g_type_register_static(
		C.GType(parentGtype),
		(*C.gchar)(cName),
		typeInfo,
		C.GTypeFlags(0),
	)

	// register the interfaces
	for _, iface := range interfaceInits {
		panic("TODO: interface init is not implemented")

		// TODO: we need to set the interface Instance types in the class init function as well and require that
		// the interfaces are registered in the same order as the user provided embedded interfaces

		ifaceInfo := iface.toInterfaceInfo()
		C.g_type_add_interface_static(gtype, C.GType(iface.InterfaceType), ifaceInfo)
	}

	t := Type(gtype)

	// FIXME: we should register the casting method, but this creates ref counting issues
	// RegisterObjectCasting(
	// 	t,
	// 	func(inst *ObjectInstance) Object {
	// 		return loadInstanceFromPrivateData(inst)
	// 	},
	// )

	return t
}

type SignalDefinition struct {
	Flags SignalFlags
	// ParamTypes is a list of parameter types. The instance parameter can be omitted, as it is
	// automatically added by the bindings.
	ParamTypes []Type
	ReturnType Type

	// Accumulator is the SignalAccumulator. Can be nil.
	Accumulator SignalAccumulator

	// Handler must be a function with the correct signature of the signal. The first (or receiver) parameter
	// must be the instance type
	Handler any
}

func (sd SignalDefinition) registerFor(name string, typ Type) {
	NewSignal(name, typ, sd.Flags, sd.Handler, sd.Accumulator, sd.ParamTypes, sd.ReturnType)
}

type SubClassInterfaceInit[InstanceT Object] struct {
	InterfaceType  Type
	ApplyOverrides func(gclass unsafe.Pointer)
	FromGlib       func(unsafe.Pointer) any
}

// toInterfaceInfo returns the GInterfaceInfo struct for the given interface type.
func (i SubClassInterfaceInit[InstanceT]) toInterfaceInfo() *C.GInterfaceInfo {
	applyOverridesData := userdata.Register(i.ApplyOverrides)

	return &C.GInterfaceInfo{
		interface_init:     C.GInterfaceInitFunc(C._goglibInterfaceInit),
		interface_finalize: nil,
		interface_data:     C.gpointer(applyOverridesData),
	}
}

func getInstanceType[T any]() reflect.Type {
	var zero T

	typ := reflect.TypeOf(zero)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return typ
}
