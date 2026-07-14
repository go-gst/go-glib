package generators

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/generators/convert"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type VirtualMethodGenerator struct {
	Doc       SubGenerator
	ParentDoc SubGenerator
	*typesystem.VirtualMethod

	// VirtualParamConverters contains all the c->go in param converters needed for the go call
	// in the virtual method trampoline
	VirtualParamConverters convert.ConverterList

	// VirtualReturnConverters contains all the go->c converters needed for the returns and out params
	VirtualReturnConverters convert.ConverterList

	// ParentParamConverters contains all the go->c in param converters needed for the cgo call for the parent
	// virtual method. This does not contain the instance parameter, as we handle it specially in the parent call
	// block.
	ParentParamConverters convert.ConverterList

	// ParentReturnConverters contains all the c->go converters needed for the returns and out params for the parent
	// virtual method.
	ParentReturnConverters convert.ConverterList
}

var _ MethodGenerator = (*VirtualMethodGenerator)(nil)

// Generate implements VirtualMethodGenerator.
func (v *VirtualMethodGenerator) Generate(w *file.Package) {
	v.generateExport(w)

	v.generateParentCall(w)
}

func (v *VirtualMethodGenerator) generateParentCall(w *file.Package) {
	v.generateParentCPreamble(w)

	v.ParentDoc.Generate(w.Go())

	// register extern callback types:
	for _, param := range v.GoParameters {
		if cb, ok := param.Type.Type.(*typesystem.Callback); ok {
			w.RegisterExternCallback(cb)
		}
	}

	for _, param := range v.CParameters() {
		if param.Skip || param.Implicit {
			continue
		}
		w.GoImportType(param.Type)
	}

	if v.CReturn != nil {
		w.GoImportType(v.CReturn.Type)
	}

	w.GoImport("runtime")

	fmt.Fprintf(w.Go(), "%s {\n", v.ParentGoSignature())
	w.Go().Indent()

	var decls file.DeclarationWriter

	fmt.Fprintf(&decls, "var\t%s\t%s\n", v.InstanceParam.CName, v.InstanceParam.CGoType())

	for i, param := range v.GoParameters {
		conv := v.ParentParamConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", param.CName, param.CGoType(), conv.Metadata())

		// w.GoImportType(param.Type) // don't import here, as the CGo type does not reference the namespace
	}

	// params that are part of the C call still need to be declared
	for _, v := range v.CParameters() {
		if v.Skip {
			fmt.Fprintf(&decls, "var\t%s\t%s\t// skipped\n", v.CName, v.CGoType())
		}
	}

	for i, ret := range v.GoReturns {
		conv := v.ParentReturnConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", ret.CName, ret.CGoType(), conv.Metadata())

		// w.GoImportType(ret.Type) // don't import here, as the CGo type does not reference the namespace
	}

	decls.WriteTo(w.Go())

	w.Go().NewSection()

	if v.isVirtualOnClass() {
		fmt.Fprintf(w.Go(), "parentclass := (*%s)(classdata.PeekParentClass(%s(%s)))\n", v.parentTypeStruct().CGoType(0), v.Parent.GoUnsafeToGlibNoneFunction(), v.InstanceParam.GoName)
	} else {
		fmt.Fprintf(w.Go(), "parentclass := (*%s)(classdata.PeekParentInterface(%s(%s), uint64(%s)))\n", v.parentTypeStruct().CGoType(0), v.Parent.GoUnsafeToGlibNoneFunction(), v.InstanceParam.GoName, v.parentInterfaceGType())
	}

	w.Go().NewSection()
	for _, c := range v.ParentParamConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	fmt.Fprintln(w.Go(), v.ParentCGoCall())

	fmt.Fprintf(w.Go(), "runtime.KeepAlive(%s)\n", v.InstanceParam.GoName)
	for _, param := range v.GoParameters {
		if param.Implicit || param.Skip {
			continue
		}
		fmt.Fprintf(w.Go(), "runtime.KeepAlive(%s)\n", param.GoName)
	}

	w.Go().NewSection()

	// declare the go counterpart of the c return if there is one
	for _, ret := range v.GoReturns {
		if ret.Implicit || ret.Skip {
			continue
		}
		fmt.Fprintf(&decls, "var\t%s\t%s\n", ret.GoName, ret.GoType())
	}
	decls.WriteTo(w.Go())

	w.Go().NewSection()

	for _, c := range v.ParentReturnConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	if len(v.GoReturns) > 0 {
		fmt.Fprintf(w.Go(), "return %s\n", v.GoReturns.GoIdentifiers())
	}

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")
}

func (c *VirtualMethodGenerator) isVirtualOnClass() bool {
	switch c.Parent.(type) {
	case *typesystem.Class:
		return true
	case *typesystem.Interface:
		return false
	default:
		panic(fmt.Sprintf("unexpected parent type %T", c.Parent))
	}
}

func (c *VirtualMethodGenerator) parentInterfaceGType() string {
	switch p := c.Parent.(type) {
	case *typesystem.Interface:
		return p.GoTypeName()
	default:
		panic(fmt.Sprintf("unexpected parent interface type %T", p))
	}
}

func (c *VirtualMethodGenerator) parentTypeStruct() *typesystem.Record {
	switch p := c.Parent.(type) {
	case *typesystem.Interface:
		return p.TypeStruct
	case *typesystem.Class:
		return p.TypeStruct
	default:
		panic(fmt.Sprintf("unexpected parent type %T", p))
	}
}

// generateParentCPreamble generates the C function that is able to call the C function pointer
// on the parent class.
func (v *VirtualMethodGenerator) generateParentCPreamble(w *file.Package) {

	var paramDecls []string
	// args is the list of arguments needed to call the casted function pointer
	var args []string

	// paramTypes is the list of arguments in the function pointer
	var paramTypes []string

	// add the function pointer as the first parameter which we will need to cast
	paramDecls = append(paramDecls, "void* fnptr")

	for _, p := range v.VirtualMethod.CParameters() {
		ctype := p.CType()
		if p.Direction == "out" {
			ctype += "*" // add the trimmed pointer back
		}

		args = append(args, p.CName)
		paramTypes = append(paramTypes, ctype)
		paramDecls = append(
			paramDecls,
			fmt.Sprintf("%s %s", ctype, p.CName),
		)
	}

	fmt.Fprintf(w.C(), "%s %s(%s) {\n", v.VirtualMethod.CReturn.CType(), v.VirtualMethod.ParentTrampolineName, strings.Join(paramDecls, ", "))
	w.C().Indent()
	fmt.Fprintf(w.C(), "return ((%s (*) (%s))(fnptr))(%s);\n", v.CReturn.CType(), strings.Join(paramTypes, ", "), strings.Join(args, ", "))
	w.C().Unindent()
	fmt.Fprintf(w.C(), "}\n")
}

// generateExport generates the exported Cgo trampoline function for the virtual method. We do not do much here, because we do not have the necessary
// type information of the go subclass to call the function directly. Instead we lookup a function pointer that accepts the same paremeters as the
// trampoline, which handles all the necessary conversions and calls the go function pointer. See [VirtualMethodGenerator.generateTrampoline] for more information.
func (c *VirtualMethodGenerator) generateExport(pkg *file.Package) {
	w := &pkg.Exported

	w.GoImportCore("classdata")
	w.GoImport("unsafe")

	fmt.Fprintf(w.Go(), "//export %s\n", c.TrampolineName)

	fmt.Fprintf(w.Go(), "%s {\n", c.CGoTrampolineSignature())

	w.Go().Indent()

	fmt.Fprintf(w.Go(), "var fn func%s\n", c.CGoTrampolineTail())
	fmt.Fprintf(w.Go(), "{\n")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(%s), \"%s\").(func%s)\n", c.InstanceParam.CName, c.TrampolineName, c.CGoTrampolineTail())

	fmt.Fprintf(w.Go(), "if fn == nil {\n")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "panic(\"%s: no function pointer found\")\n", c.TrampolineName)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")

	if c.CReturn != nil && c.CReturn.Type.Type != typesystem.Void {
		fmt.Fprintf(w.Go(), "return fn(%s)\n", c.CParameters().CIdentifiers())
	} else {
		fmt.Fprintf(w.Go(), "fn(%s)\n", c.CParameters().CIdentifiers())
	}

	w.Go().Unindent()

	fmt.Fprintln(w.Go(), "}")
	fmt.Fprintln(w.Go())
}

// generateTrampoline is called from the overrides generator to generate the virtual function in a place where we have enough type information
// to correctly call the overloaded function. This is done by using the generic Instance type from the apply overloads scope.
//
// this must end with a comma before the newline, as the overrides generator will call this from a function parameter list.
func (vf *VirtualMethodGenerator) generateTrampoline(w file.File, genericType, goFn string) {
	fmt.Fprintf(w.Go(), "func%s {\n", vf.CGoTrampolineTail())
	w.Go().Indent()

	var decls file.DeclarationWriter

	fmt.Fprintf(&decls, "var\t%s\t%s\t// go %s subclass\n", vf.InstanceParam.GoName, genericType, vf.InstanceParam.Type.Type.CType(0))

	for i, param := range vf.GoParameters {
		if param.Implicit || param.Skip {
			continue
		}
		conv := vf.VirtualParamConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", param.GoName, param.GoType(), conv.Metadata())

		w.GoImportType(param.Type)
	}

	for i, ret := range vf.GoReturns {
		if ret.Implicit || ret.Skip {
			continue
		}

		conv := vf.VirtualReturnConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", ret.GoName, ret.GoType(), conv.Metadata())

		w.GoImportType(ret.Type)
	}

	decls.WriteTo(w.Go())

	w.Go().NewSection()

	// the instanceparam needs to be borrowed and converted to the instance data
	fmt.Fprintf(w.Go(), "%s = %s(unsafe.Pointer(%s)).UnsafeLoadInstanceFromPrivateData().(%s)\n", vf.InstanceParam.GoName, vf.Parent.GoUnsafeFromGlibBorrowFunction(), vf.InstanceParam.CName, genericType)

	for _, c := range vf.VirtualParamConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	goReturns := vf.GoReturns.GoIdentifiers()

	if goReturns == "" {
		fmt.Fprintf(w.Go(), "%s(%s)\n", goFn, vf.GoIdentifiers())
	} else {
		fmt.Fprintf(w.Go(), "%s = %s(%s)\n", goReturns, goFn, vf.GoIdentifiers())
	}

	w.Go().NewSection()

	for _, c := range vf.VirtualReturnConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	if vf.CReturn != nil && vf.CReturn.Type.Type != typesystem.Void {
		fmt.Fprintf(w.Go(), "return %s\n", vf.CReturn.CName)
	}

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "},\n")
}

// GenerateClassOverrideField generates the field name and type for the class override structs field.
//
// as opposed to interface signatures this must include the instance parameter as a first parameter of
// the method signature. The type of the instance parameter is always OverridesInstanceGenericType, as that is the
// generic type of the struct.
func (m *VirtualMethodGenerator) GenerateClassOverrideField(w file.File) {
	var ret string
	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].GoType()
	} else if len(m.GoReturns) > 1 {
		ret = " (" + m.GoReturns.GoTypes() + ")"
	}

	var vparams = []string{
		overridesInstanceGenericType, // first param is the generic instance
	}

	for _, param := range m.GoParameters {
		if param.Skip || param.Implicit {
			continue
		}

		w.GoImportType(param.Type)

		vparams = append(vparams, param.GoType())
	}

	m.Doc.Generate(w.Go())
	fmt.Fprintf(w.Go(), "%s func(%s)%s\n", m.GoName, strings.Join(vparams, ", "), ret)
}

// CGoTrampolineSignature returns a string of the cgo trampoline function signature.
func (m *VirtualMethodGenerator) CGoTrampolineSignature() string {
	return fmt.Sprintf("func %s%s", m.VirtualMethod.TrampolineName, m.CGoTrampolineTail())
}

func (m *VirtualMethodGenerator) CGoTrampolineTail() string {
	var ret string
	if m.VirtualMethod.CReturn != nil && m.VirtualMethod.CReturn.Type.Type != typesystem.Void {
		ret = fmt.Sprintf(" (%s %s)", m.VirtualMethod.CReturn.CName, m.VirtualMethod.CReturn.CGoType())
	}

	paramDecls := make([]string, 0, len(m.VirtualMethod.CParameters()))

	for _, p := range m.VirtualMethod.CParameters() {
		additionalPointer := ""
		if p.Direction == "out" {
			additionalPointer = "*"
		}

		paramDecls = append(
			paramDecls,
			fmt.Sprintf("%s %s%s", p.CName, additionalPointer, p.CGoType()),
		)
	}

	return fmt.Sprintf("(%s)%s", strings.Join(paramDecls, ", "), ret)
}

func (m *VirtualMethodGenerator) GoIdentifiers() string {
	var params []string

	params = append(params, m.VirtualMethod.InstanceParam.GoName)

	for _, param := range m.VirtualMethod.GoParameters {
		if param.Skip || param.Implicit {
			continue
		}

		params = append(params, param.GoName)
	}
	return strings.Join(params, ", ")
}

// ParentGoSignature returns a string of the go function signature for the parent call.
func (m *VirtualMethodGenerator) ParentGoSignature() string {
	recv := fmt.Sprintf(" (%s *%s)", m.InstanceParam.GoName, m.InstanceParam.Type.NamespacedGoType(0))

	var ret string
	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].GoType()
	} else if len(m.GoReturns) > 1 {
		ret = " (" + m.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("func%s %s(%s)%s", recv, m.ParentName, m.GoParameters.GoDeclarations(), ret)
}

// ParentCGoCall returns a string of the cgo function call.
func (m *VirtualMethodGenerator) ParentCGoCall() string {
	creturn := m.CGoReturn()
	var ret string
	if creturn != nil {
		ret = fmt.Sprintf("%s = ", creturn.CName)
	}

	var callExpressions []string

	callExpressions = append(callExpressions, fmt.Sprintf("unsafe.Pointer(parentclass.%s)", m.Invoker.CGoIndentifier()))

	for _, param := range m.CParameters() {
		if param.Direction == "out" {
			callExpressions = append(callExpressions, "&"+param.CName)
		} else {
			callExpressions = append(callExpressions, param.CName)
		}
	}

	return fmt.Sprintf("%sC.%s(%s)", ret, m.ParentTrampolineName, strings.Join(callExpressions, ", "))
}

// GenerateInterfaceSignature implements MethodGenerator.
func (v *VirtualMethodGenerator) GenerateInterfaceSignature(w file.File) {
	v.ParentDoc.Generate(w.Go())

	var ret string
	if len(v.GoReturns) == 1 {
		ret = " " + v.GoReturns[0].GoType()
	} else if len(v.GoReturns) > 1 {
		ret = " (" + v.GoReturns.GoTypes() + ")"
	}

	fmt.Fprintf(w.Go(), "%s(%s)%s\n", v.ParentName, v.GoParameters.GoDeclarations(), ret)
}

func NewVirtualMethodGenerator(cfg *Config, vfunc *typesystem.VirtualMethod) *VirtualMethodGenerator {
	doc := cfg.DocGenerator(vfunc)
	g := &VirtualMethodGenerator{
		Doc: doc.WithPrependParagraphs(
			fmt.Sprintf("// %s allows you to override the implementation of the virtual method %s.\n", vfunc.GoName, vfunc.Invoker.CIndentifier()),
		),
		ParentDoc: doc.WithPrependParagraphs(
			fmt.Sprintf("%s calls the default implementations of the `%s.%s` virtual method.\nThis function's behavior is not defined when the parent does not implement the virtual method.\n", vfunc.ParentName, vfunc.Parent.CType(0), vfunc.Invoker.CIndentifier()),
		),
		VirtualMethod: vfunc,
	}

	for _, param := range vfunc.GoParameters {
		g.VirtualParamConverters = append(g.VirtualParamConverters, convert.NewCToGoConverter(param))
	}
	for _, param := range vfunc.GoReturns {
		g.VirtualReturnConverters = append(g.VirtualReturnConverters, convert.NewGoToCConverter(param))
	}

	g.ParentParamConverters = append(g.ParentParamConverters, convert.NewGoToCConverter(vfunc.InstanceParam))

	for _, param := range vfunc.GoParameters {
		g.ParentParamConverters = append(g.ParentParamConverters, convert.NewGoToCConverter(param))
	}
	for _, param := range vfunc.GoReturns {
		g.ParentReturnConverters = append(g.ParentReturnConverters, convert.NewCToGoConverter(param))
	}

	return g
}
