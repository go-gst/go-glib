package glib

// #include "glib.go.h"
import "C"
import "unsafe"

// NewStringParam returns a new ParamSpec that will hold a string value.
func NewStringParam(name, nick, blurb string, defaultValue *string, flags ParameterFlags) *ParamSpec {
	var cdefault *C.gchar
	if defaultValue != nil {
		cdefault = C.CString(*defaultValue)
		defer C.free(unsafe.Pointer(cdefault))
	}

	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_string(
		cname,
		cnick,
		cblurb,
		cdefault,
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewBoolParam creates a new ParamSpec that will hold a boolean value.
func NewBoolParam(name, nick, blurb string, defaultValue bool, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_boolean(
		cname,
		cnick,
		cblurb,
		gbool(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewIntParam creates a new ParamSpec that will hold a signed integer value.
func NewIntParam(name, nick, blurb string, min, max, defaultValue int, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_int(
		cname,
		cnick,
		cblurb,
		C.gint(min),
		C.gint(max),
		C.gint(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewUintParam creates a new ParamSpec that will hold an unsigned integer value.
func NewUintParam(name, nick, blurb string, min, max, defaultValue uint, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_uint(
		cname,
		cnick,
		cblurb,
		C.guint(min),
		C.guint(max),
		C.guint(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewInt64Param creates a new ParamSpec that will hold a signed 64-bit integer value.
func NewInt64Param(name, nick, blurb string, min, max, defaultValue int64, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_int64(
		cname,
		cnick,
		cblurb,
		C.gint64(min),
		C.gint64(max),
		C.gint64(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewUint64Param creates a new ParamSpec that will hold an unsigned 64-bit integer value.
func NewUint64Param(name, nick, blurb string, min, max, defaultValue uint64, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_uint64(
		cname,
		cnick,
		cblurb,
		C.guint64(min),
		C.guint64(max),
		C.guint64(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewFloat32Param creates a new ParamSpec that will hold a 32-bit float value.
func NewFloat32Param(name, nick, blurb string, min, max, defaultValue float32, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_float(
		cname,
		cnick,
		cblurb,
		C.gfloat(min),
		C.gfloat(max),
		C.gfloat(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewFloat64Param creates a new ParamSpec that will hold a 64-bit float value.
func NewFloat64Param(name, nick, blurb string, min, max, defaultValue float64, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_double(
		cname,
		cnick,
		cblurb,
		C.gdouble(min),
		C.gdouble(max),
		C.gdouble(defaultValue),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}

// NewBoxedParam creates a new ParamSpec containing a boxed type.
func NewBoxedParam(name, nick, blurb string, boxedType Type, flags ParameterFlags) *ParamSpec {
	cname := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cname))
	cnick := (*C.gchar)(C.CString(nick))
	defer C.free(unsafe.Pointer(cnick))
	cblurb := (*C.gchar)(C.CString(blurb))
	defer C.free(unsafe.Pointer(cblurb))

	paramSpec := C.g_param_spec_boxed(
		cname,
		cnick,
		cblurb,
		C.GType(boxedType),
		C.GParamFlags(flags),
	)
	return &ParamSpec{paramSpec: paramSpec}
}
