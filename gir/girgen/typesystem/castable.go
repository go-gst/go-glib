package typesystem

type CastableType interface {
	Type

	// canBeCasted is a marker method to indicate that this type can be casted
	// between C and Go.
	canBeCasted()
}
