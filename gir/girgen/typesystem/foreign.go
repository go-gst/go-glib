package typesystem

import (
	"fmt"
	"slices"
	"strings"
)

// CouldBeForeign
type CouldBeForeign[T Type] struct {
	Namespace *Namespace

	Type T
}

func (f CouldBeForeign[T]) WithForeignNamespace(goidentifier string) string {
	if slices.Contains(GoBuiltins, goidentifier) {
		return goidentifier
	}

	if f.Namespace == nil {
		return goidentifier
	}

	pointers := strings.LastIndex(goidentifier, "*")

	ptrStr := goidentifier[0 : pointers+1]
	goidentifier = goidentifier[pointers+1:]

	return fmt.Sprintf("%s%s.%s", ptrStr, f.Namespace.GoName, goidentifier)
}

// GoType is a shorthand method since the go type always needs the foreign namespace
func (f CouldBeForeign[T]) NamespacedGoType(pointers int) string {
	if f.Namespace == nil || isContainerInstance(f.Type) {
		return f.Type.GoType(pointers)
	}

	return f.WithForeignNamespace(f.Type.GoType(pointers))
}

var GoBuiltins = []string{
	"error",
	"any",
	"interface{}",
	"context.Context",
}
