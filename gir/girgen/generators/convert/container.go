package convert

import (
	"fmt"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

func newCToGoContainerConverter(p *typesystem.Param) Converter {
	container := p.Type.Type.(*typesystem.ContainerInstance)

	// convertFunc is the conversion function for the container itself.
	convertFunc := p.Type.WithForeignNamespace(container.FromGlibNoneFunction)

	if p.TransferOwnership == typesystem.TransferFull ||
		p.TransferOwnership == typesystem.TransferContainer {
		convertFunc = p.Type.WithForeignNamespace(container.FromGlibFullFunction)
	}

	// innerTransfer is the conversion direction for the inner type.
	innerTransfer := typesystem.TransferNone

	if p.TransferOwnership == typesystem.TransferFull {
		innerTransfer = typesystem.TransferFull
	}

	childConverters := make([]Converter, 0, len(container.InnerTypes))

	for _, inner := range container.InnerTypes {
		if conv, ok := inner.Type.(typesystem.ConvertibleType); ok && conv.CanTransfer(typesystem.DirectionCToGo, innerTransfer) {

			childConverters = append(childConverters, &CToGoContainerChildConvertibleConverter{
				ConvertFunc: inner.WithForeignNamespace(conv.GetTransferFromGlibFunction(innerTransfer)),
			})

			continue
		}

		switch inner.Type {
		case typesystem.Utf8, typesystem.Filename:
			childConverters = append(childConverters, &CToGoContainerChildStringConverter{
				Transfer: innerTransfer,
			})
			continue
		}

		break
	}

	if len(childConverters) != len(container.InnerTypes) {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	return &CToGoContainerConverter{
		Container:       container,
		ConvertFunc:     convertFunc,
		Param:           p,
		ChildConverters: childConverters,
	}
}

type CToGoContainerConverter struct {
	Container   *typesystem.ContainerInstance
	ConvertFunc string
	Param       *typesystem.Param

	ChildConverters []Converter
}

// Convert implements Converter.
func (c *CToGoContainerConverter) Convert(f file.File) {
	f.GoImport("unsafe")
	fmt.Fprintf(f.Go(), "%s = %s(\n", c.Param.GoName, c.ConvertFunc)
	f.Go().Indent()
	fmt.Fprintf(f.Go(), "unsafe.Pointer(%s),\n", c.Param.CName)

	for i, conv := range c.ChildConverters {
		inner := c.Container.InnerTypes[i]
		// must always be 1 pointer for containers
		innerType := inner.NamespacedGoType(1)
		fmt.Fprintf(f.Go(), "func(v unsafe.Pointer) %s {\n", innerType)
		f.Go().Indent()
		fmt.Fprintf(f.Go(), "var dst %s // %s\n", innerType, conv.Metadata())
		conv.Convert(f)
		fmt.Fprintf(f.Go(), "return dst\n")
		f.Go().Unindent()
		fmt.Fprintf(f.Go(), "},\n")
	}

	f.Go().Unindent()
	fmt.Fprintf(f.Go(), ")\n")
}

// Metadata implements Converter.
func (c *CToGoContainerConverter) Metadata() string {
	return fmt.Sprintf("container, transfer: %s", c.Param.TransferOwnership)
}

var _ Converter = &CToGoContainerConverter{}

type CToGoContainerChildStringConverter struct {
	// Transfer is the transfer mode of the child, not the container
	Transfer typesystem.TransferOwnership
}

// Convert implements Converter.
func (c *CToGoContainerChildStringConverter) Convert(f file.File) {
	fmt.Fprintf(f.Go(), "dst = C.GoString((*C.char)(v))\n")

	if c.Transfer == typesystem.TransferFull {
		fmt.Fprintf(f.Go(), "defer C.free(v)\n")
	}
}

// Metadata implements Converter.
func (c *CToGoContainerChildStringConverter) Metadata() string {
	return "string"
}

var _ Converter = &CToGoContainerChildStringConverter{}

type CToGoContainerChildConvertibleConverter struct {
	ConvertFunc string
}

// Convert implements Converter.
func (c *CToGoContainerChildConvertibleConverter) Convert(f file.File) {
	fmt.Fprintf(f.Go(), "dst = %s(v)\n", c.ConvertFunc)
}

// Metadata implements Converter.
func (c *CToGoContainerChildConvertibleConverter) Metadata() string {
	return "converted"
}

var _ Converter = &CToGoContainerChildConvertibleConverter{}
