package file

import (
	"fmt"
	"io"
	"slices"

	"github.com/go-gst/go-glib/gir/girgen/file/internal"
)

// cDefines holds all defined C declarations
type cDefines map[string]struct{}

// reader returns an io.Reader that writes the C defines
func (def cDefines) reader() io.Reader {
	if len(def) == 0 {
		return io.MultiReader()
	}

	var declarations []string

	for d := range def {
		declarations = append(declarations, fmt.Sprintf("// #define %s\n", d))
	}

	slices.Sort(declarations)

	var block internal.CodeWriter

	for _, decl := range declarations {
		block.Write([]byte(decl))
	}

	return &block
}
