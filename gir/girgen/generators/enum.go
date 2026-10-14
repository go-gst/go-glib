package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type EnumMember struct {
	Doc SubGenerator

	*typesystem.Member
}

type EnumGenerator struct {
	Doc SubGenerator

	*typesystem.Enum

	Members []EnumMember

	Marshaler Generator

	SubGenerators GeneratorList
}

func (g *EnumGenerator) Generate(w *file.Package) {
	w.GoImport("fmt")

	g.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "type %s C.int\n", g.GoType(0))
	w.Go().NewSection()

	fmt.Fprintf(w.Go(), "const (\n")

	w.Go().Indent()
	for _, member := range g.Members {
		member.Doc.Generate(w.Go())

		fmt.Fprintf(w.Go(), "%s %s = %s\n", member.GoIndentifier(), g.GoType(0), member.Value)
	}
	w.Go().Unindent()

	fmt.Fprint(w.Go(), ")\n")
	w.Go().NewSection()

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
	}

	if g.CanMarshal() {
		// GoValueInitializer assertion:
		fmt.Fprintf(w.Go(), "var _ %s = %s(0)\n", g.Value().WithForeignNamespace("GoValueInitializer"), g.GoType(0))
		w.Go().NewSection()

		fmt.Fprintf(w.Go(), "func (e %s) GoValueType() %s {\n", g.GoType(0), g.Type().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "return %s\n", g.GoTypeName())
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()

		fmt.Fprintf(w.Go(), "func (e %s) SetGoValue(v *%s) {\n", g.GoType(0), g.Value().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "v.SetEnum(int32(e))\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}

	// Stringer:
	fmt.Fprintf(w.Go(), "func (e %s) String() string {\n", g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "switch e {\n")
	for _, member := range g.Enum.Members.Uniques() {
		fmt.Fprintf(w.Go(), "case %s:\n\treturn \"%s\"\n", member.GoIndentifier(), member.GoIndentifier())
	}
	fmt.Fprintf(w.Go(), "default:\n\treturn fmt.Sprintf(\"%s(%%d)\", e)\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().NewSection()

	g.SubGenerators.Generate(w)
}

func NewEnumGenerator(cfg *Config, enum *typesystem.Enum) *EnumGenerator {
	members := make([]EnumMember, 0, len(enum.Members))

	for _, member := range enum.Members {
		members = append(members, EnumMember{
			Doc: cfg.DocGenerator(member),

			Member: member,
		})
	}

	var marshalGen Generator

	if enum.GLibGetType() != "" {
		marshalGen = NewMarshalEnumGenerator(enum)
	}

	gen := &EnumGenerator{
		Doc:       cfg.DocGenerator(enum),
		Enum:      enum,
		Members:   members,
		Marshaler: marshalGen,
	}

	for _, f := range enum.Functions {
		gen.SubGenerators = append(gen.SubGenerators, NewCallableGenerator(cfg, f))
	}

	return gen
}
