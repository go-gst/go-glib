package typesystem

import "slices"

type Type interface {
	GIRName() string
	GoType(pointers int) string
	GoTypeRequiredImport() (alias string, module string)
	CGoType(pointers int) string
	CType(pointers int) string
}

// BaseType partially implements the [Type] interface
type BaseType struct {
	GirName string
	GoTyp   string
	CGoTyp  string
	CTyp    string

	GoImportAlias string
	GoImport      string
}

// GIRName implements Type.
func (b BaseType) GIRName() string {
	return b.GirName
}

// CGoType implements Type.
func (b BaseType) CGoType(pointers int) string {
	return GetPointers(pointers) + b.CGoTyp
}

// CType implements Type.
func (b BaseType) CType(pointers int) string {
	return b.CTyp + GetPointers(pointers)
}

// GoType implements Type.
func (b BaseType) GoType(pointers int) string {
	if slices.Contains(GoBuiltins, b.GoTyp) {
		return b.GoTyp
	}
	return GetPointers(pointers) + b.GoTyp
}

// GoType implements Type.
func (b BaseType) GoTypeRequiredImport() (string, string) {
	return b.GoImportAlias, b.GoImport
}
