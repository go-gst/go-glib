package typesystem

type checkedParameterType interface {
	allowedTypeForParam(P *Param) bool
}

type minPointerConstrainedType interface {
	minPointersRequired() int
}

type maxPointerConstrainedType interface {
	maxPointersAllowed() int
}

func TypeRequiresPointer(t Type) bool {
	if r, ok := t.(minPointerConstrainedType); ok {
		return r.minPointersRequired() > 0
	}

	return false
}

func TypePointersAllowed(t Type, pointers int) bool {
	if r, ok := t.(minPointerConstrainedType); ok && r.minPointersRequired() > pointers {
		return false
	}

	if r, ok := t.(maxPointerConstrainedType); ok && r.maxPointersAllowed() < pointers {
		return false
	}

	return true
}
