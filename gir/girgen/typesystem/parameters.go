package typesystem

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

type CallbackParamScope string

const (
	CallbackParamScopeCall  CallbackParamScope = "call"
	CallbackParamScopeAsync CallbackParamScope = "async"

	// CallbackParamScopeNotified must be accompanied by a Destroy parameter.
	CallbackParamScopeNotified CallbackParamScope = "notified"
	CallbackParamScopeForever  CallbackParamScope = "forever"
)

type TransferOwnership string

const (
	TransferNone      TransferOwnership = "none"
	TransferFull      TransferOwnership = "full"
	TransferBorrow    TransferOwnership = "borrow"
	TransferContainer TransferOwnership = "container"
)

// NewManualParam constructs a param for ease of use with manual type declarations.
func NewManualParam(cname, goname string, typ Type, pointers int) *Param {
	return &Param{
		CName:  cname,
		GoName: goname,
		Type: CouldBeForeign[Type]{
			Type: typ,
		},
		CTypePointers:     pointers,
		TransferOwnership: TransferNone,
		Skip:              false,
		Optional:          false,
		Nullable:          false,
		Direction:         "in",
	}
}

type Param struct {
	Doc ParamDoc

	CName  string
	GoName string

	// Type is the type of the parameter. It is missing a pointer if this param is an out param that isn't CallerAllocates
	Type CouldBeForeign[Type]
	// CTypePointers contains the amount of "*" characters present in the original ctype. it is used to output the correct type,
	// but only if the Type itself doesn't handle it differently
	CTypePointers int

	GirCType   string
	GirCGoType string

	// Skip signifies that the parameter should be skipped in the go call
	// this happens with params that are only useful in C.
	Skip bool

	TransferOwnership TransferOwnership
	// Nullable means that NULL can be returned
	Nullable  bool
	Direction string
	Scope     CallbackParamScope

	// CallerAllocates means that the out param allocation must be provided by the caller.
	//
	// in practise this means that we need one more pointer if the param direction is out and this is false
	CallerAllocates bool

	// Implicit declares that this param is referenced by another param, either through closure, destroy or
	// array size. It will be omitted in the go call, because it will get it's value from another source.
	Implicit bool

	// IsUserData declares that this param carries the user data for a callback, which will be used to retrieve the go closure from
	// the closure registry in the trampoline function. This is only true for params in a callback.
	IsUserData bool

	// BorrowFrom is a reference to the (instance) param that this wants to borrow from. In code generation terms this means
	// that we need to connect this (return) param to the instance param so that the GC wont clean it up early
	BorrowFrom *Param

	// Closure is a pointer to the implicit param that takes the go function pointer that will be called from the trampoline function.
	// If this is non nil then the param is is the callback arg.
	Closure *Param

	// Destroy is a pointer to the implicit param of the destroy notify callback.
	Destroy *Param

	// Optional signifies that an out or inout param can be NULL to ignore it. This is not useful
	// for moving the out params to go return values, so this is only here for completeness sake
	Optional bool
}

func (p *Param) CDeclaration() string {
	return fmt.Sprintf("%s %s", p.CName, p.CType())
}

func (p *Param) CGoDeclaration() string {
	return fmt.Sprintf("%s %s", p.CName, p.CGoType())
}

func (p *Param) GoDeclaration() string {
	return fmt.Sprintf("%s %s", p.GoName, p.GoType())
}

func (p *Param) CGoType() string {
	if p.GirCGoType != "" {
		return p.GirCGoType
	}

	return p.Type.Type.CGoType(p.CTypePointers)
}

func (p *Param) GoType() string {
	return p.Type.NamespacedGoType(p.CTypePointers)
}

func (p *Param) CType() string {
	if p.GirCType != "" {
		return p.GirCType
	}

	return p.Type.Type.CType(p.CTypePointers)
}

type Callable interface {
	CallableParameters() *Parameters
}

type Parameters struct {
	Doc

	// CReturn contains the param that the c function returns
	CReturn *Param

	// GIRParameters contains the params that the gir declares. The actual CParameters need the InstanceParam prepended, hence this is private
	//
	// the parameter references (destroy, closure and array length) are relative indices in this list
	GIRParameters ParamList

	// InstanceParam contains the C instance param, which will be used as a method receiver
	// for the go function. It is also always the first parameter for the c function call
	InstanceParam *Param

	// GoReturns containts the return values of the Go function. C Params that are declared as "out"
	// will also get moved here, so this may differ from Parameters, but must contain pointers to the same
	// objects.
	GoReturns ParamList

	// GoParameters will contain the parameters of the go function. C Params that are declared as "out"
	// will not be in this list.
	GoParameters ParamList
}

// CallableParameters implements the Callable interface
func (p *Parameters) CallableParameters() *Parameters {
	return p
}

// ParameterMode is used to determine what conversion is needed for an "in" or "out" parameter, because
// they mean different things for a go->c call and a c->go call.
type ParameterMode int

const (
	// ParameterModeCallable is used for functions and methods, aka go->c calls.
	ParameterModeCallable ParameterMode = iota
	// ParameterModeCallback is used for callbacks, aka c->go calls.
	ParameterModeCallback
)

// Invert returns the inverted mode, useful for return parameters, which are always in the opposite direction of the call.
func (p ParameterMode) Invert() ParameterMode {
	switch p {
	case ParameterModeCallable:
		return ParameterModeCallback
	case ParameterModeCallback:
		return ParameterModeCallable
	default:
		panic("invalid ParameterMode")
	}
}

// validForCallable checks if the param is valid for a function or method, aka go->c call.
func (param *Param) valid(e *env, mode ParameterMode) bool {
	if param.Implicit || param.Skip {
		return true
	}

	if conv, ok := param.Type.Type.(MaybeTransferableType); ok {
		switch param.Direction {
		case "inout":
			panic("should not be inout")
		case "in":
			if mode == ParameterModeCallable && !conv.CanTransfer(DirectionGoToC, param.TransferOwnership) {
				e.logger.Warn("transfer ownership not valid for type", "type", param.Type.Type.GIRName(), "transfer", param.TransferOwnership)
				return false
			}
			if mode == ParameterModeCallback && !conv.CanTransfer(DirectionCToGo, param.TransferOwnership) {
				e.logger.Warn("transfer ownership not valid for type", "type", param.Type.Type.GIRName(), "transfer", param.TransferOwnership)
				return false
			}
		case "out", "return":
			if mode == ParameterModeCallable && !conv.CanTransfer(DirectionCToGo, param.TransferOwnership) {
				e.logger.Warn("transfer ownership not valid for type", "type", param.Type.Type.GIRName(), "transfer", param.TransferOwnership)
				return false
			}
			if mode == ParameterModeCallback && !conv.CanTransfer(DirectionGoToC, param.TransferOwnership) {
				e.logger.Warn("transfer ownership not valid for type", "type", param.Type.Type.GIRName(), "transfer", param.TransferOwnership)
				return false
			}
		}
	}

	if param.Type.Type == Gpointer || param.Type.Type == Guintptr || param.Type.Type == Gconstpointer {
		e.logger.Warn("unsafe pointer is not a valid param type", "ctype", param.Type.Type.GIRName(), "gotype", param.GoType(), "ctype", param.CType())
		return false
	}

	switch t := param.Type.Type.(type) {
	case checkedParameterType:
		return t.allowedTypeForParam(param)
	default:
		return TypePointersAllowed(t, param.CTypePointers)
	}
}

// CParameters returns the param list for the c call, since the instance param is always the first param if set
func (p *Parameters) CParameters() ParamList {
	if p.InstanceParam == nil {
		return p.GIRParameters
	}

	params := make(ParamList, 0, len(p.GIRParameters)+1)

	params = append(params, p.InstanceParam)
	params = append(params, p.GIRParameters...)

	return params
}

// CGoReturn returns the CReturn if it is set and not void/none
func (p *Parameters) CGoReturn() *Param {
	if p.CReturn == nil {
		return nil
	}

	if p.CReturn.Type.Type == Void {
		return nil
	}

	return p.CReturn
}

func NewCallableParameters(e *env, v *gir.CallableAttrs) (*Parameters, resolvedState) {
	return NewGenericParameters(e, v, ParameterModeCallable)
}

func NewCallbackParameters(e *env, v *gir.CallableAttrs) (*Parameters, resolvedState) {
	return NewGenericParameters(e, v, ParameterModeCallback)
}

func NewGenericParameters(e *env, v *gir.CallableAttrs, mode ParameterMode) (*Parameters, resolvedState) {
	params := &Parameters{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	if v.Parameters != nil {
		if v.Parameters.InstanceParameter != nil {
			// instance param must not be an array:

			girType := v.Parameters.InstanceParameter.AnyType.Type

			if girType == nil {
				e.logger.Warn("array instance param", "ctype", debugCTypeFromAnytype(v.Parameters.InstanceParameter.AnyType))
				return nil, notResolvable
			}

			ns, t := e.findType(girType)

			if t == nil {
				e.logger.Warn("instance param type not found", "ctype", girType.CType)
				return nil, notResolvable
			}

			if ns != nil {
				e.logger.Warn("foreign instance param", "ctype", girType.CType)
				return nil, notResolvable
			}

			pointers := CountCTypePointers(girType.CType)

			if pointers != 1 {
				e.logger.Warn("instance param without exactly one pointer", "ctype", girType.CType)
				return nil, notResolvable
			}

			params.InstanceParam = &Param{
				Doc: NewParamDoc(v.Parameters.InstanceParameter.ParameterAttrs),

				CName:  "carg0",
				GoName: strcases.ParamNameToGo(v.Parameters.InstanceParameter.Name),
				Type: CouldBeForeign[Type]{
					Namespace: ns,
					Type:      t,
				},
				GirCType:          CTypeFromAnytype(v.Parameters.InstanceParameter.AnyType),
				GirCGoType:        CtypeToCgoType(CTypeFromAnytype(v.Parameters.InstanceParameter.AnyType)),
				CTypePointers:     1,
				TransferOwnership: TransferNone,
				Skip:              false,
				Optional:          false,
				Nullable:          false,
				Direction:         "in",
				Scope:             "call",
				CallerAllocates:   false,
				Implicit:          false,
				Closure:           nil,
				Destroy:           nil,
			}
		}

		for i, p := range v.Parameters.Parameters {
			if p.Direction == "inout" {
				e.logger.Warn("FIXME: skipping inout param")
				return nil, notResolvable
			}

			paramType := p.AnyType

			ctypePointers := CountCTypePointers(CTypeFromAnytype(paramType))

			if p.Direction == "out" {
				ctypePointers = ctypePointers - 1

				newType, ok := decreaseAnyTypePointers(paramType)
				if !ok {
					e.logger.Warn("skipping param not valid for an out direction", "ctype", debugCTypeFromAnytype(paramType))
					return nil, notResolvable
				}

				paramType = newType
			}

			if ctypePointers < 0 {
				e.logger.Warn("skipping param not valid for an out direction", "ctype", debugCTypeFromAnytype(paramType))
				return nil, notResolvable
			}

			ns, t := e.findAnyType(paramType)

			if t == nil {
				// e.logger.Warn("type not found", "ctype", debugCTypeFromAnytype(p.AnyType))
				return nil, maybeResolvable
			}

			direction := p.Direction

			if direction == "" {
				direction = "in"
			}

			scope := CallbackParamScope(p.Scope)

			if scope == "" {
				// https://gi.readthedocs.io/en/latest/annotations/giannotations.html
				scope = CallbackParamScopeCall
			}

			// https://gi.readthedocs.io/en/latest/annotations/giannotations.html#default-annotations
			transfer := TransferOwnership(p.TransferOwnership.TransferOwnership)

			if transfer == "" {
				switch direction {
				case "in":
					transfer = TransferFull
				case "out", "inout":
					if p.CallerAllocates {
						transfer = TransferNone
					}
				}
			}

			nullable := p.Nullable

			if p.Direction == "out" && (p.Optional || p.Nullable) && ctypePointers == 0 && !isPointer(t) {
				// when the out param is converted to a value, and that valu has no pointers,
				// then it cannot be nil
				nullable = false
			}

			param := &Param{
				Doc:    NewParamDoc(p.ParameterAttrs),
				CName:  fmt.Sprintf("carg%d", i+1),
				GoName: strcases.ParamNameToGo(p.Name),
				Type: CouldBeForeign[Type]{
					Namespace: ns,
					Type:      t,
				},
				CTypePointers:     ctypePointers,
				GirCType:          CTypeFromAnytype(paramType),
				GirCGoType:        CtypeToCgoType(CTypeFromAnytype(paramType)),
				TransferOwnership: transfer,
				Skip:              p.Skip,
				Optional:          p.Optional,
				Nullable:          nullable,
				Direction:         direction,
				Scope:             scope,
				CallerAllocates:   p.CallerAllocates,
				Implicit:          false,
				Closure:           nil,
				Destroy:           nil,
				BorrowFrom:        nil,
			}

			params.GIRParameters = append(params.GIRParameters, param)

			if !param.Skip {
				if p.Direction == "out" {
					// value will be a return, must add to go params still to keep the index intact for skipping
					params.GoReturns = append(params.GoReturns, param)
				} else {
					params.GoParameters = append(params.GoParameters, param)
				}
			}
		}

		// mark the implicit params. The idx is the index in c parameters, with a given instance param
		// a parameter may have multiple implicit params
		for i, p := range v.Parameters.Parameters {
			param := params.GIRParameters[i]
			if p.Closure != nil && *p.Closure != i {
				// param is the function pointer arg and p.Closure points to the userdata arg
				param.Closure = params.GIRParameters[*p.Closure]
				param.Closure.Implicit = true
			}
			if p.Closure != nil && *p.Closure == i {
				// we are in a callback and param is the userdata arg
				param.IsUserData = true
				param.Implicit = true
				param.GoName = "_"
			}
			if p.Destroy != nil {
				param.Destroy = params.GIRParameters[*p.Destroy]
				param.Destroy.Implicit = true
			}
			if p.AnyType.Array != nil && p.AnyType.Array.Length != nil {
				param.Type.Type.(*Array).Length = params.GIRParameters[*p.Array.Length]

				params.GIRParameters[*p.Array.Length].Implicit = true
			}
		}
	}

	if v.ReturnValue != nil {
		ns, t := e.findAnyType(v.ReturnValue.AnyType)

		if t == nil {
			e.logger.Warn("return type not found", "ctype", debugCTypeFromAnytype(v.ReturnValue.AnyType))
			return nil, maybeResolvable
		}

		// https://gi.readthedocs.io/en/latest/annotations/giannotations.html#default-annotations
		transfer := TransferOwnership(v.ReturnValue.TransferOwnership.TransferOwnership)

		if transfer == "" {
			transfer = TransferFull
		}

		ctypePointers := CountCTypePointers(CTypeFromAnytype(v.ReturnValue.AnyType))

		ret := &Param{
			Doc:       NewReturnDoc(v.ReturnValue),
			CName:     "cret",
			GoName:    "goret",
			Direction: "return",
			Type: CouldBeForeign[Type]{
				Namespace: ns,
				Type:      t,
			},
			TransferOwnership: transfer,
			GirCType:          CTypeFromAnytype(v.ReturnValue.AnyType),
			GirCGoType:        CtypeToCgoType(CTypeFromAnytype(v.ReturnValue.AnyType)),
			CTypePointers:     ctypePointers,
			Nullable:          v.ReturnValue.Nullable,
		}

		if transfer == TransferBorrow {
			if params.InstanceParam == nil {
				e.logger.Error("can't borrow without an instance param")
				return nil, notResolvable
			}

			ret.BorrowFrom = params.InstanceParam
		}

		params.CReturn = ret

		if t.GIRName() != "none" {
			params.GoReturns = append(params.GoReturns, ret)
		}

	}

	if v.Throws {
		// if a callable throws then it has a GError** as the last param, which is a nullable
		// out param
		ns, throwType := e.findTypeByGIRName("GLib.Error")

		if throwType == nil {
			e.logger.Warn("GLib.Error type not found, ignoring throwing function")
			return nil, notResolvable
		}

		throwParam := &Param{
			Doc: ParamDoc{
				Name: "err",
				Doc:  "an error",
			},
			CName:  "_cerr",
			GoName: "_goerr",
			Type: CouldBeForeign[Type]{
				Namespace: ns,
				Type:      throwType,
			},
			CTypePointers:     1,
			Skip:              false,
			TransferOwnership: TransferFull,
			Optional:          true,
			Nullable:          true,
			Direction:         "out",
		}

		params.GIRParameters = append(params.GIRParameters, throwParam)
		params.GoReturns = append(params.GoReturns, throwParam)
	}

	for _, p := range params.CParameters() {
		if !p.valid(e, mode) {
			e.logger.Error("param not valid, can't be resolved", "param", p.CName)
			return nil, notResolvable
		}
	}

	if params.CReturn != nil && !params.CReturn.valid(e, mode) {
		e.logger.Error("return param not valid, can't be resolved")
		return nil, notResolvable
	}

	e.sortGoParams(params.GoParameters)
	e.sortGoReturns(params.GoReturns)

	return params, okResolved
}

type ParamList []*Param

func (pl ParamList) GoDeclarations() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoDeclaration())
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) GoIdentifiers() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoName)
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) CIdentifiers() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.CName)
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) GoTypes() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoType())
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) CGoDeclarations() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.CGoDeclaration())
	}

	return strings.Join(decls, ", ")
}
