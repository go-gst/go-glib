package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type AliasGenerator struct {
	Doc SubGenerator

	*typesystem.Alias
}

func (g *AliasGenerator) Generate(w *file.Package) {
	g.Doc.Generate(w.Go())

	// TODO: imports of foreign type namespaces

	fmt.Fprintf(w.Go(), "type %s = %s\n", g.GoType(0), g.AliasedType.NamespacedGoType(0))
}

func NewAliasGenerator(cfg *Config, alias *typesystem.Alias) *AliasGenerator {
	return &AliasGenerator{
		Doc:   cfg.DocGenerator(alias),
		Alias: alias,
	}
}
