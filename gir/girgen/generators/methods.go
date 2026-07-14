package generators

import "github.com/go-gst/go-glib/gir/girgen/file"

type MethodGeneratorList []MethodGenerator

type MethodGenerator interface {
	Generator

	// GenerateInterfaceSignature generates the interface signature for the method.
	// this is needed because the snytax differs from the normal method signatures and
	// the instance parameter is implicit.
	GenerateInterfaceSignature(w file.File)
}

var _ Generator = (MethodGeneratorList)(nil)

// Generate implements Generator.
func (list MethodGeneratorList) Generate(w *file.Package) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

// GenerateInterfaceSignatures iterates over the list of MethodGenerators and calls GenerateInterfaceSignature on each one.
func (list MethodGeneratorList) GenerateInterfaceSignatures(w file.File) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.GenerateInterfaceSignature(w)
	}
}

type VirtualMethodGeneratorList []*VirtualMethodGenerator

// overridesInstanceGenericType is the generic type in the overrides struct that will be used as
// the instance type.
var overridesInstanceGenericType = "Instance"

var _ Generator = (VirtualMethodGeneratorList)(nil)

// Generate implements Generator.
func (list VirtualMethodGeneratorList) Generate(w *file.Package) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

// GenerateClassOverrideFields iterates over the list of VirtualMethodGenerators and calls GenerateClassOverrideField on each one.
func (list VirtualMethodGeneratorList) GenerateClassOverrideFields(w file.File) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.GenerateClassOverrideField(w)
	}
}

func (list VirtualMethodGeneratorList) GenerateInterfaceSignatures(w file.File) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.GenerateInterfaceSignature(w)
	}
}
