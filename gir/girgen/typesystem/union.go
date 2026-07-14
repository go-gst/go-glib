package typesystem

import (
	"github.com/go-gst/go-glib/gir"
)

// TODO: does union need to implement pointer constraint interfaces?
type Union struct {
	BaseType
	Marshaler
	gir *gir.Union

	Doc Doc

	GetType string

	Functions    []*CallableSignature
	Methods      []*CallableSignature
	Constructors []*CallableSignature
	Fields       []*Field
}

func DeclareUnion(e *env, v *gir.Union) *Union {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	return &Union{
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		GetType: v.GLibGetType,
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		Marshaler: e.newDefaultMarshaler(v.GLibGetType, v.Name),
		gir:       v,
	}
}

func (u *Union) declareNested(e *env) resolvedState {
	for _, v := range u.gir.Functions {
		if t := DeclarePrefixedFunction(e, u, v.CallableAttrs); t != nil {
			u.Functions = append(u.Functions, t)
		}
	}

	for _, v := range u.gir.Methods {
		if t := DeclareMethod(e, u, v); t != nil {
			u.Methods = append(u.Methods, t)
		}
	}

	for _, v := range u.gir.Constructors {
		if t := DeclarePrefixedFunction(e, u, v.CallableAttrs); t != nil {
			u.Constructors = append(u.Constructors, t)
		}
	}

	for _, v := range u.gir.Fields {
		if t := NewField(e, u, v); t != nil {
			u.Fields = append(u.Fields, t)
		}
	}

	// for range v.Records {
	// 	TODO
	// }

	return okResolved
}
