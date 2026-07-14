package convert

import (
	"github.com/go-gst/go-glib/gir/girgen/file"
)

// Converter descibes a single parameter conversion for a go->C function call.
type Converter interface {
	// Metdata convey additional information about the conversion
	// this is a useful debugging information that will get output into the
	// generated variable declaration
	Metadata() string

	// Convert writes the conversion to the [file.FileI]. It expects that all needed variables are declared
	// before this function is called
	Convert(file.File)
}

type ConverterList []Converter
