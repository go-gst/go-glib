// Package gir provides GIR types, as well as functions that parse GIR files.
//
// For reference, see
// https://gitlab.gnome.org/GNOME/gobject-introspection/-/blob/HEAD/docs/gir-1.2.rnc.
package gir

import (
	"iter"
	"strings"
)

// RawFiles is a map of GIR file names to their raw contents.
type RawFiles map[string][]byte

// ParsedGIRFiles is a map of GIR file names to their parsed contents.
type Repositories map[string]*Repository

// Repository returns the Repository with the given name, or nil if it does not exist.
func (repos Repositories) Repository(name string) *Repository {
	return repos[name]
}

// Find implements Searchable. ns must be a versioned name (like "Gtk-4").
func (repos Repositories) Find(ns string) any {
	// Check if the name is versioned, e.g. "Gtk-4"
	nameParts := strings.SplitN(ns, "-", 2)

	if len(nameParts) != 2 {
		panic("invalid namespace name: " + ns)
	}

	// Versioned name, e.g. "Gtk-4"
	name := nameParts[0]
	versionStr := nameParts[1]

	version, err := ParseVersion(versionStr)

	if err != nil {
		// Invalid version string, return nil
		return nil
	}

	for _, repo := range repos {
		for _, namespace := range repo.Namespaces {
			if namespace.Name == name && namespace.Version.Major == version.Major {
				return namespace
			}
		}
	}

	return nil
}

func (repos Repositories) FindFullType(path string) any {
	var current any
	for part := range subPaths(path) {
		if current == nil {
			// initial search
			current = repos.Find(part)

			if current == nil {
				// no repository found with that name
				return nil
			}
			continue
		}

		if searchable, ok := current.(Searchable); ok {
			current = searchable.Find(part)
			continue
		}

		// we have a path we should search, but the current
		// object is not searchable, so we can't continue.
		return nil
	}

	return current
}

func subPaths(path string) iter.Seq[string] {
	return func(yield func(string) bool) {
		current := path
		for {
			first, rest, _ := strings.Cut(current, ".")

			if first == "" {
				return
			}
			if !yield(first) {
				return
			}
			current = rest
		}
	}
}
