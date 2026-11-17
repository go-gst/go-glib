package gobject_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-gst/go-glib/pkg/gio/v2"
	"github.com/go-gst/go-glib/pkg/gobject/v2"
)

func TestTrySetProperty(t *testing.T) {
	action := gio.NewSimpleAction("test", nil)
	stream := gio.NewBufferedInputStream(gio.NewMemoryInputStream())

	if err := action.TrySetProperty("enabled", false); err != nil {
		t.Fatalf("setting a writable property failed: %v", err)
	}
	if v := action.ObjectProperty("enabled"); v != false {
		t.Errorf("enabled is %v, want false", v)
	}

	if err := stream.TrySetProperty("buffer-size", int32(8)); err != nil {
		t.Fatalf("setting a value GLib can convert failed: %v", err)
	}
	if v := stream.ObjectProperty("buffer-size"); v != uint32(8) {
		t.Errorf("buffer-size is %v (%T), want 8", v, v)
	}

	errs := []struct {
		name, want string
		set        func() error
	}{
		{"missing", "has no property", func() error { return action.TrySetProperty("no-such-property", true) }},
		{"not writable", "not writable", func() error { return action.TrySetProperty("state-type", nil) }},
		{"construct-only", "only be set at construction", func() error { return action.TrySetProperty("name", "other") }},
		{"wrong type", "takes gboolean, not gchararray", func() error { return action.TrySetProperty("enabled", "yes") }},
		{"go int", "int is not supported", func() error { return stream.TrySetProperty("buffer-size", 8) }},
		{"out of range", "is invalid", func() error { return stream.TrySetProperty("buffer-size", uint32(0)) }},
	}
	for _, e := range errs {
		if err := e.set(); err == nil || !strings.Contains(err.Error(), e.want) {
			t.Errorf("%s: got error %v, want one containing %q", e.name, err, e.want)
		}
	}

	if v := action.ObjectProperty("name"); v != "test" {
		t.Errorf("name is %v, want test", v)
	}
	if v := stream.ObjectProperty("buffer-size"); v != uint32(8) {
		t.Errorf("buffer-size is %v after failed sets, want 8", v)
	}
}

func TestFindProperty(t *testing.T) {
	action := gio.NewSimpleAction("test", nil)

	pspec := action.FindProperty("enabled")
	if pspec == nil {
		t.Fatal("FindProperty returned nil for an existing property")
	}
	if pspec.Name() != "enabled" || pspec.ValueType() != gobject.TypeBoolean {
		t.Errorf("got %s of type %s, want enabled of type gboolean", pspec.Name(), pspec.ValueType().Name())
	}
	if pspec.Flags()&gobject.ParamWritable == 0 {
		t.Error("enabled is not reported as writable")
	}
	if action.FindProperty("no-such-property") != nil {
		t.Error("FindProperty returned a spec for a missing property")
	}
}

func TestFindPropertyKeepsClassReference(t *testing.T) {
	action := gio.NewSimpleAction("test", nil)

	for range 10 {
		action.FindProperty("enabled")
		runtime.GC()
		time.Sleep(time.Millisecond)
	}

	if pspec := action.FindProperty("enabled"); pspec == nil || pspec.Name() != "enabled" {
		t.Fatal("the class lost its property spec after the returned specs were collected")
	}
}
