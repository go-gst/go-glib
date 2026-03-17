package typesystem

import (
	"fmt"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

type VirtualMethod struct {
	Doc
	// Parent is the class or interface that this virtual method belongs to.
	Parent ConvertibleType

	// TrampolineName is the name of the trampoline function that needs to be
	// called when the virtual function was overridden.
	TrampolineName string

	// ParentTrampolineName is the C function name that is used to call the C function pointer
	// of the virtual method of the parent class. This is needed because we cannot cast c function pointers
	// to callable functions.
	ParentTrampolineName string

	Invoker *Field

	// GoName is the name of the override in the Overrides struct.
	GoName string

	// ParentName is the name of the method on the instance that calls the default implementation
	// on the parent class.
	ParentName string

	*Parameters
}

func NewVirtualMethod(e *env, parent ConvertibleType, typestruct *Record, v *gir.VirtualMethod) *VirtualMethod {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	e = e.sub("virtual method", v.Name)

	trampoline := fmt.Sprintf("%s_%s_%s", e.trampolinePrefix(), parent.GoType(1), v.Name)

	parentTrampoline := fmt.Sprintf("%s_%s_virtual_%s", e.trampolinePrefix(), parent.GoType(1), v.Name)

	params, _ := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		e.logger.Warn("skipping because parameters are not supported")
		return nil
	}

	// trampoline parameters are inverted, so we need to check them as well
	trampolineParams, _ := NewCallbackParameters(e, v.CallableAttrs)

	if trampolineParams == nil {
		e.logger.Warn("skipping because trampoline parameters are not supported")
		return nil
	}

	if params.InstanceParam == nil {
		e.logger.Warn("skipping because of missing instance parameter")
		return nil
	}

	for _, param := range params.CParameters() {
		if _, ok := param.Type.Type.(*Callback); ok {
			e.logger.Warn("skipping because of callback parameter")
			return nil
		}
	}

	field := findTypeStructField(v, typestruct)

	if field == nil {
		e.logger.Warn("could not find type struct field name")
		return nil
	}

	goname := strcases.SnakeToGo(true, field.CIndentifier())

	return &VirtualMethod{
		Doc:                  NewDoc(&v.InfoAttrs, &v.InfoElements),
		Parent:               parent,
		TrampolineName:       trampoline,
		ParentTrampolineName: parentTrampoline,

		Invoker:    field,
		GoName:     goname,
		ParentName: fmt.Sprintf("Parent%s", goname),
		Parameters: params,
	}
}

func findTypeStructField(virtual *gir.VirtualMethod, ts *Record) *Field {
	name := virtual.Name
	// if virtual.Invoker != "" {
	// 	name = virtual.Invoker
	// }

	for _, field := range ts.Fields {
		if field.CIndentifier() == name {
			return field
		}
	}

	return nil
}
