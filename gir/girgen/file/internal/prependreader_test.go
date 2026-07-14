package internal_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-gst/go-glib/gir/girgen/file/internal"
)

func Test(t *testing.T) {
	r := strings.NewReader("foo\nbar\nbaz\nbam")

	pr := internal.NewPrependLinesReader("// ", r)

	out, _ := io.ReadAll(pr)

	fmt.Println(string(out))
}
