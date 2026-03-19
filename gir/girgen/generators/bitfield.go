package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type BitfieldMember struct {
	Doc SubGenerator

	*typesystem.Member
}

type BitfieldGenerator struct {
	Doc SubGenerator

	*typesystem.Bitfield

	MethodReceiver string

	Marshaler Generator

	Members []BitfieldMember

	SubGenerators GeneratorList
}

func (g *BitfieldGenerator) Generate(w *file.Package) {
	g.Doc.Generate(w.Go())

	w.GoImport("strings")

	fmt.Fprintf(w.Go(), "type %s C.gint\n\n", g.GoType(0))

	fmt.Fprintln(w.Go(), "const (")
	w.Go().Indent()
	for _, m := range g.Members {
		m.Doc.Generate(w.Go())
		fmt.Fprintf(w.Go(), "%s %s = %s\n", m.GoIndentifier(), g.GoType(0), m.Value)
	}
	w.Go().Unindent()
	fmt.Fprint(w.Go(), ")\n\n")

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
	}

	fmt.Fprintf(w.Go(), "// Has returns true if %s contains other\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "func (%s %s) Has(other %s) bool {\n", g.MethodReceiver, g.GoType(0), g.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn (%s & other) == other\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "}\n\n")

	if g.CanMarshal() {
		// GoValueInitializer assertion:
		fmt.Fprintf(w.Go(), "var _ %s = %s(0)\n\n", g.Value().WithForeignNamespace("GoValueInitializer"), g.GoType(0))

		fmt.Fprintf(w.Go(), "func (f %s) GoValueType() %s {\n", g.GoType(0), g.Type().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "return %s\n", g.GoTypeName())
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")

		fmt.Fprintf(w.Go(), "func (f %s) SetGoValue(v *%s) {\n", g.GoType(0), g.Value().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "v.SetFlags(int32(f))\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	// Stringer:
	fmt.Fprintf(w.Go(), "func (f %s) String() string {\n", g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "if f == 0 {\n")
	fmt.Fprintf(w.Go(), "\treturn \"%s(0)\"\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "}\n\n")
	fmt.Fprintf(w.Go(), "var parts []string\n")
	for _, member := range g.Members {
		fmt.Fprintf(w.Go(), "if (f & %s) != 0 {\n", member.GoIndentifier())
		fmt.Fprintf(w.Go(), "\tparts = append(parts, \"%s\")\n", member.GoIndentifier())
		fmt.Fprintf(w.Go(), "}\n")
	}
	fmt.Fprintf(w.Go(), "return \"%s(\" + strings.Join(parts, \"|\") + \")\"\n", g.GoType(0))
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	g.SubGenerators.Generate(w)
}

func NewBitfieldGenerator(cfg *Config, bf *typesystem.Bitfield) *BitfieldGenerator {
	var members []BitfieldMember

	for _, m := range bf.Members {
		mm := BitfieldMember{
			Doc:    cfg.DocGenerator(m),
			Member: m,
		}
		members = append(members, mm)
	}

	var marshalGen Generator

	if bf.GLibGetType() != "" {
		marshalGen = NewMarshalBifieldGenerator(bf)
	}

	gen := &BitfieldGenerator{
		Doc:      cfg.DocGenerator(bf),
		Bitfield: bf,

		Members:        members,
		MethodReceiver: strcases.ReceiverName(bf.GoType(0)),
		Marshaler:      marshalGen,
	}

	for _, f := range bf.Functions {
		gen.SubGenerators = append(gen.SubGenerators, NewCallableGenerator(cfg, f))
	}

	return gen
}
