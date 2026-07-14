package gobject

import (
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
import "C"

func init() {
	RegisterGValueMarshalers([]TypeMarshaler{
		TypeMarshaler{T: TypeParam, F: marshalParamSpec},
	})
}

func marshalParamSpec(p unsafe.Pointer) (any, error) {
	native := ValueFromNative(p).Param()

	return UnsafeParamSpecFromGlibNone(native), nil
}

// ParamSpec is a go representation of a C GParamSpec
type ParamSpec struct{ *paramSpec }

// paramSpec is the struct that is finalized
type paramSpec struct {
	native *C.GParamSpec
}

func UnsafeParamSpecFromGlibBorrow(paramspec unsafe.Pointer) *ParamSpec {
	return &ParamSpec{
		paramSpec: &paramSpec{(*C.GParamSpec)(paramspec)},
	}
}

func UnsafeParamSpecFromGlibFull(p unsafe.Pointer) *ParamSpec {
	pspec := UnsafeParamSpecFromGlibBorrow(p)

	runtime.SetFinalizer(pspec.paramSpec, func(p *paramSpec) {
		C.g_param_spec_unref(p.native)
	})

	return pspec
}

func UnsafeParamSpecFromGlibNone(p unsafe.Pointer) *ParamSpec {
	pspec := UnsafeParamSpecFromGlibBorrow(p)

	C.g_param_spec_ref(pspec.native)

	return pspec
}

func UnsafeParamSpecToGlibFull(p *ParamSpec) unsafe.Pointer {
	runtime.SetFinalizer(p.paramSpec, nil)

	return unsafe.Pointer(p.paramSpec.native)
}

func UnsafeParamSpecToGlibNone(p *ParamSpec) unsafe.Pointer {
	return unsafe.Pointer(p.paramSpec.native)
}

// Name returns the name of this parameter.
func (p *ParamSpec) Name() string {
	return C.GoString(C.g_param_spec_get_name(p.native))
}

// Blurb returns the blurb for this parameter.
func (p *ParamSpec) Blurb() string {
	return C.GoString(C.g_param_spec_get_blurb(p.native))
}

// Flags returns the flags for this parameter.
func (p *ParamSpec) Flags() ParamFlags {
	return ParamFlags(p.native.flags)
}

// ValueType returns the GType for the value inside this parameter.
func (p *ParamSpec) ValueType() Type {
	return Type(p.native.value_type)
}

// OwnerType returns the Gtype for the owner of this parameter.
func (p *ParamSpec) OwnerType() Type {
	return Type(p.native.owner_type)
}

// UnsafeRef adds a reference to the parameter spec. This will leak if not paired with
// a call to UnsafeUnref.
func (p *ParamSpec) UnsafeRef() {
	C.g_param_spec_ref(p.native)
}

// UnsafeUnref removes a reference to the parameter spec. This should be called
// when the parameter spec is no longer needed. This will not free the parameter
// spec, but will decrement the reference count. If the reference count reaches
// zero, the parameter spec will be freed.
func (p *ParamSpec) UnsafeUnref() {
	C.g_param_spec_unref(p.native)
}
