package typesystem

import (
	"slices"
	"strings"

	"github.com/go-gst/go-glib/gir"
)

func CountCTypePointers(ctype string) int {
	return strings.Count(ctype, "*")
}

func GetPointers(count int) string {
	return strings.Repeat("*", count)
}

func trimCTypePointers(ctype string) string {
	return strings.ReplaceAll(ctype, "*", "")
}

// decreaseAnyTypePointers returns a copy of the given AnyType with one less pointer.
func decreaseAnyTypePointers(typ gir.AnyType) (gir.AnyType, bool) {
	var res gir.AnyType

	if typ.Type != nil {
		copy := *typ.Type
		res.Type = &copy

		pointers := CountCTypePointers(copy.CType)

		if pointers == 0 {
			return gir.AnyType{}, false
		}

		copy.CType = trimCTypePointers(copy.CType) + GetPointers(pointers-1)
	}

	if typ.Array != nil {
		copy := *typ.Array
		res.Array = &copy

		pointers := CountCTypePointers(copy.CType)

		if pointers == 0 {
			return gir.AnyType{}, false
		}

		copy.CType = trimCTypePointers(copy.CType) + GetPointers(pointers-1)
	}

	return res, true
}

var pointerTypes = []Type{
	Gpointer,
	Gconstpointer,
}

func isPointer(typ Type) bool {
	return slices.Contains(pointerTypes, typ)
}
