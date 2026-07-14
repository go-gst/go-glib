package generators

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

type DocGenerator interface {
	Generate(w file.CodeWriter)
	WithPrependParagraphs(paragraphs ...string) DocGenerator
}

// InlineGoDocGenerator generates the Go documentation as inline comments. This may violate the
// license of the original documentation, so be careful when using this.
type InlineGoDocGenerator struct {
	DocParagraphs []string
	GIRDoc        typesystem.Doc
}

func (docg *InlineGoDocGenerator) Generate(w file.CodeWriter) {
	// scan the lines of the comment and prefix each line with "// "
	for i, paragraph := range docg.DocParagraphs {
		r := strings.NewReader(paragraph)
		scanner := bufio.NewScanner(r) // scan lines

		for scanner.Scan() {
			w.Write([]byte("// "))
			w.Write(scanner.Bytes())
			w.Write([]byte("\n"))
		}

		// add a blank line between paragraphs
		if i < len(docg.DocParagraphs)-1 {
			w.Write([]byte("// \n"))
		}
	}

	var zerodoc typesystem.Doc

	if docg.GIRDoc == zerodoc {
		return
	}

	w.Write([]byte("//\n"))

	// scan the lines of the comment and prefix each line with "// "
	r := strings.NewReader(docg.GIRDoc.Doc)
	scanner := bufio.NewScanner(r) // scan lines

	for scanner.Scan() {
		w.Write([]byte("// "))
		w.Write(scanner.Bytes())
		w.Write([]byte("\n"))
	}

	if docg.GIRDoc.Deprecated {
		w.Write([]byte("//\n"))
		w.Write([]byte("// Deprecated: "))
		if docg.GIRDoc.DeprecatedVersion != "" {
			fmt.Fprintf(w, "(since %s) ", docg.GIRDoc.DeprecatedVersion)
		}

		r := strings.NewReader(docg.GIRDoc.DocDeprecated)
		scanner := bufio.NewScanner(r) // scan lines

		scanner.Scan() // initial deprecated line is already prefied

		fmt.Fprintf(w, "%s\n", scanner.Bytes())

		for scanner.Scan() {
			w.Write([]byte("// "))
			w.Write(scanner.Bytes())
			w.Write([]byte("\n"))
		}
	}
}

// NewInlineGoDocGenerator creates a new InlineGoDocGenerator for the given documented element.
// This generator will include the documentation directly in the generated Go code as comments.
// This may violate the license of the original documentation, so be careful when using this.
func NewInlineGoDocGenerator(_ *typesystem.Namespace, documented typesystem.Documented) DocGenerator {
	gen := &InlineGoDocGenerator{
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

	if callable, ok := documented.(typesystem.Callable); ok {
		gen.DocParagraphs = append(gen.DocParagraphs, documentCallable(callable)...)
	}

	return gen
}

// Copy creates a deep copy of the doc generator.
func (docg *InlineGoDocGenerator) Copy() *InlineGoDocGenerator {
	newDocg := &InlineGoDocGenerator{
		DocParagraphs: make([]string, len(docg.DocParagraphs)),
		GIRDoc:        docg.GIRDoc,
	}

	copy(newDocg.DocParagraphs, docg.DocParagraphs)

	return newDocg
}

// WithPrependParagraph prepends a paragraph to the doc generator without modifying the instance.
func (docg *InlineGoDocGenerator) WithPrependParagraphs(paragraphs ...string) DocGenerator {
	newDocg := docg.Copy()
	newDocg.DocParagraphs = append(paragraphs, newDocg.DocParagraphs...)
	return newDocg
}

func documentSignal(sig *typesystem.Signal) string {
	if sig.Action {
		return fmt.Sprintf("%s emits the \"%s\" signal", sig.GoName, sig.Name)
	} else {
		return fmt.Sprintf("%s connects the provided callback to the \"%s\" signal", sig.GoName, sig.Name)
	}
}

func documentIdentifier(identifier typesystem.Identifier) string {
	return fmt.Sprintf("%s wraps %s", identifier.GoIndentifier(), identifier.CIndentifier())
}

func documentType(typ typesystem.Type) string {
	gotype := typ.GoType(0)

	switch t := typ.(type) {
	case *typesystem.Class:
		gotype = t.GoInterfaceName
	case *typesystem.Interface:
		gotype = t.GoInterfaceName
	}

	return fmt.Sprintf("%s wraps %s", gotype, typ.CType(0))
}

func documentCallable(callable typesystem.Callable) []string {
	var docParagraphs []string

	params := callable.CallableParameters()

	if len(params.GoParameters) > 0 {
		docParagraphs = append(docParagraphs, "The function takes the following parameters:\n")

		var paramDoc strings.Builder

		for _, param := range params.GoParameters {
			if param.Skip || param.Implicit {
				continue
			}
			fmt.Fprintf(&paramDoc, "\t- %s \n", paramDocListItem(param))
		}

		docParagraphs = append(docParagraphs, paramDoc.String())
	}

	if len(params.GoReturns) > 0 {
		docParagraphs = append(docParagraphs, "The function returns the following values:\n")

		var returnDoc strings.Builder

		for _, rv := range params.GoReturns {
			fmt.Fprintf(&returnDoc, "\t- %s \n", paramDocListItem(rv))
		}

		docParagraphs = append(docParagraphs, returnDoc.String())
	}

	return docParagraphs
}

func paramDocListItem(p *typesystem.Param) string {
	docStr := fmt.Sprintf("%s %s", p.GoName, p.GoType())

	if p.Nullable {
		docStr += " (nullable)"
	}

	// if p.Optional { // out params may be optional, but we always use them as go returns
	// 	docStr = " (optional)"
	// }

	if p.Doc.Doc != "" {
		docStr += ": "
		docStr += p.Doc.Doc
	}

	return docStr
}
