package typesystem

import (
	"fmt"

	"github.com/go-gst/go-glib/gir"
)

// Array is the type for array params. It has an inner type and may reference another [Param] for its length.
type Array struct {
	GirName         string
	CTypeOverride   string
	CGoTypeOverride string
	GoTypeOverride  string

	Inner         CouldBeForeign[Type]
	InnerPointers int

	Length         *Param // length is filled out by [NewParameters]
	ZeroTerminated bool
	FixedSize      int
}

// GoTypeRequiredImport implements Type.
func (a *Array) GoTypeRequiredImport() (alias string, module string) {
	if a.Inner.Type == nil {
		return "", ""
	}
	return a.Inner.Type.GoTypeRequiredImport()
}

var _ Type = (*Array)(nil)
var _ minPointerConstrainedType = (*Array)(nil)
var _ maxPointerConstrainedType = (*Array)(nil)

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Array) maxPointersAllowed() int {
	return a.InnerPointers + 1
}

// minPointersRequired implements minPointerConstrainedType.
func (a *Array) minPointersRequired() int {
	return a.InnerPointers + 1
}

func (a *Array) CGoType(pointers int) string {
	if a.CGoTypeOverride != "" {
		return a.CGoTypeOverride
	}
	return "*" + a.Inner.Type.CGoType(a.InnerPointers)
}

// CType implements Type.
func (a *Array) CType(pointers int) string {
	if a.CTypeOverride != "" {
		return a.CTypeOverride
	}
	return a.Inner.Type.CType(a.InnerPointers) + "*"
}

// GoType implements Type.
func (a *Array) GoType(pointers int) string {
	if a.GoTypeOverride != "" {
		return a.GoTypeOverride
	}
	inner := a.Inner.NamespacedGoType(a.InnerPointers)

	if a.FixedSize > 0 {
		return fmt.Sprintf("[%d]%s", a.FixedSize, inner)
	}
	return fmt.Sprintf("[]%s", inner)
}

// GIRName implements Type.
func (a *Array) GIRName() string {
	if a.GirName != "" {
		return a.GirName
	}
	if a.Inner.Type == nil {
		return "array[unknown]"
	}
	return fmt.Sprintf("array[%s]", a.Inner.Type.GIRName())
}

// getArrayType resolves the array type in the current env
func (e *env) getArrayType(arr *gir.Array) *Array {
	if arr.Type == nil {
		e.logger.Warn("array type is nil", "ctype", arr.CType)
		return nil
	}

	if arr.Length == nil && arr.FixedSize == 0 && !arr.IsZeroTerminated() {
		// this is an unbounded array, which requires some unsafe preconditions not
		// documented in GIR, must be handled manually
		e.logger.Warn("unbounded array, not supported", "name", arr.Name)
		return nil
	}

	if arr.CType == "" {
		// this is true for some dummy fields, we just ignore them
		e.logger.Warn("ignoring array with empty ctype")
		return nil
	}

	if arr.CType == "gpointer" || arr.CType == "gconstpointer" {
		// this represents a bytes array, e.g. for g_bytes_get_data
		return &Array{
			GirName:         arr.Name,
			CTypeOverride:   arr.CType,
			CGoTypeOverride: "C." + arr.CType,
			GoTypeOverride:  "[]byte",
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),

			Inner: CouldBeForeign[Type]{}, // no inner type
		}
	}
	if arr.CType == "void*" {
		// this represents a bytes array, e.g. for g_bytes_get_data
		return &Array{
			GirName:         arr.Name,
			CTypeOverride:   arr.CType,
			CGoTypeOverride: "unsafe.Pointer", // FIXME: this needs an import
			GoTypeOverride:  "[]byte",
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),

			Inner: CouldBeForeign[Type]{}, // no inner type
		}
	}

	cleanedCtype := cleanCType(arr.CType)

	if cleanedCtype == "gchar*" {
		// this is a string where the length is somehow given
		return &Array{
			GirName:         arr.Name,
			CTypeOverride:   arr.CType, // may contain "const"
			CGoTypeOverride: "*C.gchar",
			GoTypeOverride:  "string",
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),

			Inner: CouldBeForeign[Type]{}, // no inner type
		}
	}

	if cleanedCtype == "char*" {
		// this is a string where the length is somehow given
		return &Array{
			GirName:         arr.Name,
			CTypeOverride:   arr.CType, // may contain "const"
			CGoTypeOverride: "*C.char",
			GoTypeOverride:  "string",
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),

			Inner: CouldBeForeign[Type]{}, // no inner type
		}
	}

	ns, inner := e.findType(arr.Type)

	innerpointers := CountCTypePointers(arr.Type.CType)

	// if the ctype did not provide any pointers then we override it with the min
	// required pointers of the inner type
	if constrained, ok := inner.(minPointerConstrainedType); ok && innerpointers == 0 {
		innerpointers = constrained.minPointersRequired()
	}

	if inner == nil {
		e.logger.Warn("could not find array inner type", "name", arr.Type.Name)
		return nil
	}

	// if innerpointers+1 != CountCTypePointers(arr.CType) {
	// 	e.logger.Warn("array pointer count does not match inner type", "name", arr.Type.Name, "ctype", arr.CType)
	// }

	cgoType := "C." + trimCTypePointers(cleanCType(arr.CType))
	cGopointers := CountCTypePointers(arr.CType)

	array := &Array{
		GirName:         arr.Name,
		CTypeOverride:   arr.CType,
		CGoTypeOverride: GetPointers(cGopointers) + cgoType,
		InnerPointers:   innerpointers,
		Inner: CouldBeForeign[Type]{
			Namespace: ns,
			Type:      inner,
		},
		Length:         nil, // will be set by params if relevant
		ZeroTerminated: arr.IsZeroTerminated(),
		FixedSize:      arr.FixedSize,
	}

	return array
}
