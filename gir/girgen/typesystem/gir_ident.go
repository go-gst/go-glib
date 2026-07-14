package typesystem

import (
	"fmt"
	"strings"
)

const GIRWildCard = "*"

type GIRIdentifier struct {
	Parent string
	Name   string
}

func (id GIRIdentifier) Matches(parent string, self string) bool {
	if id.Name == GIRWildCard {
		return parent == id.Parent
	}

	return parent == id.Parent && self == id.Name
}

func GIRPattern(str string) GIRIdentifier {
	parts := strings.Split(str, ".")

	switch len(parts) {
	case 1:
		return GIRIdentifier{
			Parent: "",
			Name:   parts[0],
		}
	case 2:
		return GIRIdentifier{
			Parent: parts[0],
			Name:   parts[1],
		}
	default:
		panic("invalid GIR identifier")
	}
}

func (id GIRIdentifier) String() string {
	if id.Parent == "" {
		return id.Name
	}
	// Dot notation: Parent.Name
	return fmt.Sprintf("%s.%s", id.Parent, id.Name)
}
