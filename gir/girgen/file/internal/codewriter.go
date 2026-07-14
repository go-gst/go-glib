package internal

import (
	"bytes"
	"strings"
)

type CodeWriter struct {
	buf         bytes.Buffer
	indentLevel int
	hasWritten  bool
	atLineStart bool

	pendingNewLine bool
}

const tabStr = "\t"

func (p *CodeWriter) Write(b []byte) (n int, err error) {
	totalWritten := 0
	lines := bytes.SplitAfter(b, []byte("\n"))

	p.atLineStart = !p.hasWritten || p.atLineStart
	p.hasWritten = true

	// If there was a pending newline, write it first
	if p.pendingNewLine {
		p.buf.WriteByte('\n')
		p.pendingNewLine = false
	}

	for _, line := range lines {
		// if the line is shorter than on byte, then it's empty
		if p.atLineStart && p.indentLevel > 0 && len(line) > 1 {
			indent := strings.Repeat(tabStr, p.indentLevel)
			// must not count the bytes we write additionally
			_, err := p.buf.Write([]byte(indent))
			if err != nil {
				return totalWritten, err
			}

			p.atLineStart = false
		}

		written, err := p.buf.Write(line)
		totalWritten += written
		if err != nil {
			return totalWritten, err
		}

		p.atLineStart = p.atLineStart || len(line) > 0 && line[len(line)-1] == '\n'
	}

	return totalWritten, nil
}

func (p *CodeWriter) Read(b []byte) (n int, err error) {
	return p.buf.Read(b)
}

func (p *CodeWriter) Indent() {
	p.indentLevel++
}

func (p *CodeWriter) Unindent() {
	if p.indentLevel == 0 {
		panic("indent mismatch: can't unindent level 0")
	}

	p.indentLevel--
	p.pendingNewLine = false
}

func (p *CodeWriter) Len() int {
	return p.buf.Len()
}

// NewSection schedules a new line, but only writes it if [CodeWriter.Unindent] is not called before.
func (p *CodeWriter) NewSection() {
	p.pendingNewLine = true
}
