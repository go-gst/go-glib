package typesystem

import (
	"strings"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/strcases"
)

var specialCallableNames = map[string]string{
	// automatically implement Stringer interface
	"to_string": "String",
}

type CallableType string

const (
	CallableTypeFunction CallableType = "function"
	CallableTypeMethod   CallableType = "method"
)

// CallableIdentifier is an identifier that prefixes the parent type, so that renaming the
// parent struct reflects to renaming the constructors / methods. It has a special case for
// constructors, which are prefixed with "New" and the parent type name.
type CallableIdentifier struct {
	ParentForPrefix Type
	Girname         string
	Girtype         CallableType
	GirCIdentifier  string
}

// CGoIndentifier implements Identifier.
func (c *CallableIdentifier) CGoIndentifier() string {
	return "C." + c.GirCIdentifier
}

// CIndentifier implements Identifier.
func (c *CallableIdentifier) CIndentifier() string {
	return c.GirCIdentifier
}

// GoIndentifier turns the girname into a hopefully unique function name
//
// e.g. BufferList.new_sized -> NewBufferListSized
func (c *CallableIdentifier) GoIndentifier() string {
	girname := c.Girname

	if specialName, ok := specialCallableNames[c.Girname]; ok {
		girname = specialName
	}

	pascal := strcases.SnakeToGo(true, girname)

	pascal, hasNewPrefix := strings.CutPrefix(pascal, "New")
	pascal, hasNewSuffix := strings.CutSuffix(pascal, "New")

	var parentTypeName string

	if c.ParentForPrefix != nil {
		if c.ParentForPrefix.CType(0) == "GstAudioFormatInfo" {
			println("foo")
		}

		switch p := c.ParentForPrefix.(type) {
		case *Class:
			parentTypeName = p.GoInterfaceName
		case *Interface:
			parentTypeName = p.GoInterfaceName
		default:
			parentTypeName = c.ParentForPrefix.GoType(0)
		}
	}

	if hasNewPrefix || hasNewSuffix {
		return "New" + parentTypeName + pascal
	}

	return parentTypeName + pascal
}

var _ Identifier = &CallableIdentifier{}

type CallableSignature struct {
	*CallableIdentifier
	*Parameters

	// Parent is the parent type
	Parent Type
}

func DeclareFunction(e *env, v *gir.CallableAttrs) *CallableSignature {
	e = e.sub("function", v.CIdentifier)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if v.ShadowedBy != "" {
		e.logger.Warn("skipping because shadowed", "by", v.ShadowedBy)
		return nil
	}

	if v.MovedTo != "" {
		e.logger.Warn("skipping because moved", "to", v.MovedTo)
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	params, _ := NewCallableParameters(e, v)

	if params == nil {
		return nil
	}

	return &CallableSignature{
		CallableIdentifier: &CallableIdentifier{
			ParentForPrefix: nil,
			Girname:         v.Name,
			Girtype:         CallableTypeFunction,
			GirCIdentifier:  v.CIdentifier,
		},
		Parameters: params,
	}
}

func DeclarePrefixedFunction(e *env, parent Type, v *gir.CallableAttrs) *CallableSignature {
	e = e.sub("function", v.CIdentifier)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if v.ShadowedBy != "" {
		e.logger.Warn("skipping because shadowed", "by", v.ShadowedBy)
		return nil
	}

	if v.MovedTo != "" {
		e.logger.Warn("skipping because moved", "to", v.MovedTo)
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params, _ := NewCallableParameters(e, v)

	if params == nil {
		return nil
	}

	return &CallableSignature{
		CallableIdentifier: &CallableIdentifier{
			ParentForPrefix: parent,
			Girname:         v.Name,
			Girtype:         CallableTypeFunction,
			GirCIdentifier:  v.CIdentifier,
		},
		Parameters: params,
		Parent:     parent,
	}
}

func DeclareMethod(e *env, parent Type, v *gir.Method) *CallableSignature {
	e = e.sub("method", v.CIdentifier)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if v.ShadowedBy != "" {
		e.logger.Warn("skipping because shadowed", "by", v.ShadowedBy)
		return nil
	}

	if v.MovedTo != "" {
		e.logger.Warn("skipping because moved", "to", v.MovedTo)
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params, _ := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		return nil
	}

	return &CallableSignature{
		CallableIdentifier: &CallableIdentifier{
			// Methods are scoped to the parent, so we don't need a parent prefix
			ParentForPrefix: nil,
			Girname:         v.Name,
			Girtype:         CallableTypeMethod,
			GirCIdentifier:  v.CIdentifier,
		},
		Parameters: params,
		Parent:     parent,
	}
}
