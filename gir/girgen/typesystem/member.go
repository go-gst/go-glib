package typesystem

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

type Member struct {
	Doc
	Identifier
	GirName string
	Parent  Type
	Value   string
}

func GetMembers(e *env, parent Type, ms []*gir.Member) []*Member {
	var mems []*Member

	for _, girM := range ms {
		if m := NewMember(e, parent, girM); m != nil {
			mems = append(mems, m)
		}
	}

	return mems
}

func NewMember(e *env, parent Type, m *gir.Member) *Member {
	if !m.IsIntrospectable() {
		return nil
	}

	if e.skip(parent, m) {
		return nil
	}

	return &Member{
		Doc:     NewDoc(&m.InfoAttrs, &m.InfoElements),
		Parent:  parent,
		GirName: m.Name(),
		Identifier: &baseIdentifier{
			cIndentifier: m.CIdentifier,
			//
			goIndentifier:  e.identifierToGo(strcases.SnakeToGo(true, strings.ToLower(m.CIdentifier))),
			cGoIndentifier: "C." + m.CIdentifier,
		},
		Value: valueToInt32(m.Value),
	}
}

// valueToInt32 parses the value as an int64 integer and casts the value to
// int32, overflowing it if necessary. If the value is not a valid integer, this panics.
//
// This is needed because we generate int32 values for the enum members, but
// the C values may sometimes exceed the int32 range. This is not a problem in C, but
// it is in Go, so we need to cast the value to int32.
func valueToInt32(value string) string {
	v, err := strconv.ParseInt(value, 0, 64)

	if err != nil {
		panic("value is not a valid integer")
	}

	return fmt.Sprintf("%d", int32(v))
}

type Members []*Member

func (m Members) Uniques() []*Member {
	unique := make(map[string]*Member)

	for _, mem := range m {
		if _, ok := unique[mem.Value]; !ok {
			unique[mem.Value] = mem
		}
	}

	var uniques []*Member

	for _, mem := range unique {
		uniques = append(uniques, mem)
	}

	// need to sort because the map ordering above is not stable:
	slices.SortFunc(uniques, func(i, j *Member) int {
		return cmp.Compare(i.GoIndentifier(), j.GoIndentifier())
	})

	return uniques
}
