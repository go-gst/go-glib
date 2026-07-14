package gir

import (
	"encoding/xml"
	"fmt"
)

// https://gitlab.gnome.org/GNOME/gobject-introspection/-/blob/HEAD/docs/gir-1.2.rnc

// Repository represents a GObject Introspection Repository, which contains the
// includes, C includes and namespaces of a single gir file.
type Repository struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 repository"`
	Version Version  `xml:"version,attr"`

	Includes   []*Include   `xml:"http://www.gtk.org/introspection/core/1.0 include"`
	CIncludes  []*CInclude  `xml:"http://www.gtk.org/introspection/c/1.0 include"`
	Packages   []*Package   `xml:"http://www.gtk.org/introspection/core/1.0 package"`
	Namespaces []*Namespace `xml:"http://www.gtk.org/introspection/core/1.0 namespace"`
}

// ParseAll parses the given XML data into a Repository.
func ParseAll(raw RawFiles) (Repositories, error) {
	repos := make(Repositories)
	for name, data := range raw {
		var repo Repository
		if err := xml.Unmarshal(data, &repo); err != nil {
			return nil, fmt.Errorf("failed to parse gir XML for %s: %w", name, err)
		}
		repos[name] = &repo
	}

	return repos, nil
}
