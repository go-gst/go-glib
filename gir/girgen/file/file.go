package file

import (
	"fmt"
	"io"

	"github.com/go-gst/go-glib/gir/girgen/file/internal"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type CodeWriter interface {
	io.Writer
	Indent()
	Unindent()
	NewSection()
}
type File interface {
	GoImportCore(pkg string)
	GoImportNamespace(ns *typesystem.Namespace)
	GoImportType(typ typesystem.CouldBeForeign[typesystem.Type])
	GoImport(pkg string)

	Go() CodeWriter
	C() CodeWriter
}

// file contains the shared logic between the file.go and file_export.go
type file struct {
	cPreamble internal.CodeWriter

	goContents internal.CodeWriter

	currentNs *typesystem.Namespace

	importBaseURIs map[string]string

	goImports goImports
}

var coreglibPkg = "github.com/go-gst/go-glib/pkg/core"

func (d *file) GoImportCore(pkg string) {
	if d.goImports == nil {
		d.goImports = make(goImports)
	}

	d.goImports.add(coreglibPkg+"/"+pkg, "", false)
}

// GoImportType imports either the namespace or the go package contained in the go type: FIXME: how to import nested packages?
func (d *file) GoImportType(typ typesystem.CouldBeForeign[typesystem.Type]) {
	if d.goImports == nil {
		d.goImports = make(goImports)
	}

	if typ.Namespace != nil {
		d.GoImportNamespace(typ.Namespace)
	}

	if typ.Type == nil {
		panic("tried to import nil type")
	}

	alias, module := typ.Type.GoTypeRequiredImport()
	if module == "" {
		return
	}

	d.goImports.add(module, alias, true)
}

func (d *file) GoImportNamespace(ns *typesystem.Namespace) {
	if d.goImports == nil {
		d.goImports = make(goImports)
	}

	if ns == nil {
		return
	}
	if ns == d.currentNs {
		return
	}

	base, ok := d.importBaseURIs[fmt.Sprintf("%s-%d", ns.Name, ns.Version.Major)]

	if !ok {
		panic("tried to import unknown namespace")
	}

	path := base + "/" + ns.GoName

	if ns.Version.Major > 1 {
		path = fmt.Sprintf("%s/v%d", path, ns.Version.Major)
	}

	d.goImports.add(path, "", false)
}

func (d *file) GoImport(pkg string) {
	if d.goImports == nil {
		d.goImports = make(goImports)
	}

	d.goImports.add(pkg, "", true)
}

func (d *file) Go() CodeWriter {
	return &d.goContents
}

func (d *file) C() CodeWriter {
	return &d.cPreamble
}

func (d *file) empty() bool {
	return len(d.goImports) == 0 &&
		d.cPreamble.Len() == 0 &&
		d.goContents.Len() == 0
}

func (d *file) c() io.Reader {
	if d.cPreamble.Len() == 0 {
		return empty
	}

	return internal.NewPrependLinesReader("// ", &d.cPreamble)
}
