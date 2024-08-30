package glib_test

import (
	"testing"

	"github.com/go-gst/go-glib/glib"
)

func TestArbitraryValue(t *testing.T) {
	// Create a new ArbitraryValue
	av := glib.ArbitraryValue{Data: 42}
	// Convert it to a GValue
	v, err := av.ToGValue()
	if err != nil {
		t.Fatal(err)
	}

	retI, err := v.GoValue()

	if err != nil {
		t.Fatal(err)
	}

	ret, ok := retI.(glib.ArbitraryValue)

	if !ok {
		t.Fatalf("Expected ArbitraryValue, got %T", retI)
	}

	data, ok := ret.Data.(int)

	if !ok {
		t.Fatalf("Expected int, got %T", ret.Data)
	}

	if data != 42 {
		t.Fatalf("Expected 42, got %v", ret.Data)
	}
}
