package gir

import (
	"encoding/xml"
)

// https://gitlab.gnome.org/GNOME/gobject-introspection/-/blob/HEAD/docs/gir-1.2.rnc

type Namespace struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 namespace"`

	Name    string  `xml:"name,attr"`
	Version Version `xml:"version,attr"`

	// CIdentifierPrefixes contains a list of prefixes that need to be stripped from data structures and types
	CIdentifierPrefixes CommaSeparated `xml:"http://www.gtk.org/introspection/c/1.0 identifier-prefixes,attr"`

	// CSymbolPrefixes contains a list of prefixes that need to be stripped from c functions
	CSymbolPrefixes CommaSeparated `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefixes,attr"`

	// Deprecated: Prefix is deprecated, use CIdentifierPrefixes instead.
	Prefix string `xml:"http://www.gtk.org/introspection/c/1.0 prefix,attr"`

	SharedLibrary string `xml:"shared-library,attr"`

	Aliases     []*Alias      `xml:"http://www.gtk.org/introspection/core/1.0 alias"`
	Classes     []*Class      `xml:"http://www.gtk.org/introspection/core/1.0 class"`
	Interfaces  []*Interface  `xml:"http://www.gtk.org/introspection/core/1.0 interface"`
	Records     []*Record     `xml:"http://www.gtk.org/introspection/core/1.0 record"`
	Enums       []*Enum       `xml:"http://www.gtk.org/introspection/core/1.0 enumeration"`
	Functions   []*Function   `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Unions      []*Union      `xml:"http://www.gtk.org/introspection/core/1.0 union"`
	Bitfields   []*Bitfield   `xml:"http://www.gtk.org/introspection/core/1.0 bitfield"`
	Callbacks   []*Callback   `xml:"http://www.gtk.org/introspection/core/1.0 callback"`
	Constants   []*Constant   `xml:"http://www.gtk.org/introspection/core/1.0 constant"`
	Annotations []*Annotation `xml:"http://www.gtk.org/introspection/core/1.0 attribute"`
	Boxeds      []*Boxed      `xml:"http://www.gtk.org/introspection/core/1.0 boxed"`
}

// Prefixes returns the CIdentifierPrefixes and CSymbolPrefixes to use.
func (ns *Namespace) Prefixes() (cIdentifierPrefixes, cSymbolPrefixes []string) {
	if len(ns.CSymbolPrefixes) == 0 {
		return ns.CIdentifierPrefixes, ns.CIdentifierPrefixes
	}

	return ns.CIdentifierPrefixes, ns.CSymbolPrefixes
}

// Find implements Searchable.
func (n *Namespace) Find(typ string) any {
	for _, alias := range n.Aliases {
		if alias.Name != typ {
			continue
		}

		return alias
	}

	for _, constant := range n.Constants {
		if constant.Name != typ {
			continue
		}

		return constant
	}

	for _, class := range n.Classes {
		if class.Name != typ {
			continue
		}

		return class
	}

	for _, iface := range n.Interfaces {
		if iface.Name != typ {
			continue
		}

		return iface
	}

	for _, record := range n.Records {
		if record.Name != typ {
			continue
		}

		return record
	}

	for _, enum := range n.Enums {
		if enum.Name != typ {
			continue
		}

		return enum
	}

	for i, function := range n.Functions {
		if function.Name != typ {
			continue
		}
		_ = i
		return function
	}

	for _, union := range n.Unions {
		if union.Name != typ {
			continue
		}

		return union
	}

	for _, bitfield := range n.Bitfields {
		if bitfield.Name != typ {
			continue
		}

		return bitfield
	}

	for _, callback := range n.Callbacks {
		if callback.Name != typ {
			continue
		}

		return callback
	}

	// TODO: Boxed

	return nil
}
