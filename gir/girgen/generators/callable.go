package generators

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/generators/convert"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

// CallableGenerator generates the code for a Go->C call. We convert the parameters, do the cgo call
// and convert the returns back to go. Some of the go returns are actually "out" params in C though.
type CallableGenerator struct {
	Doc SubGenerator

	//ReceiverConverter contains the go->c converters of the go function receiver if there is one
	ReceiverConverter convert.Converter

	// ParamConverters contains all the go->c converters needed for the c call
	ParamConverters convert.ConverterList

	// ReturnConverters contains all the c->go converters needed to return all values
	ReturnConverters convert.ConverterList

	Signature *typesystem.CallableSignature
}

func (m *CallableGenerator) importReferencedTypes(w file.File) {
	for _, param := range m.Signature.CParameters() {
		if param.Skip || param.Implicit {
			continue
		}
		w.GoImportType(param.Type)
	}

	if m.Signature.CReturn != nil {
		w.GoImportType(m.Signature.CReturn.Type)
	}
}

// GenerateInterfaceSignature implements MethodGenerator.
func (m *CallableGenerator) GenerateInterfaceSignature(w file.File) {
	m.Doc.Generate(w.Go())

	m.importReferencedTypes(w)

	fmt.Fprintln(w.Go(), m.GoInterfaceDeclaration())
}

// Generate implements Generator.
func (m *CallableGenerator) Generate(w *file.Package) {
	m.Doc.Generate(w.Go())

	// register extern callback types:
	for _, param := range m.Signature.GoParameters {
		if cb, ok := param.Type.Type.(*typesystem.Callback); ok {
			w.RegisterExternCallback(cb)
		}
	}

	m.importReferencedTypes(w)

	w.GoImport("runtime")

	fmt.Fprintf(w.Go(), "%s {\n", m.GoSignature())
	w.Go().Indent()

	var decls file.DeclarationWriter

	if m.Signature.InstanceParam != nil {
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", m.Signature.InstanceParam.CName, m.Signature.InstanceParam.CGoType(), m.ReceiverConverter.Metadata())
	}
	for i, param := range m.Signature.GoParameters {
		conv := m.ParamConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", param.CName, param.CGoType(), conv.Metadata())

		// w.GoImportType(param.Type) // don't import here, as the CGo type does not reference the namespace
	}

	// params that are part of the C call still need to be declared
	for _, v := range m.Signature.CParameters() {
		if v.Skip {
			fmt.Fprintf(&decls, "var\t%s\t%s\t// skipped\n", v.CName, v.CGoType())
		}
	}

	for i, ret := range m.Signature.GoReturns {
		conv := m.ReturnConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", ret.CName, ret.CGoType(), conv.Metadata())

		// w.GoImportType(ret.Type) // don't import here, as the CGo type does not reference the namespace
	}

	decls.WriteTo(w.Go())

	w.Go().NewSection()

	if m.ReceiverConverter != nil {
		m.ReceiverConverter.Convert(w)
	}
	for _, c := range m.ParamConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	fmt.Fprintln(w.Go(), m.CGoCall())

	if m.ReceiverConverter != nil {
		fmt.Fprintf(w.Go(), "runtime.KeepAlive(%s)\n", m.Signature.InstanceParam.GoName)
	}
	for _, param := range m.Signature.GoParameters {
		if param.Implicit || param.Skip {
			continue
		}
		fmt.Fprintf(w.Go(), "runtime.KeepAlive(%s)\n", param.GoName)
	}

	w.Go().NewSection()

	// declare the go counterpart of the c return if there is one
	for _, ret := range m.Signature.GoReturns {
		if ret.Implicit || ret.Skip {
			continue
		}
		fmt.Fprintf(&decls, "var\t%s\t%s\n", ret.GoName, ret.GoType())
	}
	decls.WriteTo(w.Go())

	w.Go().NewSection()

	for _, c := range m.ReturnConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	if len(m.Signature.GoReturns) > 0 {
		fmt.Fprintf(w.Go(), "return %s\n", m.Signature.GoReturns.GoIdentifiers())
	}

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewCallableGenerator(cfg *Config, f *typesystem.CallableSignature) *CallableGenerator {
	gen := &CallableGenerator{
		Doc:              cfg.DocGenerator(f),
		Signature:        f,
		ParamConverters:  nil,
		ReturnConverters: nil,
	}

	if f.InstanceParam != nil {
		gen.ReceiverConverter = convert.NewGoToCConverter(f.InstanceParam)
	}
	for _, param := range f.GoParameters {
		gen.ParamConverters = append(gen.ParamConverters, convert.NewGoToCConverter(param))
	}

	for _, param := range f.GoReturns {
		gen.ReturnConverters = append(gen.ReturnConverters, convert.NewCToGoConverter(param))
	}

	return gen
}

// GoSignature returns a string of the go function signature.
func (m *CallableGenerator) GoSignature() string {
	var recv string
	if m.Signature.InstanceParam != nil {
		// tiny hack: use the namespaced go type with zero pointers to not get the interface name for classes and interfaces
		recv = fmt.Sprintf(" (%s *%s)", m.Signature.InstanceParam.GoName, m.Signature.InstanceParam.Type.NamespacedGoType(0))
	}

	var ret string
	if len(m.Signature.GoReturns) == 1 {
		ret = " " + m.Signature.GoReturns[0].GoType()
	} else if len(m.Signature.GoReturns) > 1 {
		ret = " (" + m.Signature.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("func%s %s(%s)%s", recv, m.Signature.GoIndentifier(), m.Signature.GoParameters.GoDeclarations(), ret)
}

// GoInterfaceDeclaration returns a string of the go function signature needed for an interface declaration.
func (m *CallableGenerator) GoInterfaceDeclaration() string {
	var ret string
	if len(m.Signature.GoReturns) == 1 {
		ret = " " + m.Signature.GoReturns[0].GoType()
	} else if len(m.Signature.GoReturns) > 1 {
		ret = " (" + m.Signature.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("%s(%s)%s", m.Signature.GoIndentifier(), m.Signature.GoParameters.GoTypes(), ret)
}

// CGoCall returns a string of the cgo function call.
func (m *CallableGenerator) CGoCall() string {
	creturn := m.Signature.CGoReturn()
	var ret string
	if creturn != nil {
		ret = fmt.Sprintf("%s = ", creturn.CName)
	}

	var callExpressions []string

	for _, param := range m.Signature.CParameters() {
		if param.Direction == "out" {
			callExpressions = append(callExpressions, "&"+param.CName)
		} else {
			callExpressions = append(callExpressions, param.CName)
		}
	}

	return fmt.Sprintf("%s%s(%s)", ret, m.Signature.CGoIndentifier(), strings.Join(callExpressions, ", "))
}
