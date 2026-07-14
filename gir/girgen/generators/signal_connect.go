package generators

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type SignalConnectGenerator struct {
	Doc SubGenerator

	*typesystem.Signal
}

// Generate implements MethodGenerator.
func (s *SignalConnectGenerator) Generate(w *file.Package) {
	s.Doc.Generate(w.Go())

	for _, p := range s.Parameters() {
		w.GoImportType(p.Type)
	}

	var ret string
	if s.Signal.GoReturn() != nil {
		ret = " " + s.Signal.GoReturn().GoType()
	}

	// we don't need to import gobject here, because class generators already import it

	fmt.Fprintf(w.Go(), "func (o *%s) %s(fn func(%s)%s) %s {\n", s.InstanceParam.Type.Type.GoType(0), s.GoName, s.Parameters().GoTypes(), ret, s.GObject.WithForeignNamespace("SignalHandle"))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "return %s.Connect(\"%s\", fn)\n", s.objectIdentifier(), s.Name)
	w.Go().Unindent()
	fmt.Fprintln(w.Go(), "}")
	fmt.Fprintln(w.Go(), "")
}

// objectIdentifier returns the struct field that contains the gobject. This is only needed for interfaces.
func (s *SignalConnectGenerator) objectIdentifier() string {
	if _, ok := s.InstanceParam.Type.Type.(*typesystem.Interface); ok {
		return fmt.Sprintf("o.%s", InterfaceInstanceStructFieldName)
	}

	return "o"
}

// GenerateInterfaceSignature implements MethodGenerator.
func (s *SignalConnectGenerator) GenerateInterfaceSignature(w file.File) {
	s.Doc.Generate(w.Go())

	var ret string
	if s.Signal.GoReturn() != nil {
		ret = " " + s.Signal.GoReturn().GoType()
	}

	fmt.Fprintf(w.Go(), "%s(func(%s)%s) %s\n", s.GoName, s.Parameters().GoTypes(), ret, s.GObject.WithForeignNamespace("SignalHandle"))
}

func NewSignalConnectGenerator(cfg *Config, sig *typesystem.Signal) *SignalConnectGenerator {
	return &SignalConnectGenerator{
		Doc:    cfg.DocGenerator(sig),
		Signal: sig,
	}
}
