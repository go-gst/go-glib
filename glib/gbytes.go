package glib

// #include "glib.go.h"
import "C"
import (
	"runtime"
	"unsafe"
)

func init() {

	tm := []TypeMarshaler{
		{Type(C.g_bytes_get_type()), marshalBytes},
	}

	RegisterGValueMarshalers(tm)
}

type Bytes struct {
	ptr *C.GBytes
}

func wrapBytes(cbytes *C.GBytes) *Bytes {
	bytes := &Bytes{
		ptr: cbytes,
	}

	runtime.SetFinalizer(bytes, func(b *Bytes) {
		b.Unref()
	})

	return bytes
}

// NewBytes copies the passed data and creates a new GBytes
func NewBytes(data []byte) *Bytes {
	addr := unsafe.SliceData(data)

	cbytes := C.g_bytes_new(C.gconstpointer(addr), C.gsize(len(data)))

	return wrapBytes(cbytes)
}

// Data copies the data from the GBytes and returns them
func (b *Bytes) Data() []byte {
	var len C.gsize = 0
	addr := C.g_bytes_get_data(b.ptr, &len)

	return C.GoBytes(unsafe.Pointer(addr), C.int(len))
}

func (b *Bytes) Ref() {
	C.g_bytes_ref(b.ptr)
}

func (b *Bytes) Unref() {
	C.g_bytes_unref(b.ptr)
}

func (b *Bytes) ToGValue() (*Value, error) {
	val, err := ValueInit(Type(C.g_bytes_get_type()))
	if err != nil {
		return nil, err
	}
	val.SetBoxed(unsafe.Pointer(b.ptr))
	return val, nil
}

func marshalBytes(p unsafe.Pointer) (interface{}, error) {
	cbytes := C.g_value_get_boxed((*C.GValue)(p))
	b := wrapBytes((*C.GBytes)(cbytes))

	b.Ref()

	return b, nil
}
