package glib

/*
#include "glib.go.h"

GFlagsValue *   getParamSpecFlags     (GParamSpec * p, guint * size)
{
	GParamSpecFlags * pflags = G_PARAM_SPEC_FLAGS (p);
	GFlagsValue * vals = pflags->flags_class->values;
	guint i = 0;
	while (vals[i].value_name) {
    	++i;
    }
	*size = i;
	return vals;
}
*/
import "C"

import (
	"math"
	"unsafe"
)

// ParamSpec is a go representation of a C GParamSpec
type ParamSpec struct{ paramSpec *C.GParamSpec }

// ToParamSpec wraps the given pointer in a ParamSpec instance.
func ToParamSpec(paramspec unsafe.Pointer) *ParamSpec {
	return &ParamSpec{
		paramSpec: (*C.GParamSpec)(paramspec),
	}
}

// Name returns the name of this parameter.
func (p *ParamSpec) Name() string {
	return C.GoString(C.g_param_spec_get_name(p.paramSpec))
}

// Blurb returns the blurb for this parameter.
func (p *ParamSpec) Blurb() string {
	return C.GoString(C.g_param_spec_get_blurb(p.paramSpec))
}

// Flags returns the flags for this parameter.
func (p *ParamSpec) Flags() ParameterFlags {
	return ParameterFlags(p.paramSpec.flags)
}

// ValueType returns the GType for the value inside this parameter.
func (p *ParamSpec) ValueType() Type {
	return Type(p.paramSpec.value_type)
}

// OwnerType returns the Gtype for the owner of this parameter.
func (p *ParamSpec) OwnerType() Type {
	return Type(p.paramSpec.owner_type)
}

// Unref the underlying paramater spec.
func (p *ParamSpec) Unref() { C.g_param_spec_unref(p.paramSpec) }

// FlagsValue is a go representation of GFlagsValue
type FlagsValue struct {
	Value                int
	ValueName, ValueNick string
}

// GetFlagValues returns the possible flags for this parameter.
func (p *ParamSpec) GetFlagValues() []*FlagsValue {
	var gSize C.guint
	gFlags := C.getParamSpecFlags(p.paramSpec, &gSize)
	size := int(gSize)
	out := make([]*FlagsValue, size)

	for idx, flag := range (*[(math.MaxInt32 - 1) / unsafe.Sizeof(C.GFlagsValue{})]C.GFlagsValue)(unsafe.Pointer(gFlags))[:size:size] {
		out[idx] = &FlagsValue{
			Value:     int(flag.value),
			ValueNick: C.GoString(flag.value_nick),
			ValueName: C.GoString(flag.value_name),
		}
	}
	return out
}

// ParameterFlags is a go cast of GParamFlags.
type ParameterFlags int

// Has returns true if these flags contain the provided ones.
func (p ParameterFlags) Has(b ParameterFlags) bool { return p&b != 0 }

// Type casting of GParamFlags
const (
	ParameterReadable       ParameterFlags = C.G_PARAM_READABLE // the parameter is readable
	ParameterWritable                      = C.G_PARAM_WRITABLE // the parameter is writable
	ParameterReadWrite                     = ParameterReadable | ParameterWritable
	ParameterConstruct                     = C.G_PARAM_CONSTRUCT       // the parameter will be set upon object construction
	ParameterConstructOnly                 = C.G_PARAM_CONSTRUCT_ONLY  // the parameter can only be set upon object construction
	ParameterLaxValidation                 = C.G_PARAM_LAX_VALIDATION  // upon parameter conversion (see g_param_value_convert()) strict validation is not required
	ParameterStaticName                    = C.G_PARAM_STATIC_NAME     // the string used as name when constructing the parameter is guaranteed to remain valid and unmodified for the lifetime of the parameter. Since 2.8
	ParameterStaticNick                    = C.G_PARAM_STATIC_NICK     // the string used as nick when constructing the parameter is guaranteed to remain valid and unmmodified for the lifetime of the parameter. Since 2.8
	ParameterStaticBlurb                   = C.G_PARAM_STATIC_BLURB    // the string used as blurb when constructing the parameter is guaranteed to remain valid and unmodified for the lifetime of the parameter. Since 2.8
	ParameterExplicitNotify                = C.G_PARAM_EXPLICIT_NOTIFY // calls to g_object_set_property() for this property will not automatically result in a "notify" signal being emitted: the implementation must call g_object_notify() themselves in case the property actually changes. Since: 2.42.
	ParameterDeprecated                    = C.G_PARAM_DEPRECATED      // the parameter is deprecated and will be removed in a future version. A warning will be generated if it is used while running with G_ENABLE_DIAGNOSTIC=1. Since 2.26
)
