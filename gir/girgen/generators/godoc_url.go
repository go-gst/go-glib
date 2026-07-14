package generators

import (
	"bufio"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

// UrlGodocGenerator generates documentation comments that link to an external url for the
// given [typesystem.Documented] element.
type UrlGodocGenerator struct {
	// GIRDoc is needed for deprecation information. Must not generate other documentation from it, because
	// that could break the license of the original documentation.
	GIRDoc typesystem.Doc

	DocParagraphs []string
}

// Generate implements DocGenerator.
func (g *UrlGodocGenerator) Generate(w file.CodeWriter) {
	// scan the lines of the comment and prefix each line with "// "
	for i, paragraph := range g.DocParagraphs {
		r := strings.NewReader(paragraph)
		scanner := bufio.NewScanner(r) // scan lines

		for scanner.Scan() {
			w.Write([]byte("// "))
			w.Write(scanner.Bytes())
			w.Write([]byte("\n"))
		}

		// add a blank line between paragraphs
		if i < len(g.DocParagraphs)-1 {
			w.Write([]byte("// \n"))
		}
	}

	var zerodoc typesystem.Doc

	if g.GIRDoc == zerodoc {
		return
	}

	// must not write the main doc, only deprecation info

	if !g.GIRDoc.Deprecated {
		return
	}

	w.Write([]byte("//\n"))
	w.Write([]byte("// Deprecated: "))
	if g.GIRDoc.DeprecatedVersion != "" {
		fmt.Fprintf(w, "(since %s) ", g.GIRDoc.DeprecatedVersion)
	}

	w.Write([]byte("see the provided link for the reason\n"))
}

// Copy creates a deep copy of the doc generator.
func (docg *UrlGodocGenerator) Copy() *UrlGodocGenerator {
	newDocg := &UrlGodocGenerator{
		DocParagraphs: make([]string, len(docg.DocParagraphs)),
	}

	copy(newDocg.DocParagraphs, docg.DocParagraphs)

	return newDocg
}

// WithPrependParagraphs implements DocGenerator.
func (g *UrlGodocGenerator) WithPrependParagraphs(paragraphs ...string) DocGenerator {
	newDocg := g.Copy()
	newDocg.DocParagraphs = append(paragraphs, newDocg.DocParagraphs...)
	return newDocg
}

var _ DocGenerator = (*UrlGodocGenerator)(nil)

type MkDocUrlFunc func(namespace *typesystem.Namespace, documented typesystem.Documented) string

type DocGeneratorFactory func(namespace *typesystem.Namespace, documented typesystem.Documented) DocGenerator

// NewGtkGodocGenerator creates a DocGenerator that generates GTK documentation links.
func NewGtkGodocGenerator(namespace *typesystem.Namespace, documented typesystem.Documented) DocGenerator {
	return NewUrlGodocGenerator(mkGtkDocURL, namespace, documented)
}

// NewHotDocGodocGeneratorFactory creates a DocGeneratorFactory that generates GTK documentation links.
func NewHotDocGodocGeneratorFactory(baseUrl string, namemap func(string) string) DocGeneratorFactory {
	return func(namespace *typesystem.Namespace, documented typesystem.Documented) DocGenerator {

		var urlFunc MkDocUrlFunc = func(namespace *typesystem.Namespace, documented typesystem.Documented) string {
			return mkHotDocURL(baseUrl, namespace, documented, namemap)
		}

		return NewUrlGodocGenerator(urlFunc, namespace, documented)
	}
}

func NewUrlGodocGenerator(mkUrl MkDocUrlFunc, namespace *typesystem.Namespace, documented typesystem.Documented) DocGenerator {
	gen := &UrlGodocGenerator{
		GIRDoc: documented.Documentation(),
	}

	if sig, ok := documented.(*typesystem.Signal); ok {
		gen.DocParagraphs = append(gen.DocParagraphs, documentSignal(sig))
	}

	if identifier, ok := documented.(typesystem.Identifier); ok {
		gen.DocParagraphs = append(gen.DocParagraphs, documentIdentifier(identifier))
	}

	if typ, ok := documented.(typesystem.Type); ok {
		gen.DocParagraphs = append(gen.DocParagraphs, documentType(typ))
	}

	gen.DocParagraphs = append(gen.DocParagraphs, fmt.Sprintf("see also %s", mkUrl(namespace, documented)))

	return gen
}

func mkGtkDocURL(namespace *typesystem.Namespace, documented typesystem.Documented) string {
	baseURL := "https://docs.gtk.org"

	nsName := namespace.GoName

	var docPath string
	switch d := documented.(type) {
	case *typesystem.Constant:
		docPath = "const." + d.GirName + ".html"
	case *typesystem.Alias:
		docPath = "alias." + d.GirName + ".html"
	case *typesystem.Class:
		docPath = "class." + d.GirName + ".html"
	case *typesystem.Interface:
		docPath = "interface." + d.GirName + ".html"
	case *typesystem.Enum:
		docPath = "enum." + d.GirName + ".html"
	case *typesystem.Bitfield:
		docPath = "flags." + d.GirName + ".html"
	case *typesystem.Member:
		// members are documented under their parent type and do not have their own page
		parent := d.Parent.GIRName()
		docPath = "flags." + parent + ".html#" + d.GirName

	case *typesystem.Record:
		docPath = "struct." + d.GirName + ".html"
	case *typesystem.CallableSignature:
		switch d.Girtype {
		case typesystem.CallableTypeFunction:
			docPath = "func." + d.CIndentifier() + ".html"
		case typesystem.CallableTypeMethod:
			parent := d.GirCIdentifier
			docPath = "method." + parent + "." + d.CIndentifier() + ".html"
		default:
			panic(fmt.Sprintf("unexpected typesystem.CallableType: %#v", d.Girtype))
		}
	case *typesystem.Callback:
		docPath = "callback." + d.GirName + ".html"
	case *typesystem.VirtualMethod:
		parent := d.Parent.GIRName()

		docPath = "method." + parent + "." + d.Invoker.CIndentifier() + ".html"
	case *typesystem.Signal:
		parent := d.InstanceParam.Type.Type.GIRName()

		docPath = "signal." + parent + "." + d.Name + ".html"
	default:
		fmt.Printf("unhandled type %T for GTK doc URL generation\n", d)
		panic("unsupported documented type for GTK doc URL generation")
	}

	return baseURL + "/" + nsName + "/" + docPath
}

// mkHotDocURL creates a URL to the hotdoc documentation for the given documented element.
func mkHotDocURL(baseUrl string, namespace *typesystem.Namespace, documented typesystem.Documented, namemap func(string) string) string {
	nsName := namemap(namespace.Name)

	docURL, err := url.Parse(baseUrl)

	if err != nil {
		panic("could not parse given base URL")
	}

	docURL.Path, _ = url.JoinPath(docURL.Path, nsName)

	// if the type is a method, then it will be on the same page as the parent definition. Otherwise it has its own page
	switch t := documented.(type) {
	case *typesystem.CallableSignature:
		if parent, ok := t.Parent.(typesystem.Documented); ok {
			file := getHotdocFilename(parent.Documentation())

			docURL.Path, _ = url.JoinPath(docURL.Path, file)

			docURL.Fragment = t.CIndentifier()
		}

	case *typesystem.VirtualMethod:
		if parent, ok := t.Parent.(typesystem.Documented); ok {
			file := getHotdocFilename(parent.Documentation())

			docURL.Path, _ = url.JoinPath(docURL.Path, file)

			docURL.Fragment = t.Invoker.CIndentifier()
		}
	case typesystem.Identifier:
		file := getHotdocFilename(documented.Documentation())
		docURL.Path, _ = url.JoinPath(docURL.Path, file)

		docURL.Fragment = t.CIndentifier()
	case typesystem.Type:
		file := getHotdocFilename(documented.Documentation())
		docURL.Path, _ = url.JoinPath(docURL.Path, file)

		docURL.Fragment = t.CType(0)
	default:
		file := getHotdocFilename(documented.Documentation())
		docURL.Path, _ = url.JoinPath(docURL.Path, file)
	}

	return docURL.String()
}

// getHotdocFilename returns the filename from the documentation that hotdoc uses the filename for the url, e.g.:
// ../subprojects/gstreamer/gst/gstbin.h -> gstbin.html
func getHotdocFilename(doc typesystem.Doc) string {
	if doc.Filename == "" {
		return ""
	}

	// derive hotdoc html filename from header path
	fn := filepath.Base(doc.Filename) // e.g. gstbin.h
	ext := filepath.Ext(fn)           // e.g. .h

	return strings.TrimSuffix(fn, ext) + ".html"
}
