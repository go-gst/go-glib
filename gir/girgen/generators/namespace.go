package generators

import (
	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type NamespaceGenerator struct {
	Namespace *typesystem.Namespace

	// Sub generators:
	SubGenerators GeneratorList
}

// Generate implements Generator.
func (g *NamespaceGenerator) Generate(w *file.Package) {
	w.SetNamespace(g.Namespace)

	// run sub generators
	g.SubGenerators.Generate(w)
}

func NewNamespaceGenerator(
	cfg *Config,
) *NamespaceGenerator {
	ns := cfg.Namespace

	// namespace := ctx.Namespace(ns)
	gen := &NamespaceGenerator{
		Namespace: ns,
	}

	for _, c := range ns.Constants {
		if cgen := NewConstantGenerator(cfg, c); cgen != nil {
			gen.SubGenerators = append(gen.SubGenerators, cgen)
		}
	}

	for _, a := range ns.Aliases {
		if agen := NewAliasGenerator(cfg, a); agen != nil {
			gen.SubGenerators = append(gen.SubGenerators, agen)
		}
	}
	for _, e := range ns.Enums {
		if egen := NewEnumGenerator(cfg, e); egen != nil {
			gen.SubGenerators = append(gen.SubGenerators, egen)
		}
	}
	for _, b := range ns.Bitfields {
		if bgen := NewBitfieldGenerator(cfg, b); bgen != nil {
			gen.SubGenerators = append(gen.SubGenerators, bgen)
		}
	}
	for _, cb := range ns.Callbacks {
		if cbgen := NewCallbackGenerator(cfg, cb); cbgen != nil {
			gen.SubGenerators = append(gen.SubGenerators, cbgen)
		}
	}
	for _, f := range ns.Functions {
		if fgen := NewCallableGenerator(cfg, f); fgen != nil {
			gen.SubGenerators = append(gen.SubGenerators, fgen)
		}
	}
	for _, inter := range ns.Interfaces {
		if intergen := NewInterfaceGenerator(cfg, inter); intergen != nil {
			gen.SubGenerators = append(gen.SubGenerators, intergen)
		}
	}
	for _, class := range ns.Classes {
		if classgen := NewClassGenerator(cfg, class); classgen != nil {
			gen.SubGenerators = append(gen.SubGenerators, classgen)
		}
	}
	for _, r := range ns.Records {
		if rgen := NewRecordGenerator(cfg, r); rgen != nil {
			gen.SubGenerators = append(gen.SubGenerators, rgen)
		}
	}
	// for _, v := range ns.Unions {
	// }

	return gen
}
