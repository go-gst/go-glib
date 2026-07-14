package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type MarshalGenerator struct {
	Type typesystem.Marshalable

	WrapFunction string // may be equal to GoName for Enums and simple conversions
	// ValueFunction is the method on coregllib.Value to use
	ValueFunction string
}

func (g *MarshalGenerator) Generate(w *file.Package) {
	w.GoImport("unsafe")

	w.GoImportNamespace(g.Type.Value().Namespace)

	fmt.Fprintf(w.Go(), "func marshal%s(p unsafe.Pointer) (any, error) {\n", g.Type.GoType(0))
	if g.WrapFunction != "" {
		fmt.Fprintf(w.Go(), "\treturn %s(%s(p).%s()), nil\n", g.WrapFunction, g.Type.Value().WithForeignNamespace(g.Type.Value().Type.FromGlibBorrowFunction), g.ValueFunction)
	} else {
		fmt.Fprintf(w.Go(), "\treturn %s(p).%s(), nil\n", g.Type.Value().WithForeignNamespace(g.Type.Value().Type.FromGlibBorrowFunction), g.ValueFunction)
	}
	fmt.Fprintf(w.Go(), "}\n")
}

func NewMarshalEnumGenerator(typ typesystem.Marshalable) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  typ.GoType(0),
		ValueFunction: "Enum",
	}
}

func NewMarshalBifieldGenerator(typ typesystem.Marshalable) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  typ.GoType(0),
		ValueFunction: "Flags",
	}
}

func NewMarshalObjectGenerator(typ typesystem.Marshalable) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  "",
		ValueFunction: "Object",
	}
}
