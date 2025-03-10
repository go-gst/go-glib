package glib

//#include "glib.go.h"
import "C"

import (
	"unsafe"
)

// ParamSpecUInt is a go representation of a C ParamSpecUInt
type ParamSpecUInt struct{ paramSpecUInt *C.GParamSpecUInt }

// ToParamSpec wraps the given pointer in a ParamSpec instance.
func ToParamSpecUInt(paramspecuint unsafe.Pointer) *ParamSpecUInt {
	return &ParamSpecUInt{
		paramSpecUInt: (*C.GParamSpecUInt)(paramspecuint),
	}
}

// Unref the underlying paramater spec.
func (p *ParamSpecUInt) Unref() {
	C.g_param_spec_unref((*C.GParamSpec)((unsafe.Pointer)(p.paramSpecUInt)))
}

// Minimum returns the minimum value of this parameter.
func (p *ParamSpecUInt) Minimum() uint {
	return (uint)(p.paramSpecUInt.minimum)
}

// Maximum returns the maximum value of this parameter.
func (p *ParamSpecUInt) Maximum() uint {
	return (uint)(p.paramSpecUInt.maximum)
}

// DefaultValue returns the default value of this parameter.
func (p *ParamSpecUInt) DefaultValue() uint {
	return (uint)(p.paramSpecUInt.default_value)
}

// ParamSpecUInt64 is a go representation of a C ParamSpecUInt64
type ParamSpecUInt64 struct{ paramSpecUInt64 *C.GParamSpecUInt64 }

// ToParamSpec64 wraps the given pointer in a ParamSpecUInt64 instance.
func ToParamSpecUInt64(paramspecuint64 unsafe.Pointer) *ParamSpecUInt64 {
	return &ParamSpecUInt64{
		paramSpecUInt64: (*C.GParamSpecUInt64)(paramspecuint64),
	}
}

// Unref the underlying paramater spec.
func (p *ParamSpecUInt64) Unref() {
	C.g_param_spec_unref((*C.GParamSpec)((unsafe.Pointer)(p.paramSpecUInt64)))
}

// Minimum returns the minimum value of this parameter.
func (p *ParamSpecUInt64) Minimum() uint {
	return (uint)(p.paramSpecUInt64.minimum)
}

// Maximum returns the maximum value of this parameter.
func (p *ParamSpecUInt64) Maximum() uint {
	return (uint)(p.paramSpecUInt64.maximum)
}

// DefaultValue returns the default value of this parameter.
func (p *ParamSpecUInt64) DefaultValue() uint {
	return (uint)(p.paramSpecUInt64.default_value)
}
