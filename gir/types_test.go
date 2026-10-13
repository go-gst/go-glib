package gir_test

import (
	"encoding/xml"
	"testing"

	"github.com/go-gst/go-glib/gir"
)

func TestVersionLessEqual(t *testing.T) {
	tests := []struct {
		v1       gir.Version
		v2       gir.Version
		expected bool
	}{
		{gir.Version{1, 0, 0}, gir.Version{2, 0, 0}, true},  // Major version difference
		{gir.Version{2, 0, 0}, gir.Version{1, 0, 0}, false}, // Major version difference
		{gir.Version{1, 2, 0}, gir.Version{1, 3, 0}, true},  // Minor version difference
		{gir.Version{1, 3, 0}, gir.Version{1, 2, 0}, false}, // Minor version difference
		{gir.Version{1, 2, 3}, gir.Version{1, 2, 4}, true},  // Patch version difference
		{gir.Version{1, 2, 4}, gir.Version{1, 2, 3}, false}, // Patch version difference
		{gir.Version{1, 2, 3}, gir.Version{1, 2, 3}, true},  // Equal versions
	}

	for _, tt := range tests {
		result := tt.v1.LessEqual(tt.v2)
		if result != tt.expected {
			t.Errorf("Expected %v.LessEqual(%v) to be %v, but got %v", tt.v1, tt.v2, tt.expected, result)
		}
	}
}

func TestRecordOpaque(t *testing.T) {
	var record gir.Record
	err := xml.Unmarshal([]byte(`<record xmlns="http://www.gtk.org/introspection/core/1.0" name="Sample" opaque="1"/>`), &record)
	if err != nil {
		t.Fatal(err)
	}
	if !record.Opaque {
		t.Fatal("GIR opaque attribute was not preserved")
	}
}
