package typesystem

type resolvedState int

const (
	// notResolvable marks the type as invalid
	notResolvable resolvedState = iota
	// maybeResolvable tells the resolver to try again after other types
	// have been resolved
	maybeResolvable
	// okResolved means that the type is valid
	okResolved
)
