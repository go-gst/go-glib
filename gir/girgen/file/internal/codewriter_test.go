package internal_test

import (
	"io"
	"testing"

	"github.com/go-gst/go-glib/gir/girgen/file/internal"
)

func TestCodeWriter(t *testing.T) {
	var w internal.CodeWriter

	w.Write([]byte("root\n"))
	w.Indent()
	w.Write([]byte("child 1\nchild 2\n"))
	w.Indent()
	w.Write([]byte("grandchild\n"))
	w.Unindent()
	w.Write([]byte("child 3\n"))
	w.Unindent()
	w.Write([]byte("root again\n"))

	expected := "" +
		"root\n" +
		"\tchild 1\n" +
		"\tchild 2\n" +
		"\t\tgrandchild\n" +
		"\tchild 3\n" +
		"root again\n"

	output, _ := io.ReadAll(&w)

	out := string(output)

	if out != expected {
		t.Errorf("unexpected out:\n--- got ---\n%s\n--- want ---\n%s", out, expected)
	}
}

func TestCodeWriterEdgeCases(t *testing.T) {
	var w internal.CodeWriter

	// Initial write
	w.Write([]byte("root\n"))

	// Empty write (should be a no-op)
	w.Write([]byte(""))

	// Indent and write without newline
	w.Indent()
	w.Write([]byte("child 1")) // no newline

	// Empty write again
	w.Write([]byte(""))

	// Now finish the line with newline
	w.Write([]byte("\n"))

	// Write multiple lines in one go
	w.Write([]byte("child 2\nchild 3\n"))

	// Indent again and write a partial line
	w.Indent()
	w.Write([]byte("grandchild A"))

	// Write newline later
	w.Write([]byte("\n"))

	// Another child at same indent level
	w.Write([]byte("grandchild B\n"))

	// Unindent and add more
	w.Unindent()
	w.Write([]byte("child 4\n"))

	// Final unindent to root level
	w.Unindent()
	w.Write([]byte("root again\n"))

	expected := "" +
		"root\n" +
		"\tchild 1\n" +
		"\tchild 2\n" +
		"\tchild 3\n" +
		"\t\tgrandchild A\n" +
		"\t\tgrandchild B\n" +
		"\tchild 4\n" +
		"root again\n"

	output, _ := io.ReadAll(&w)

	out := string(output)

	if out != expected {
		t.Errorf("unexpected out:\n--- got ---\n%s\n--- want ---\n%s", out, expected)
	}
}
