package gobject

import (
	"errors"
	"log"
	"sync"
	"unsafe"
)

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <stdlib.h>
// #include <glib-object.h>
// #include <glib.h>
//
// static GType _g_value_fundamental(GType type) {
//   return (G_TYPE_FUNDAMENTAL(type));
// }
// static gboolean _g_type_is_value(GType g_type) {
//   return (G_TYPE_IS_VALUE(g_type));
// }
import "C"

type Type uint64

const (
	TypeInvalid   Type = C.G_TYPE_INVALID
	TypeNone      Type = C.G_TYPE_NONE
	TypeInterface Type = C.G_TYPE_INTERFACE
	TypeChar      Type = C.G_TYPE_CHAR
	TypeUchar     Type = C.G_TYPE_UCHAR
	TypeBoolean   Type = C.G_TYPE_BOOLEAN
	TypeInt       Type = C.G_TYPE_INT
	TypeUint      Type = C.G_TYPE_UINT
	TypeLong      Type = C.G_TYPE_LONG
	TypeUlong     Type = C.G_TYPE_ULONG
	TypeInt64     Type = C.G_TYPE_INT64
	TypeUint64    Type = C.G_TYPE_UINT64
	TypeEnum      Type = C.G_TYPE_ENUM
	TypeBitflags  Type = C.G_TYPE_FLAGS // renamed bacause it collides with the TypeFlags
	TypeFloat     Type = C.G_TYPE_FLOAT
	TypeDouble    Type = C.G_TYPE_DOUBLE
	TypeString    Type = C.G_TYPE_STRING
	TypePointer   Type = C.G_TYPE_POINTER
	TypeBoxed     Type = C.G_TYPE_BOXED
	TypeParam     Type = C.G_TYPE_PARAM
	TypeVariant   Type = C.G_TYPE_VARIANT
)

// FundamentalType returns the fundamental type of the given actual type.
func FundamentalType(actual Type) Type {
	return Type(C._g_value_fundamental(C.GType(actual)))
}

// TypeIsValue checks whether the passed in type can be used for g_value_init().
func TypeIsValue(t Type) bool {
	return C._g_type_is_value(C.GType(t)) != 0
}

// Name is a wrapper around g_type_name().
func (t Type) Name() string {
	return C.GoString((*C.char)(C.g_type_name(C.GType(t))))
}

// String calls t.Name(). It satisfies fmt.Stringer.
func (t Type) String() string {
	return t.Name()
}

// Depth is a wrapper around g_type_depth().
func (t Type) Depth() uint32 {
	return uint32(C.g_type_depth(C.GType(t)))
}

// Parent is a wrapper around g_type_parent().
func (t Type) Parent() Type {
	return Type(C.g_type_parent(C.GType(t)))
}

// Interfaces returns the interfaces of the given type.
func (t Type) Interfaces() []Type {
	ifaces := t.interfaces()
	if len(ifaces) > 0 {
		defer C.free(unsafe.Pointer(&ifaces[0]))
		return append([]Type(nil), ifaces...)
	}

	return nil
}

func (t Type) interfaces() []Type {
	var n C.guint
	c := C.g_type_interfaces(C.GType(t), &n)

	if n > 0 {
		return unsafe.Slice((*Type)(unsafe.Pointer(c)), n)
	}

	C.free(unsafe.Pointer(c))
	return nil
}

// IsA is a wrapper around g_type_is_a().
func (t Type) IsA(isAType Type) bool {
	return C.g_type_is_a(C.GType(t), C.GType(isAType)) != 0
}

// GValueMarshaler is a marshal function to convert a GValue into an
// appropriate Go type.  The uintptr parameter is a *C.GValue.
type GValueMarshaler func(unsafe.Pointer) (interface{}, error)

// TypeMarshaler represents an actual type and it's associated marshaler.
type TypeMarshaler struct {
	T Type
	F GValueMarshaler
}

type marshalMap sync.Map

var marshalers = new(marshalMap)

// RegisterGValueMarshaler registers a single GValue marshaler. If the function
// has already been called before on the same Type, then it does nothing, and
// the new function is ignored.
func RegisterGValueMarshaler(t Type, f GValueMarshaler) {
	(*sync.Map)(marshalers).LoadOrStore(t, f)
}

// RegisterGValueMarshalers adds marshalers for several types to the internal
// marshalers map. Once registered, calling GoValue on any Value with a
// registered type will return the data returned by the marshaler.
func RegisterGValueMarshalers(marshalers []TypeMarshaler) {
	for _, m := range marshalers {
		RegisterGValueMarshaler(m.T, m.F)
	}
}

func init() {
	RegisterGValueMarshaler(TypeInvalid, marshalInvalid)
	RegisterGValueMarshaler(TypeNone, marshalNone)
	RegisterGValueMarshaler(TypeInterface, marshalInterface)
	RegisterGValueMarshaler(TypeChar, marshalChar)
	RegisterGValueMarshaler(TypeUchar, marshalUchar)
	RegisterGValueMarshaler(TypeBoolean, marshalBoolean)
	RegisterGValueMarshaler(TypeInt, marshalInt32)
	RegisterGValueMarshaler(TypeLong, marshalLong)
	RegisterGValueMarshaler(TypeInt64, marshalInt64)
	RegisterGValueMarshaler(TypeUint, marshalUint32)
	RegisterGValueMarshaler(TypeUlong, marshalUlong)
	RegisterGValueMarshaler(TypeUint64, marshalUint64)
	RegisterGValueMarshaler(TypeFloat, marshalFloat)
	RegisterGValueMarshaler(TypeDouble, marshalDouble)
	RegisterGValueMarshaler(TypeString, marshalString)
	RegisterGValueMarshaler(TypePointer, marshalPointer)
	RegisterGValueMarshaler(TypeBoxed, marshalBoxed)
	// RegisterGValueMarshaler(TypeVariant, marshalVariant)
	RegisterGValueMarshaler(Type(C.g_value_get_type()), marshalValue)

	// included for completeness, each Bitflag/Enum type should implement it's own marshaller
	RegisterGValueMarshaler(TypeBitflags, marshalFlags)
	RegisterGValueMarshaler(TypeEnum, marshalEnum)
}

// lookup returns the closest available GValueMarshaler for the given value's
// type.
func (m *marshalMap) lookup(v *Value) GValueMarshaler {
	typ := v.Type()

	// Check the inheritance tree for concrete classes up until TypeObject.
	for t := typ; t != 0 && t != TypeObject; t = t.Parent() {
		f, ok := m.lookupType(t)
		if ok {
			return f
		}
	}

	// Check the tree again for interfaces.
	for t := typ; t != 0; t = t.Parent() {
		if f := m.lookupIfaces(t); f != nil {
			return f
		}
	}

	fundamental := FundamentalType(typ)
	if f, ok := m.lookupType(fundamental); ok {
		return f
	}

	log.Printf("goglib: missing marshaler for type %q (i.e. %q)", v.Type(), fundamental)
	return nil
}

// lookupWalk is like lookup, except the function walks the user through every
// single possible marshaler until it returns true.
func (m *marshalMap) lookupWalk(v *Value, testFn func(GValueMarshaler) bool) bool {
	typ := v.Type()

	// Check the inheritance tree for concrete classes.
	for t := typ; t != 0; t = t.Parent() {
		f, ok := m.lookupType(t)
		if ok {
			if testFn(f) {
				return true
			}
		}
	}

	// Check the tree again for interfaces.
	for t := typ; t != 0; t = t.Parent() {
		if f := m.lookupIfaces(t); f != nil {
			if testFn(f) {
				return true
			}
		}
	}

	fundamental := FundamentalType(typ)
	if f, ok := m.lookupType(fundamental); ok {
		if testFn(f) {
			return true
		}
	}

	log.Printf("goglib: missing marshaler for type %q (i.e. %q)", v.Type(), fundamental)
	return false
}

func (m *marshalMap) lookupIfaces(t Type) GValueMarshaler {
	ifaces := t.interfaces()
	if len(ifaces) > 0 {
		defer C.free(unsafe.Pointer(&ifaces[0]))
	}

	for _, t := range ifaces {
		f, ok := m.lookupType(t)
		if ok {
			return f
		}
	}

	return nil
}

func (m *marshalMap) lookupType(t Type) (GValueMarshaler, bool) {
	v, ok := (*sync.Map)(m).Load(t)
	if ok {
		return v.(GValueMarshaler), true
	}
	return nil, false
}

func marshalInvalid(unsafe.Pointer) (interface{}, error) {
	return nil, errors.New("invalid type")
}

func marshalNone(unsafe.Pointer) (interface{}, error) {
	return nil, nil
}

func marshalInterface(unsafe.Pointer) (interface{}, error) {
	return nil, errors.New("interface conversion not yet implemented")
}

func marshalChar(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_schar((*C.GValue)(unsafe.Pointer(p)))
	return int8(c), nil
}

func marshalUchar(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_uchar((*C.GValue)(unsafe.Pointer(p)))
	return uint8(c), nil
}

func marshalBoolean(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_boolean((*C.GValue)(unsafe.Pointer(p)))
	return c != 0, nil
}

func marshalInt32(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_int((*C.GValue)(unsafe.Pointer(p)))
	return int32(c), nil
}

func marshalLong(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_long((*C.GValue)(unsafe.Pointer(p)))
	return int64(c), nil
}

func marshalEnum(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return int32(c), nil
}

func marshalInt64(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_int64((*C.GValue)(unsafe.Pointer(p)))
	return int64(c), nil
}

func marshalUint32(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_uint((*C.GValue)(unsafe.Pointer(p)))
	return uint32(c), nil
}

func marshalUlong(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_ulong((*C.GValue)(unsafe.Pointer(p)))
	return uint64(c), nil
}

func marshalFlags(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_flags((*C.GValue)(unsafe.Pointer(p)))
	return uint32(c), nil
}

func marshalUint64(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_uint64((*C.GValue)(unsafe.Pointer(p)))
	return uint64(c), nil
}

func marshalFloat(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_float((*C.GValue)(unsafe.Pointer(p)))
	return float32(c), nil
}

func marshalDouble(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_double((*C.GValue)(unsafe.Pointer(p)))
	return float64(c), nil
}

func marshalString(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_string((*C.GValue)(unsafe.Pointer(p)))
	return C.GoString((*C.char)(c)), nil
}

func marshalBoxed(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	return unsafe.Pointer(c), nil
}

func marshalPointer(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_pointer((*C.GValue)(unsafe.Pointer(p)))
	return unsafe.Pointer(c), nil
}

// func marshalVariant(p unsafe.Pointer) (interface{}, error) {
// 	c := C.g_value_get_variant((*C.GValue)(unsafe.Pointer(p)))
// 	return newVariant((*C.GVariant)(c)), nil
// }
