package file

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file/internal"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

// externCallbacks holds all extern C declarations of callbacks.
type externCallbacks map[*typesystem.Callback]struct{}

// reader returns an io.Reader that declares the extern C trampoline functions to be used in the C preamble of the
// generated file
func (cbs externCallbacks) reader() io.Reader {
	if len(cbs) == 0 {
		return io.MultiReader()
	}

	var declarations []string

	for cb := range cbs {
		var params []string

		for _, p := range cb.CParameters() {
			params = append(params, p.CType())
		}

		declarations = append(declarations, fmt.Sprintf("// extern %s %s(%s);\n", cb.CReturn.CType(), cb.TrampolineName, strings.Join(params, ", ")))
	}

	slices.Sort(declarations)

	var block internal.CodeWriter

	for _, decl := range declarations {
		block.Write([]byte(decl))
	}

	return &block
}
