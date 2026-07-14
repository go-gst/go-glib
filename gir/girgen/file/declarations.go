package file

import (
	"bytes"
	"io"
	"text/tabwriter"
)

// DeclarationWriter uses tabwriter to vertically align the declarations without formatting
type DeclarationWriter struct {
	w          *tabwriter.Writer
	hasWritten bool
	buf        bytes.Buffer
}

// Write implements io.Writer.
func (d *DeclarationWriter) Write(p []byte) (n int, err error) {
	d.hasWritten = true

	if d.w == nil {
		d.w = tabwriter.NewWriter(&d.buf, 0, 0, 1, ' ', 0)
	}

	return d.w.Write(p)
}

func (d *DeclarationWriter) WriteTo(out io.Writer) (int64, error) {
	if !d.hasWritten {
		return 0, nil
	}

	err := d.w.Flush()

	if err != nil {
		return 0, err
	}

	written, err := io.Copy(out, &d.buf)

	d.hasWritten = false
	d.buf.Reset()

	return written, err
}
