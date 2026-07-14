package typesystem

type Identifier interface {
	CIndentifier() string
	CGoIndentifier() string
	GoIndentifier() string
}

type baseIdentifier struct {
	cIndentifier   string
	cGoIndentifier string
	goIndentifier  string
}

// CGoIndentifier implements Identifier.
func (b *baseIdentifier) CGoIndentifier() string {
	return b.cGoIndentifier
}

// CIndentifier implements Identifier.
func (b *baseIdentifier) CIndentifier() string {
	return b.cIndentifier
}

// GoIndentifier implements Identifier.
func (b *baseIdentifier) GoIndentifier() string {
	return b.goIndentifier
}

var _ Identifier = &baseIdentifier{}
