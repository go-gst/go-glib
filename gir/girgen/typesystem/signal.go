package typesystem

import (
	"fmt"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

type SignalWhen string

const (
	SignalWhenFirst   = "first"
	SignalWhenLast    = "last"
	SignalWhenCleanup = "cleanup"
)

type Signal struct {
	Doc
	Name      string
	GoName    string
	Detailed  bool
	When      SignalWhen
	Action    bool
	NoHooks   bool
	NoRecurse bool

	// GObject is used to resolve the gobject namespace for the signal handler.
	GObject CouldBeForeign[Type]

	InstanceParam *Param
	GirParameters ParamList
	Return        *Param
}

func NewSignal(e *env, parent Type, v *gir.Signal) *Signal {
	e = e.sub("signal", v.Name)

	if e.skip(parent, v) {
		return nil
	}

	goNamePrefix := "Connect"

	if v.Action {
		goNamePrefix = "Emit"
	}

	ns, obj := e.findTypeByGIRName("GObject.Object")

	if obj == nil {
		e.logger.Warn("GObject.Object not found")
		return nil
	}

	s := &Signal{
		Doc:       NewDoc(nil, &v.InfoElements),
		Name:      v.Name,
		GoName:    goNamePrefix + strcases.KebabToGo(true, v.Name),
		Detailed:  v.Detailed,
		When:      SignalWhen(v.When),
		Action:    v.Action,
		NoHooks:   v.NoHooks,
		NoRecurse: v.NoRecurse,

		GObject: CouldBeForeign[Type]{
			Namespace: ns,
			Type:      obj,
		},

		InstanceParam: &Param{
			Doc:    NewParamDoc(gir.ParameterAttrs{}),
			CName:  "// invalid identifier",
			GoName: "instance",
			Type: CouldBeForeign[Type]{
				Namespace: nil,
				Type:      parent,
			},
			CTypePointers: 1, // always uses a pointer
		},
	}

	// Signal params often don't specify a Ctype, especially for records and classes etc. So we
	// cannot use the normal CallableParameter resolution here

	if v.Parameters != nil {
		for i, param := range v.Parameters.Parameters {
			ns, t := e.findAnyType(param.AnyType)

			if t == nil {
				e.logger.Warn("type not found", "type", debugCTypeFromAnytype(param.AnyType))
				return nil
			}

			pointers := CountCTypePointers(CTypeFromAnytype(param.AnyType))

			if pointers == 0 {
				if res, ok := t.(minPointerConstrainedType); ok {
					pointers = res.minPointersRequired()
				}
			}

			s.GirParameters = append(s.GirParameters, &Param{
				Doc:    NewParamDoc(param.ParameterAttrs),
				GoName: fmt.Sprintf("arg%d", i),
				CName:  "// invalid identifier",
				Type: CouldBeForeign[Type]{
					Namespace: ns,
					Type:      t,
				},
				CTypePointers: pointers,
			})
		}
	}

	if v.ReturnValue != nil {
		ns, t := e.findAnyType(v.ReturnValue.AnyType)

		if t == nil {
			e.logger.Warn("type not found", "type", debugCTypeFromAnytype(v.ReturnValue.AnyType))
			return nil
		}

		pointers := CountCTypePointers(CTypeFromAnytype(v.ReturnValue.AnyType))

		s.Return = &Param{
			Doc:    NewParamDoc(gir.ParameterAttrs{}),
			GoName: "// invalid identifier",
			CName:  "// invalid identifier",
			Type: CouldBeForeign[Type]{
				Namespace: ns,
				Type:      t,
			},
			CTypePointers: pointers,
		}
	}

	return s
}

// Parameters returns the parameters that need to be generated. For action signals
// this is omits the instance parameter, because it is already in the receiver.
// For non-action signals, it returns all parameters.
func (s *Signal) Parameters() ParamList {
	if s.Action {
		return s.GirParameters
	}

	// For non-action signals, we need to return the instance parameter as well.

	params := make(ParamList, len(s.GirParameters)+1)

	params[0] = s.InstanceParam
	copy(params[1:], s.GirParameters)
	return params
}

func (s *Signal) GoReturn() *Param {
	if s.Return == nil {
		return nil
	}

	if s.Return.Type.Type == Void {
		return nil
	}

	return s.Return
}
