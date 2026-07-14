package typesystem

import "fmt"

// Marshalable describes a type that needs to be registered as a Gvalue marshaler
type Marshalable interface {
	Type
	CanMarshal() bool
	GLibGetType() string

	// Value is used to know from where we need to import GValue for the marshaler/SetValue method
	Value() CouldBeForeign[*Record]

	// Type is used to know from where we need to import Type for the marshaler/SetValue method
	Type() CouldBeForeign[*Alias]

	// GoTypeName returns the name of the variable that holds the GType
	GoTypeName() string
}

type Marshaler struct {
	GlibGetType string
	Gotypename  string

	Gvalue CouldBeForeign[*Record]
	Gtype  CouldBeForeign[*Alias]
}

// Type implements Marshalable.
func (m Marshaler) Type() CouldBeForeign[*Alias] {
	return m.Gtype
}

// CanMarshal implements Marshalable.
func (m Marshaler) CanMarshal() bool {
	return m.GlibGetType != "" && m.Gvalue.Type != nil
}

// GLibGetType implements Marshalable.
func (m Marshaler) GLibGetType() string {
	return m.GlibGetType
}

// GoTypeName implements Marshalable.
func (m Marshaler) GoTypeName() string {
	return m.Gotypename
}

// Value implements Marshalable.
func (m Marshaler) Value() CouldBeForeign[*Record] {
	return m.Gvalue
}

func (e *env) newDefaultMarshaler(gettype string, name string) Marshaler {
	if gettype == "" {
		e.logger.Warn("skipping marshaler because not get type function was provided")
		return Marshaler{}
	}

	valuens, valuetyp := e.findTypeByGIRName("GObject.Value")

	if valuetyp == nil {
		e.logger.Warn("skipping marshaler because gvalue was not found")
		return Marshaler{}
	}

	typens, typetyp := e.findTypeByGIRName("GType")

	if typetyp == nil {
		e.logger.Warn("skipping marshaler because GType was not found")
		return Marshaler{}
	}

	return Marshaler{
		GlibGetType: gettype,
		Gotypename:  fmt.Sprintf("Type%s", name),
		Gvalue: CouldBeForeign[*Record]{
			Namespace: valuens,
			Type:      valuetyp.(*Record),
		},
		Gtype: CouldBeForeign[*Alias]{
			Namespace: typens,
			Type:      typetyp.(*Alias),
		},
	}
}
