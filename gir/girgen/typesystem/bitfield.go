package typesystem

import (
	"github.com/go-gst/go-glib/gir"
)

// Bitfield is always a wrapper type for int in go
type Bitfield struct {
	BaseType
	Doc
	Marshaler

	gir *gir.Bitfield

	Members []*Member

	Functions []*CallableSignature
}

func DeclareBitfield(e *env, v *gir.Bitfield) *Bitfield {
	e = e.sub("bitfield", v.CType)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	b := &Bitfield{
		gir: v,
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   e.identifierToGo(v.CType),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		Marshaler: e.newDefaultMarshaler(v.GLibGetType, v.Name),
		Doc:       NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	b.Members = GetMembers(e, b, v.Members)

	return b
}

func (b *Bitfield) declareNested(e *env) {
	for _, v := range b.gir.Functions {
		if t := DeclarePrefixedFunction(e, b, v.CallableAttrs); t != nil {
			b.Functions = append(b.Functions, t)
		}
	}
}

func (a *Bitfield) maxPointersAllowed() int {
	return 0
}
