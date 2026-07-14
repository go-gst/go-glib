package generators

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/generators/convert"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CallbackGenerator struct {
	Doc SubGenerator

	*typesystem.Callback

	// ParamConverters contains all the c->go in param converters needed for the go call
	ParamConverters convert.ConverterList

	// ReturnConverters contains all the go->c converters needed for the returns and out params
	ReturnConverters convert.ConverterList
}

// Generate implements Generator.
func (c *CallbackGenerator) Generate(w *file.Package) {
	c.generateGo(w)
	c.generateExport(w)
}

func (c *CallbackGenerator) generateGo(w *file.Package) {
	c.Doc.Generate(w.Go())

	ret := c.GoReturns.GoDeclarations()

	if ret != "" {
		ret = " (" + ret + ")"
	}

	for _, param := range c.GoParameters {
		if param.Skip || param.Implicit {
			continue
		}

		w.GoImportType(param.Type)
	}

	for _, param := range c.GoReturns {
		if param.Skip || param.Implicit {
			continue
		}

		w.GoImportType(param.Type)
	}

	fmt.Fprintf(w.Go(), "type %s func(%s)%s\n", c.GoType(0), c.GoParameters.GoDeclarations(), ret)

	fmt.Fprintln(w.Go())
}

func (c *CallbackGenerator) generateExport(pkg *file.Package) {
	w := &pkg.Exported

	fmt.Fprintf(w.Go(), "//export %s\n", c.TrampolineName)

	// TODO: the out params here are missing a pointer because it was stripped in the type resolution

	fmt.Fprintf(w.Go(), "%s {\n", c.CGoTrampolineSignature())

	w.Go().Indent()

	fmt.Fprintf(w.Go(), "var fn %s\n", c.GoType(0)) // declare fn as the callback itself

	w.GoImport("unsafe")
	w.GoImportCore("userdata")

	fmt.Fprintf(w.Go(), "{\n")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "v := userdata.Load(unsafe.Pointer(%s))\n", c.UserdataParam.CName)
	fmt.Fprintf(w.Go(), "if v == nil {\n")
	fmt.Fprintf(w.Go(), "\tpanic(`callback not found`)\n")
	fmt.Fprintf(w.Go(), "}\n")
	fmt.Fprintf(w.Go(), "fn = v.(%s)\n", c.GoType(0))
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")

	w.Go().NewSection()

	var decls file.DeclarationWriter

	for i, param := range c.Callback.GoParameters {
		if param.Implicit || param.Skip {
			continue
		}
		conv := c.ParamConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", param.GoName, param.GoType(), conv.Metadata())

		w.GoImportType(param.Type)
	}
	for i, ret := range c.Callback.GoReturns {
		if ret.Implicit || ret.Skip {
			continue
		}

		conv := c.ReturnConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", ret.GoName, ret.GoType(), conv.Metadata())

		w.GoImportType(ret.Type)
	}

	decls.WriteTo(w.Go())

	w.Go().NewSection()

	for _, c := range c.ParamConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	goReturns := c.GoReturns.GoIdentifiers()

	if goReturns == "" {
		fmt.Fprintf(w.Go(), "fn(%s)\n", c.GoParameters.GoIdentifiers())
	} else {
		fmt.Fprintf(w.Go(), "%s = fn(%s)\n", goReturns, c.GoParameters.GoIdentifiers())
	}

	w.Go().NewSection()

	for _, c := range c.ReturnConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	if c.CReturn != nil && c.CReturn.Type.Type != typesystem.Void {
		fmt.Fprintf(w.Go(), "return %s\n", c.CReturn.CName)
	}

	w.Go().Unindent()

	fmt.Fprintln(w.Go(), "}")
	fmt.Fprintln(w.Go())
}

func NewCallbackGenerator(cfg *Config, cb *typesystem.Callback) *CallbackGenerator {
	if cb.InstanceParam != nil {
		panic("callback with instance param unimplemented")
	}

	g := &CallbackGenerator{
		Doc:      cfg.DocGenerator(cb),
		Callback: cb,
	}

	if cb.InstanceParam != nil {
		panic("callback with instance param")
	}

	for _, param := range cb.GoParameters {
		g.ParamConverters = append(g.ParamConverters, convert.NewCToGoConverter(param))
	}

	for _, param := range cb.GoReturns {
		g.ReturnConverters = append(g.ReturnConverters, convert.NewGoToCConverter(param))
	}

	return g
}

// CGoTrampolineSignature returns a string of the cgo trampoline function signature.
func (m *CallbackGenerator) CGoTrampolineSignature() string {
	var ret string
	if m.Callback.CReturn != nil && m.Callback.CReturn.Type.Type != typesystem.Void {
		ret = fmt.Sprintf(" (%s %s)", m.Callback.CReturn.CName, m.Callback.CReturn.CGoType())
	}

	paramDecls := make([]string, 0, len(m.Callback.CParameters()))

	for _, p := range m.Callback.CParameters() {
		additionalPointer := ""
		if p.Direction == "out" {
			additionalPointer = "*"
		}

		paramDecls = append(
			paramDecls,
			fmt.Sprintf("%s %s%s", p.CName, additionalPointer, p.CGoType()),
		)
	}

	return fmt.Sprintf("func %s(%s)%s", m.Callback.TrampolineName, strings.Join(paramDecls, ", "), ret)
}
