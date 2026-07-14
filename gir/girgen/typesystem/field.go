package typesystem

import (
	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

type Field struct {
	Doc

	Parent Type

	Identifier

	Type CouldBeForeign[Type]

	// TODO: handle getters and setters. This is not as trivial as it seems.
	// even a simple "int" field could mean the length of an array and thus create
	// memory curruption if misused. Also the owner of the field is not always clear.
	// GoGetterName string
	// GoSetterName string

	Readable bool
	Writable bool

	Bits int
}

func NewField(e *env, parent Type, v *gir.Field) *Field {
	e = e.sub("field", v.Name)

	if e.skip(parent, v) {
		return nil
	}

	if v.Private || !(v.IsReadable() || v.Writable) {
		return nil
	}

	if v.Bits > 0 {
		return nil // TODO: what does bits mean?
	}

	return &Field{
		Doc:    NewSimpleDoc(v.Doc),
		Parent: parent,
		Identifier: &baseIdentifier{
			cIndentifier:   v.Name,
			cGoIndentifier: strcases.CGoField(v.Name),
			goIndentifier:  strcases.CGoField(v.Name),
		},
		Bits:     v.Bits,
		Readable: v.IsReadable(),
		Writable: v.Writable,
	}
}
