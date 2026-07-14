package generators

import "github.com/go-gst/go-glib/gir/girgen/file"

type Generator interface {
	Generate(*file.Package)
}

type SubGenerator interface {
	Generate(file.CodeWriter)
}

type GeneratorList []Generator

var _ Generator = GeneratorList{}

func (list GeneratorList) Generate(w *file.Package) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

type NoopGenerator struct{}

func (list NoopGenerator) Generate(w *file.Package) {}

func GenerateAll(w *file.Package, gens ...Generator) {
	GeneratorList(gens).Generate(w)
}
