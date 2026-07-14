package generators

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type GoOverridesGenerator struct {
	For typesystem.Type

	VirtualMethods VirtualMethodGeneratorList
}

// Generate implements Generator.
func (g *GoOverridesGenerator) Generate(w *file.Package) {
	var typestruct *typesystem.Record
	var overridesName string
	var parentOverridesFieldName string
	var parentOverridesType string
	var parentName string
	var applyOverridesName string
	var parentApplyOverridesName string
	switch t := g.For.(type) {
	case *typesystem.Class:
		overridesName = t.GoExtendOverrideStructName
		parentOverridesFieldName = t.Parent.Type.GoExtendOverrideStructName
		parentOverridesType = t.Parent.WithForeignNamespace(t.Parent.Type.GoExtendOverrideStructName)
		parentName = t.Parent.NamespacedGoType(1)
		applyOverridesName = t.GoUnsafeApplyOverridesName
		parentApplyOverridesName = t.Parent.WithForeignNamespace(t.Parent.Type.GoUnsafeApplyOverridesName)
		typestruct = t.TypeStruct
	case *typesystem.Interface:
		overridesName = t.GoExtendOverrideStructName
		applyOverridesName = t.GoUnsafeApplyOverridesName
		typestruct = t.TypeStruct
	default:
		panic("invalid type")
	}

	fmt.Fprintf(w.Go(), "// %s is the struct used to override the default implementation of virtual methods.\n", overridesName)
	fmt.Fprintf(w.Go(), "// it is generic over the extending instance type.\n")
	fmt.Fprintf(w.Go(), "type %s[%s %s] struct {\n", overridesName, overridesInstanceGenericType, g.For.GoType(1))
	w.Go().Indent()
	if parentOverridesType != "" {
		fmt.Fprintf(w.Go(), "// %s allows you to override virtual methods from the parent class %s\n", parentOverridesType, parentName)
		fmt.Fprintf(w.Go(), "%s[%s]\n\n", parentOverridesType, overridesInstanceGenericType)
	}
	g.VirtualMethods.GenerateClassOverrideFields(w)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "// %s applies the overrides to init the gclass by setting the trampoline functions.\n", applyOverridesName)
	fmt.Fprintf(w.Go(), "// This is used by the bindings internally and only exported for visibility to other bindings code.\n")
	fmt.Fprintf(w.Go(), "func %s[%s %s](gclass unsafe.Pointer, overrides %s[%s]) {\n", applyOverridesName, overridesInstanceGenericType, g.For.GoType(1), overridesName, overridesInstanceGenericType)
	w.Go().Indent()

	if parentApplyOverridesName != "" {
		fmt.Fprintf(w.Go(), "%s(gclass, overrides.%s)\n", parentApplyOverridesName, parentOverridesFieldName)
		w.Go().NewSection()
	}

	if len(g.VirtualMethods) > 0 {
		fmt.Fprintf(w.Go(), "pclass := (%s)(gclass)\n", typestruct.CGoType(1))
		w.Go().NewSection()
	}

	if len(g.VirtualMethods) > 0 {
		w.GoImportCore("classdata")
	}

	for _, virtual := range g.VirtualMethods {
		overridesFnFieldName := fmt.Sprintf("overrides.%s", virtual.GoName)

		// declare the extern C callback in the C preamble:
		var cparams []string
		for _, param := range virtual.CParameters() {
			cparams = append(cparams, param.CType())
		}
		fmt.Fprintf(w.C(), "extern %s %s(%s);\n", virtual.CReturn.CType(), virtual.TrampolineName, strings.Join(cparams, ", "))

		fmt.Fprintf(w.Go(), "if %s != nil {\n", overridesFnFieldName)
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "pclass.%s = (*[0]byte)(C.%s)\n", virtual.Invoker.CGoIndentifier(), virtual.TrampolineName)

		fmt.Fprintf(w.Go(), "classdata.StoreVirtualMethod(\n")
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "unsafe.Pointer(pclass),\n")
		fmt.Fprintf(w.Go(), "%q,\n", virtual.TrampolineName)
		virtual.generateTrampoline(w, overridesInstanceGenericType, overridesFnFieldName) // must end with a comma before newline
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), ")\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n")
		w.Go().NewSection()
	}
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	g.VirtualMethods.Generate(w)
}

var _ Generator = (*GoOverridesGenerator)(nil)

func NewGoOverridesGenerator(cfg *Config, parent typesystem.Type) *GoOverridesGenerator {
	var virtuals []*typesystem.VirtualMethod

	switch t := parent.(type) {
	case *typesystem.Class:
		virtuals = t.VirtualMethods
	case *typesystem.Interface:
		virtuals = t.VirtualMethods
	default:
		panic("invalid type")
	}

	gen := &GoOverridesGenerator{
		For: parent,
	}

	for _, v := range virtuals {
		gen.VirtualMethods = append(gen.VirtualMethods, NewVirtualMethodGenerator(cfg, v))
	}

	return gen
}
