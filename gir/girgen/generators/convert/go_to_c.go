package convert

import (
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

func NewGoToCConverter(p *typesystem.Param) Converter {
	if p.Implicit {
		return &ImplicitConverter{
			Param: p,
		}
	}
	if p.Skip {
		return &SkippedConverter{
			Param: p,
		}
	}

	if _, ok := p.Type.Type.(*typesystem.Array); ok {
		return newGoToCArrayConverter(p)
	}

	if p.CallerAllocates {
		return &UnimplementedConverter{
			Param:  p,
			Reason: "caller-allocates not implemented",
		}
	}

	// out params may be nullable, but we don't care about that
	if p.Direction != "out" && p.Nullable {
		if p.Type.Type.GoType(0) == "string" {
			return &GoToCNullableStringConverter{
				Param: p,
				SubConverter: &GoToCStringConverter{
					Param: p,
				},
			}
		}
		return &GoToCNullableConverter{
			Param:        p,
			SubConverter: newGoToCBasicConverter(p),
		}
	}

	return newGoToCBasicConverter(p)
}

func newGoToCBasicConverter(p *typesystem.Param) Converter {

	if p.Type.Type.GoType(0) == "string" {
		return &GoToCStringConverter{Param: p}
	}

	if p.CTypePointers == 0 && p.Type.Type.GoType(0) == "bool" {
		return &GoToCBooleanConverter{Param: p}
	}

	switch p.Type.Type.(type) {
	case typesystem.CastableType, *typesystem.Bitfield, *typesystem.Enum:
		if p.CTypePointers == 0 {
			return &GoToCCastingConverter{
				Param: p,
			}
		}
	case *typesystem.Callback:
		return &GoToCCallbackConverter{
			Param: p,
		}
	case *typesystem.Alias:
		return newGoToCAliasedConverter(p)
	}

	conv, ok := p.Type.Type.(typesystem.ConvertibleType)

	if ok && p.CTypePointers == 1 {

		if ok && conv.CanTransfer(typesystem.DirectionGoToC, p.TransferOwnership) {
			return &GoToCConvertibleConverter{
				Param:       p,
				ConvertFunc: p.Type.WithForeignNamespace(conv.GetTransferToGlibFunction(p.TransferOwnership)),
			}
		}
	}

	return &UnimplementedConverter{
		Param:  p,
		Reason: "no basic converter found",
	}
}

// tricky sub cases here, depending on the aliased type
func newGoToCAliasedConverter(p *typesystem.Param) Converter {
	subtype := p.Type.Type.(*typesystem.Alias).AliasedType

	if _, ok := subtype.Type.(typesystem.CastableType); ok && p.CTypePointers == 0 {
		return &AliasConverter{
			SubConverter: &GoToCCastingConverter{
				Param: p,
			},
		}
	}

	return &UnimplementedConverter{
		Param:  p,
		Reason: "unsupported aliased type",
	}
}

func newGoToCArrayConverter(p *typesystem.Param) Converter {
	array := p.Type.Type.(*typesystem.Array)

	if array.InnerPointers > 0 {
		return &UnimplementedConverter{
			Param:  p,
			Reason: "unimplemented: inner pointers in array",
		}
	}

	if array.FixedSize == 0 || array.Length != nil {
		return &UnimplementedConverter{
			Param:  p,
			Reason: "unimplemented: non-fixed size array",
		}
	}

	if array.ZeroTerminated {
		return &UnimplementedConverter{
			Param:  p,
			Reason: "unimplemented: zero-terminated array",
		}
	}

	if p.Direction == "out" {
		return &UnimplementedConverter{
			Param:  p,
			Reason: "unimplemented: out array",
		}
	}

	switch array.Inner.Type.(type) {
	case typesystem.CastableType, *typesystem.Bitfield, *typesystem.Enum:
		if array.InnerPointers == 0 {
			return &GoToCFixedSizeArrayConvertibleConverter{
				Param: p,
			}
		}
	}

	return &UnimplementedConverter{
		Param:  p,
		Reason: "unsupported array inner type",
	}
}
