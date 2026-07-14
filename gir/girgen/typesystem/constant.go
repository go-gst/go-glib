package typesystem

import (
	"github.com/go-gst/go-glib/gir"
)

type Constant struct {
	Doc
	Identifier

	GirName string
	GoValue string
}

func DeclareConstant(e *env, v *gir.Constant) *Constant {
	e = e.sub("constant", v.CType)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	// for some reason the name of the constant is listed under the ctype
	cIdentifier := v.CType

	ns, underlying := e.findType(&v.Type)

	if ns != nil {
		e.logger.Warn("skipping foreign constant")
		return nil
	}

	if underlying == nil {
		return nil
	}

	if _, ok := underlying.(*CastablePrimitive); !ok {
		e.logger.Warn("skipping because not a castable primitive")
		return nil
	}

	return &Constant{
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		GirName: v.Name,
		Identifier: &baseIdentifier{
			// goIdentifier must be used directly, because e.g. gdk defines GDK_KEY_Armenian_at and GDK_KEY_Armenian_AT
			// which break when we transform to PascalCase
			goIndentifier:  v.Name,
			cIndentifier:   cIdentifier,
			cGoIndentifier: "C." + cIdentifier,
		},
		GoValue: v.Value,
	}
}
