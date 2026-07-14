package gir

type Searchable interface {
	// Find returns a definition with the given name or C identifier.
	Find(typ string) any
}

var _ Searchable = &Repositories{}
var _ Searchable = &Namespace{}
var _ Searchable = &Class{}
var _ Searchable = &Interface{}
var _ Searchable = &Record{}
var _ Searchable = &Enum{}
var _ Searchable = &Bitfield{}

// var _ Searchable = Union{} // TODO
