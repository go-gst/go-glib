package gir

import (
	"encoding"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// CommaSeparated is a list of strings that are xml marshaled as a comma-separated
// string attribute
type CommaSeparated []string

// UnmarshalXML unmarshals the comma-separated string into a slice of strings.
func (c *CommaSeparated) UnmarshalXMLAttr(attr xml.Attr) error {
	// Split the string by commas and trim spaces.
	*c = strings.Split(attr.Value, ",")

	return nil
}

// MarshalXMLAttr marshals the slice of strings into a comma-separated string.
func (c CommaSeparated) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	// Join the strings with commas and return as an xml attribute.
	return xml.Attr{
		Name:  name,
		Value: strings.Join(c, ","),
	}, nil
}

func ParseVersion(str string) (Version, error) {
	var v Version

	err := v.UnmarshalText([]byte(str))

	if err != nil {
		return v, err
	}

	return v, nil
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Equals implements equality comparison
func (v Version) Equals(other Version) bool {
	return v.Major == other.Major &&
		v.Minor == other.Minor &&
		v.Patch == other.Patch
}

// Less implements the less than relation
func (v Version) Less(other Version) bool {
	if v.Major < other.Major {
		return true
	}

	if v.Major > other.Major {
		return false
	}

	if v.Minor < other.Minor {
		return true
	}

	if v.Minor > other.Minor {
		return false
	}

	if v.Patch < other.Patch {
		return true
	}

	return false
}

// Greater implements the greater than relation
func (v Version) Greater(other Version) bool {
	if v.Major > other.Major {
		return true
	}

	if v.Major < other.Major {
		return false
	}

	if v.Minor > other.Minor {
		return true
	}

	if v.Minor < other.Minor {
		return false
	}

	if v.Patch > other.Patch {
		return true
	}

	return false
}

// LessEqual implements the less than or equal relation
func (v Version) LessEqual(other Version) bool {
	return v.Less(other) || v.Equals(other)
}

// GreaterEqual implements the greater than or equal relation
func (v Version) GreaterEqual(other Version) bool {
	return v.Greater(other) || v.Equals(other)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (v *Version) UnmarshalText(in []byte) error {
	text := string(in)

	if len(text) == 0 {
		return nil
	}

	parts := strings.Split(text, ".")

	if len(parts) > 3 {
		return fmt.Errorf("invalid version")
	}

	var major int
	var minor int
	var patch int
	var err error

	if len(parts) > 0 {
		major, err = strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf(`error while parsing major version "%s" of "%s": %w`, parts[0], text, err)
		}
	}
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf(`error while parsing minor version "%s" of "%s": %w`, parts[1], text, err)
		}
	}
	if len(parts) > 2 && parts[2] != "" {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf(`error while parsing patch version "%s" of "%s": %w`, parts[2], text, err)
		}
	}

	*v = Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	return nil
}

var _ encoding.TextUnmarshaler = &Version{}

// UnusedAttr is an attribute that always errors when unmarshaled.
// It is used to mark an attribute that is present in the spec but not used as invalid
type UnusedAttr struct{}

func (u UnusedAttr) UnmarshalXMLAttr(attr xml.Attr) error {
	return fmt.Errorf("unused attribute %s", attr.Name.Local)
}
