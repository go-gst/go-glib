package internal

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type prependLinesReader struct {
	prefix []byte
	s      *bufio.Scanner
	b      bytes.Buffer
}

// NewPrependLinesReader returns a reader that prepends every line from r with the prefix
func NewPrependLinesReader(linePrefix string, r io.Reader) io.Reader {
	return &prependLinesReader{
		prefix: []byte(linePrefix),
		s:      bufio.NewScanner(r),
	}
}

var ErrBufTooShort = errors.New("buffer to small for prefix")

func (pr *prependLinesReader) Read(b []byte) (int, error) {
	if pr.b.Len() >= len(b) {
		return pr.b.Read(b)
	}

	if pr.s.Scan() {
		pr.b.Write(pr.prefix)
		pr.b.Write(pr.s.Bytes())
		pr.b.Write([]byte("\n"))

		return pr.b.Read(b)
	}

	return pr.b.Read(b)
}
