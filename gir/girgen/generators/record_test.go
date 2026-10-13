package generators

import (
	"io"
	"strings"
	"testing"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

func TestRecordOwnerPointer(t *testing.T) {
	for _, opaque := range []bool{false, true} {
		name := "concrete"
		if opaque {
			name = "opaque"
		}
		t.Run(name, func(t *testing.T) {
			record := &typesystem.Record{
				BaseType: typesystem.BaseType{GoTyp: "Box", CGoTyp: "C.Box"},
				PrivateGoType: "box",
				Opaque: opaque,
				BaseConversions: typesystem.BaseConversions{
					FromGlibBorrowFunction: "UnsafeBoxFromGlibBorrow",
					FromGlibFullFunction: "UnsafeBoxFromGlibFull",
					FromGlibNoneFunction: "UnsafeBoxFromGlibNone",
					ToGlibNoneFunction: "UnsafeBoxToGlibNone",
					ToGlibFullFunction: "UnsafeBoxToGlibFull",
				},
				CgoRefFunction: "C.box_ref",
				CgoUnrefFunction: "C.box_unref",
				GoUnsafeRefFunction: "UnsafeBoxRef",
				GoUnsafeUnrefFunction: "UnsafeBoxUnref",
			}
			generator := &RecordGenerator{
				Doc: &InlineGoDocGenerator{},
				Record: record,
				ReceiverName: "b",
			}
			w := file.NewPackage(t.TempDir(), nil)
			generator.Generate(w)
			code, err := io.ReadAll(w.Go().(io.Reader))
			if err != nil {
				t.Fatal(err)
			}
			want := []string{
				"runtime.SetFinalizer(\n\t\twrapped.box,",
				"runtime.SetFinalizer(b.box, nil)",
				"b.native = nil",
				"C.box_ref((*C.Box)(b.native))",
			}
			if opaque {
				want = append(want, "native unsafe.Pointer", "return (*C.Box)(b.native)", "return &Box{&box{p}}", "C.box_unref((*C.Box)(intern.native))", "C.box_unref((*C.Box)(b.native))")
			} else {
				want = append(want, "native *C.Box", "return b.native", "return &Box{&box{(*C.Box)(p)}}", "C.box_unref(intern.native)", "C.box_unref(b.native)")
			}
			for _, fragment := range want {
				if !strings.Contains(string(code), fragment) {
					t.Errorf("missing generated fragment %q:\n%s", fragment, code)
				}
			}
		})
	}
}
