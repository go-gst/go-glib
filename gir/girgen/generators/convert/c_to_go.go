package convert

import (
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

func NewCToGoConverter(p *typesystem.Param) Converter {
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
	if p.CallerAllocates {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	if p.IsUserData {
		return &CToGoCallbackConverter{
			Param: p,
		}
	}

	if _, ok := p.Type.Type.(*typesystem.ContainerInstance); ok {
		return newCToGoContainerConverter(p)
	}

	if _, ok := p.Type.Type.(*typesystem.Array); ok {
		return newCToGoArrayConverter(p)
	}

	if p.Nullable {
		if p.Type.Type == typesystem.Utf8 {
			return &CToGoNullableStringConverter{
				Param: p,
				SubConverter: &CToGoStringConverter{
					Param: p,
				},
			}
		}
		return &CToGoNullableConverter{
			Param:        p,
			SubConverter: newCToGoBasicConverter(p),
		}
	}

	return newCToGoBasicConverter(p)
}

func newCToGoBasicConverter(p *typesystem.Param) Converter {
	if p.Type.Type.GoType(0) == "string" {
		return &CToGoStringConverter{Param: p}
	}

	if p.CTypePointers == 0 && p.Type.Type.CType(0) == "_Bool" { // needed for e.g. graphene
		return &CToGoCastingConverter{Param: p}
	}

	if p.CTypePointers == 0 && p.Type.Type.GoType(0) == "bool" {
		return &CToGoBooleanConverter{Param: p}
	}

	switch p.Type.Type.(type) {
	case typesystem.CastableType, *typesystem.Bitfield, *typesystem.Enum:
		if p.CTypePointers == 0 {
			return &CToGoCastingConverter{
				Param: p,
			}
		}
	case *typesystem.Alias:
		return newCToGoAliasedConverter(p)
	}

	conv, ok := p.Type.Type.(typesystem.ConvertibleType)

	if ok && p.CTypePointers == 1 {

		if ok && conv.CanTransfer(typesystem.DirectionCToGo, p.TransferOwnership) {
			return &CToGoConvertibleConverter{
				Param:       p,
				ConvertFunc: p.Type.WithForeignNamespace(conv.GetTransferFromGlibFunction(p.TransferOwnership)),
			}
		}
	}

	return &UnimplementedConverter{Param: p}
}

// tricky sub cases here, depending on the aliased type
func newCToGoAliasedConverter(p *typesystem.Param) Converter {
	subtype := p.Type.Type.(*typesystem.Alias).AliasedType

	if _, ok := subtype.Type.(typesystem.CastableType); ok && p.CTypePointers == 0 {
		return &AliasConverter{
			SubConverter: &CToGoCastingConverter{
				Param: p,
			},
		}
	}

	return &UnimplementedConverter{Param: p}
}

func newCToGoArrayConverter(p *typesystem.Param) Converter {
	// array := p.Type.Type.(*typesystem.Array)

	return &UnimplementedConverter{
		Param: p,
	}
}
