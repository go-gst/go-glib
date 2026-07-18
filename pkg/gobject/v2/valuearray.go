package gobject

// #cgo pkg-config: glib-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <stdlib.h>
// #include <glib-object.h>
// #include <glib.h>
import "C"
import (
	"runtime"
	"unsafe"
)

var TypeValueArray Type = Type(C.g_value_array_get_type())

func init() {
	RegisterGValueMarshaler(TypeValueArray, marshalValueArray)
}

// ValueArray is a generic representation of GValueArray.
//
// See https://docs.gtk.org/gobject/struct.ValueArray.html for more details.
//
// Since GValueArray is deprecated, this type is provided for compatibility purposes only.
//
// A nil ValueArray is marshaled as an empty GValueArray in C.
type ValueArray []any

var _ GoValueInitializer = ValueArray(nil)

// GoValueType implements GoValueInitializer.
func (a ValueArray) GoValueType() Type {
	return TypeValueArray
}

// SetGoValue implements GoValueInitializer.
func (a ValueArray) SetGoValue(v *Value) {
	nativeArray := C.g_value_array_new(C.guint(len(a)))

	for _, elem := range a {
		elemValue := NewValue(elem)
		C.g_value_array_append(nativeArray, elemValue.native())
		runtime.KeepAlive(elemValue)
	}

	C.g_value_take_boxed(v.native(), C.gconstpointer(unsafe.Pointer(nativeArray)))
	runtime.KeepAlive(v)
}

func marshalValueArray(p unsafe.Pointer) (interface{}, error) {
	nativeArray := (*C.GValueArray)(C.g_value_get_boxed((*C.GValue)(p)))

	if nativeArray == nil {
		return ValueArray(nil), nil
	}

	length := int(nativeArray.n_values)

	goArray := make(ValueArray, length)

	valuesArray := unsafe.Slice(nativeArray.values, length)

	for i := range length {
		v := ValueFromNative(unsafe.Pointer(&valuesArray[i]))
		goArray[i] = v.GoValue()
	}

	return goArray, nil
}
