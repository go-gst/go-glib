package typesystem

import "github.com/go-gst/go-glib/gir"

type Documented interface {
	Documentation() Doc
}

// Doc holds documentation and deprecation info extracted from GIR. Using this text in the
// generated code may be risky because of license issues.
type Doc struct {
	Doc               string
	DocDeprecated     string
	DeprecatedVersion string
	Deprecated        bool

	Filename string
}

func (d Doc) Documentation() Doc {
	return d
}

func NewSimpleDoc(girdoc *gir.Doc) Doc {
	var doc string

	if girdoc != nil {
		doc = girdoc.String
	}

	return Doc{
		Doc: doc,
	}
}

func NewDoc(attrs *gir.InfoAttrs, elements *gir.InfoElements) Doc {
	var doc string
	var docDeprecated string
	var deprecated bool
	var deprecatedVersion string
	var filename string

	if attrs != nil {
		deprecated = attrs.Deprecated
	}

	var zeroversion gir.Version

	if attrs != nil && attrs.DeprecatedVersion != zeroversion {
		deprecatedVersion = attrs.DeprecatedVersion.String()
	}

	if elements != nil && elements.Doc != nil {
		doc = elements.Doc.String
	}

	if elements != nil && elements.SourcePosition != nil {
		filename = elements.SourcePosition.Filename
	}

	if filename == "" && elements.Doc != nil && elements.Doc.Filename != "" {
		filename = elements.Doc.Filename
	}

	if elements != nil && elements.DocDeprecated != nil {
		docDeprecated = elements.DocDeprecated.String
	}

	return Doc{
		Doc:               doc,
		DocDeprecated:     docDeprecated,
		Deprecated:        deprecated,
		DeprecatedVersion: deprecatedVersion,
		Filename:          filename,
	}
}

type ParamDoc struct {
	Doc      string
	Name     string
	Optional bool
	Nullable bool
}

func NewParamDoc(attrs gir.ParameterAttrs) ParamDoc {
	var doc string
	if attrs.Doc != nil {
		doc = attrs.Doc.String
	}
	return ParamDoc{
		Doc:      doc,
		Optional: attrs.Optional,
		Nullable: attrs.Nullable,
		Name:     attrs.Name,
	}
}

func NewReturnDoc(attrs *gir.ReturnValue) ParamDoc {
	return ParamDoc{
		Name:     "",
		Doc:      "",
		Optional: false,
		Nullable: attrs.Nullable,
	}
}
