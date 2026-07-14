package typesystem

import "github.com/go-gst/go-glib/gir"

// Container is a Type that must be manually implemented. It has a [ContainerInstance] counterpart that
// represents a single instance of this type. Containers are used to represent generic data structures
// see https://docs.gtk.org/glib/data-structures.html
//
// Containers are not converted go to C but instead copy the contents to a go datastructure.
type Container struct {
	GirName string
	C       string
	CGo     string

	// FromGlibFullFunction will also be used for transfer "container", where
	// the container is owned but not the contents.
	FromGlibFullFunction string
	FromGlibNoneFunction string

	MakeGoType func([]CouldBeForeign[Type]) string

	NumInnerTypes int
}

// allowedTypeForParam implements checkedParameterType.
func (c *Container) allowedTypeForParam(p *Param) bool {
	// containers are not allowed as parameters
	// because they are not converted to C
	return p.Direction == "return"
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (c *Container) maxPointersAllowed() int {
	return 1
}

// minPointersRequired implements minPointerConstrainedType.
func (c *Container) minPointersRequired() int {
	return 1
}

// CGoType implements Type.
func (c *Container) CGoType(pointers int) string {
	return GetPointers(pointers) + c.CGo
}

// CType implements Type.
func (c *Container) CType(pointers int) string {
	return c.C + GetPointers(pointers)
}

// GIRName implements Type.
func (c *Container) GIRName() string {
	return c.GirName
}

// GoType implements Type.
func (c *Container) GoType(pointers int) string {
	panic("tried to use a container as an instance")
}

// GoTypeRequiredImport implements Type.
func (c *Container) GoTypeRequiredImport() (alias string, module string) {
	return "", ""
}

var _ Type = (*Container)(nil)
var _ minPointerConstrainedType = (*Container)(nil)
var _ maxPointerConstrainedType = (*Container)(nil)
var _ checkedParameterType = (*Container)(nil)

type ContainerInstance struct {
	*Container

	InnerTypes []CouldBeForeign[Type]
}

// GoType implements Type.
func (c *ContainerInstance) GoType(pointers int) string {
	return c.MakeGoType(c.InnerTypes)
}

func (e *env) resolveContainerInnerTypes(c *Container, inner []*gir.Type) Type {
	if len(inner) != c.NumInnerTypes {
		e.logger.Warn("container inner type count is mismatched", "type", c.GirName, "inner-types", inner, "desired", c.NumInnerTypes)
		return nil
	}

	innerTypes := make([]CouldBeForeign[Type], 0, len(inner))

	for _, t := range inner {
		ns, typ := e.findOuterType(t)

		if typ == nil {
			return nil
		}

		if typ == Gpointer {
			e.logger.Warn("container inner type is unsafe Pointer", "type", c.GirName, "inner-type", t.Name)
			return nil
		}

		innerTypes = append(innerTypes, CouldBeForeign[Type]{Namespace: ns, Type: typ})
	}

	return &ContainerInstance{
		Container:  c,
		InnerTypes: innerTypes,
	}
}

func isContainerInstance(t Type) bool {
	_, ok := t.(*ContainerInstance)
	return ok
}
