package glib

import (
	"reflect"
	"testing"
)

func TestBytesMarshal(t *testing.T) {
	bytes := NewBytes([]byte("foobarbaz"))

	gv, err := bytes.ToGValue()

	if err != nil {
		t.Fatal(err)
	}

	marshaledBytesI, err := gv.GoValue()

	if err != nil {
		t.Fatal(err)
	}

	marshaledBytes, ok := marshaledBytesI.(*Bytes)

	if !ok {
		t.Fatal("could not cast")
	}

	if !reflect.DeepEqual(bytes.Data(), marshaledBytes.Data()) {
		t.Fatal("not equal data")
	}
}
