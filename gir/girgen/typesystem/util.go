package typesystem

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/go-gst/go-glib/gir"
)

func goPackageNameRuneAllowed(r rune) bool {
	return unicode.IsLetter(r) ||
		unicode.IsDigit(r)
}

// goPackageName converts a GIR package name to a Go package name. It's only
// tested against a known set of GIR files.
func goPackageName(girPkgName string) string {
	return strings.Map(func(r rune) rune {
		if goPackageNameRuneAllowed(r) {
			return unicode.ToLower(r)
		}

		return -1
	}, girPkgName)
}

// this is an invalid type indentifier, that can be used for types that do not have a
// c or go type, e.g. callbacks. This is chosen becaus it will also break go/cgo compilation when used
const typeInvalid = "// invalid type"

func cleanCType(ctype string) string {
	// use spaces to prevent valid infix replaces

	ctype = strings.ReplaceAll(ctype, "const ", "")
	ctype = strings.ReplaceAll(ctype, " const", "")

	ctype = strings.ReplaceAll(ctype, "volatile ", "")
	ctype = strings.ReplaceAll(ctype, " volatile", "")

	return ctype
}

func CTypeFromAnytype(t gir.AnyType) string {
	switch {
	case t.Array != nil:
		return t.Array.CType
	case t.Type != nil:
		return t.Type.CType
	default:
		panic("invalid anytype")
	}
}

// cgoPrimitiveTypes contains edge cases for referencing C primitive types from
// CGo.
//
// See https://gist.github.com/zchee/b9c99695463d8902cd33.
var cgoPrimitiveTypes = map[string]string{
	"long long": "longlong",

	"unsigned char":      "uchar",
	"unsigned int":       "uint",
	"unsigned short":     "ushort",
	"unsigned long":      "ulong",
	"unsigned long long": "ulonglong",
}

func CtypeToCgoType(ctype string) string {
	if ctype == "void*" {
		// FIXME: this would need a go import, but we practically always
		// import unsafe already.
		return "unsafe.Pointer"
	}

	pointers := CountCTypePointers(ctype)

	base := trimCTypePointers(cleanCType(ctype))

	if replace, ok := cgoPrimitiveTypes[base]; ok {
		base = replace
	}

	return GetPointers(pointers) + "C." + base
}

func debugCTypeFromAnytype(t gir.AnyType) string {
	switch {
	case t.Array != nil:
		return "(array) " + t.Array.CType
	case t.Type != nil:
		return t.Type.CType
	default:
		panic("invalid anytype")
	}
}

func infoFromAnyGir(girAny any) (string, gir.InfoAttrs, gir.InfoElements) {
	var attrs gir.InfoAttrs
	var elements gir.InfoElements

	if t, ok := girAny.(girWithInfoAttrs); ok {
		attrs = t.GetInfoAttrs()
	}
	if t, ok := girAny.(girWithInfoElements); ok {
		elements = t.GetInfoElements()
	}

	switch t := girAny.(type) {
	case *gir.Class:
		return t.Name, attrs, elements
	case *gir.Interface:
		return t.Name, attrs, elements
	case *gir.Callback:
		return t.Name, attrs, elements
	case *gir.Field:
		return t.Name, attrs, elements
	case *gir.Enum:
		return t.Name, attrs, elements
	case *gir.Bitfield:
		return t.Name, attrs, elements
	case *gir.Member:
		return t.Name(), attrs, elements
	case *gir.Record:
		return t.Name, attrs, elements
	case *gir.Constant:
		return t.Name, attrs, elements
	case *gir.Constructor:
		return t.Name, attrs, elements
	case *gir.CallableAttrs:
		return t.Name, attrs, elements
	case *gir.Method:
		return t.Name, attrs, elements
	case *gir.VirtualMethod:
		return t.Name, attrs, elements
	case *gir.Union:
		return t.Name, attrs, elements
	case *gir.Alias:
		return t.Name, attrs, elements
	case *gir.Function:
		return t.Name, attrs, elements
	case *gir.Signal:
		return t.Name, attrs, elements
	default:
		panic(fmt.Sprintf("received unhandled type: %T", t))
	}
}
