package gir

import (
	"encoding/xml"
)

// https://gitlab.gnome.org/GNOME/gobject-introspection/-/blob/HEAD/docs/gir-1.2.rnc

type Alias struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 alias"`

	Name  string `xml:"name,attr"`
	CType string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`

	InfoAttrs
	InfoElements

	Type Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`
}

type AnyType struct {
	// Possible variants.
	Type  *Type  `xml:"http://www.gtk.org/introspection/core/1.0 type"`
	Array *Array `xml:"http://www.gtk.org/introspection/core/1.0 array"`
}

type Annotation struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 attribute"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type Array struct {
	XMLName        xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 array"`
	Name           string   `xml:"name,attr"`
	CType          string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	Length         *int     `xml:"length,attr"` // ix of .Parameters
	ZeroTerminated *bool    `xml:"zero-terminated,attr"`
	FixedSize      int      `xml:"fixed-size,attr"`
	Introspectable bool     `xml:"introspectable,attr"`
	// Type is the array's inner type.
	Type *Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`
}

// IsZeroTerminated returns true if the Array is zero-terminated. It accounts
// for edge cases of the structure.
func (a Array) IsZeroTerminated() bool {
	return a.Name == "" && (a.ZeroTerminated == nil || *a.ZeroTerminated)
}

type Bitfield struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 bitfield"`

	Name         string `xml:"name,attr"` // Go case
	CType        string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	GLibTypeName string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType  string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`

	Members   []*Member   `xml:"http://www.gtk.org/introspection/core/1.0 member"`
	Functions []*Function `xml:"http://www.gtk.org/introspection/core/1.0 function"`

	InfoAttrs
	InfoElements
}

// Find implements Searchable.
func (b Bitfield) Find(typ string) any {
	for _, member := range b.Members {
		if member.Name() == typ || member.CIdentifier == typ || member.GLibName() == typ {
			return member
		}
	}

	for _, function := range b.Functions {
		if function.Name == typ {
			return function
		}
	}

	return nil
}

type Boxed struct{}

type CInclude struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/c/1.0 include"`
	Name    string   `xml:"name,attr"`
}

type CallableAttrs struct {
	Name        string       `xml:"name,attr"`
	CIdentifier string       `xml:"http://www.gtk.org/introspection/c/1.0 identifier,attr"`
	ShadowedBy  string       `xml:"shadowed-by,attr"`
	Shadows     string       `xml:"shadows,attr"`
	Throws      bool         `xml:"throws,attr"`
	MovedTo     string       `xml:"moved-to,attr"`
	Parameters  *Parameters  `xml:"http://www.gtk.org/introspection/core/1.0 parameters"`
	ReturnValue *ReturnValue `xml:"http://www.gtk.org/introspection/core/1.0 return-value"`
	InfoAttrs
	InfoElements
}

func (c *CallableAttrs) FindParameter(name string) *Parameter {
	for _, param := range c.Parameters.Parameters {
		if param.Name == name {
			return param
		}
	}

	return nil
}

type Callback struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 callback"`
	CType   string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	*CallableAttrs
}

type Class struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 class"`
	Name     string   `xml:"name,attr"`
	Parent   string   `xml:"parent,attr"`
	Abstract bool     `xml:"abstract,attr"`

	CType         string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CSymbolPrefix string `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`

	GLibTypeName   string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType    string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	GLibTypeStruct string `xml:"http://www.gtk.org/introspection/glib/1.0 type-struct,attr"`

	InfoAttrs
	InfoElements

	Functions      []*Function      `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Implements     []*Implements    `xml:"http://www.gtk.org/introspection/core/1.0 implements"`
	Constructors   []*Constructor   `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	Methods        []*Method        `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	VirtualMethods []*VirtualMethod `xml:"http://www.gtk.org/introspection/core/1.0 virtual-method"`
	Fields         []*Field         `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Signals        []*Signal        `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`
}

// Find implements Searchable.
func (c Class) Find(ident string) any {
	for _, method := range c.Methods {
		if method.Name == ident {
			return method
		}
	}

	for _, virtualMethod := range c.VirtualMethods {
		if virtualMethod.Name == ident {
			return virtualMethod
		}
	}

	for _, function := range c.Functions {
		if function.Name == ident {
			return function
		}
	}

	for _, signal := range c.Signals {
		if signal.Name == ident {
			return signal
		}
	}

	for _, field := range c.Fields {
		if field.Name == ident {
			return field
		}
	}

	for _, conststructor := range c.Constructors {
		if conststructor.Name == ident {
			return conststructor
		}
	}

	return nil
}

type Constant struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 constant"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
	CType   string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`

	Type Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`

	InfoAttrs
	InfoElements
}

type Constructor struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	*CallableAttrs
}

type Doc struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 doc"`
	Filename string   `xml:"filename,attr"`
	String   string   `xml:",innerxml"`
	Line     int      `xml:"line,attr"`
}

type DocDeprecated struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 doc-deprecated"`
	String  string   `xml:",innerxml"`
}

type DocElements struct {
	Doc            *Doc
	DocDeprecated  *DocDeprecated
	SourcePosition *SourcePosition
}

type Enum struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 enumeration"`

	Name            string `xml:"name,attr"` // Go case
	CType           string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	GLibTypeName    string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType     string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	GLibErrorDomain string `xml:"http://www.gtk.org/introspection/glib/1.0 error-domain,attr"`

	Members   []*Member   `xml:"http://www.gtk.org/introspection/core/1.0 member"`
	Functions []*Function `xml:"http://www.gtk.org/introspection/core/1.0 function"`

	InfoAttrs
	InfoElements
}

// Find implements Searchable.
func (e Enum) Find(typ string) any {
	for _, member := range e.Members {
		if member.Name() == typ || member.CIdentifier == typ || member.GLibName() == typ {
			return member
		}
	}

	for _, function := range e.Functions {
		if function.Name == typ {
			return function
		}
	}

	return nil
}

type Field struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Name     string   `xml:"name,attr"`
	Bits     int      `xml:"bits,attr"`
	Private  bool     `xml:"private,attr"`
	Writable bool     `xml:"writable,attr"` // default false
	Readable *bool    `xml:"readable,attr"` // default true
	AnyType
	Callback *Callback
	Doc      *Doc
}

// IsReadable returns true if the field is readable.
func (f Field) IsReadable() bool {
	return f.Readable == nil || *f.Readable
}

type Function struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	*CallableAttrs
}

type Implements struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 implements"`
	Name    string   `xml:"name,attr"`
}

type Include struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 include"`
	Name    string   `xml:"name,attr"`
	Version Version  `xml:"version,attr"`
}

type InfoAttrs struct {
	Introspectable    *bool   `xml:"introspectable,attr"` // default true
	Deprecated        bool    `xml:"deprecated,attr"`
	DeprecatedVersion Version `xml:"deprecated-version,attr"`
	Version           Version `xml:"version,attr"`
	Stability         string  `xml:"stability,attr"`
}

func (inf InfoAttrs) GetInfoAttrs() InfoAttrs {
	return inf
}

// IsIntrospectable returns true if the InfoAttrs indicates that the type is
// introspectable.
func (inf InfoAttrs) IsIntrospectable() bool {
	return inf.Introspectable == nil || *inf.Introspectable
}

type InfoElements struct {
	DocElements
	Annotations []Annotation `xml:"http://www.gtk.org/introspection/core/1.0 attribute"`
}

func (inf InfoElements) GetInfoElements() InfoElements {
	return inf
}

type InstanceParameter struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 instance-parameter"`
	ParameterAttrs
}

type Interface struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 interface"`
	Name    string   `xml:"name,attr"`

	CType         string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CSymbolPrefix string `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`

	GLibTypeName   string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType    string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	GLibTypeStruct string `xml:"http://www.gtk.org/introspection/glib/1.0 type-struct,attr"`

	Functions      []*Function      `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Methods        []*Method        `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	VirtualMethods []*VirtualMethod `xml:"http://www.gtk.org/introspection/core/1.0 virtual-method"`
	Prerequisites  []*Prerequisite  `xml:"http://www.gtk.org/introspection/core/1.0 prerequisite"`
	Signals        []*Signal        `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`

	InfoAttrs
	InfoElements
}

// Find implements Searchable.
func (i Interface) Find(ident string) any {
	for _, method := range i.Methods {
		if method.Name == ident {
			return method
		}
	}

	for _, virtualMethod := range i.VirtualMethods {
		if virtualMethod.Name == ident {
			return virtualMethod
		}
	}

	for _, function := range i.Functions {
		if function.Name == ident {
			return function
		}
	}

	for _, signal := range i.Signals {
		if signal.Name == ident {
			return signal
		}
	}

	return nil
}

type Member struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 member"`

	Names       []xml.Attr `xml:"name,attr"`
	Value       string     `xml:"value,attr"`
	CIdentifier string     `xml:"http://www.gtk.org/introspection/c/1.0 identifier,attr"`
	GLibNick    string     `xml:"http://www.gtk.org/introspection/glib/1.0 nick,attr"`

	InfoAttrs
	InfoElements
}

func (m Member) Name() string {
	return m.nameAttr(xml.Name{Local: "name"})
}

func (m Member) GLibName() string {
	return m.nameAttr(xml.Name{Space: "http://www.gtk.org/introspection/glib/1.0", Local: "name"})
}

func (m Member) nameAttr(name xml.Name) string {
	for _, attr := range m.Names {
		if attr.Name == name {
			return attr.Value
		}
	}
	return ""
}

type Method struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	*CallableAttrs
}

type Package struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 package"`
	Name    string   `xml:"name,attr"`
}

type Parameter struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 parameter"`
	ParameterAttrs
}

type ParameterAttrs struct {
	Name            string `xml:"name,attr"`
	Direction       string `xml:"direction,attr"`
	Scope           string `xml:"scope,attr"`
	Closure         *int   `xml:"closure,attr"`
	Destroy         *int   `xml:"destroy,attr"`
	CallerAllocates bool   `xml:"caller-allocates,attr"`
	Skip            bool   `xml:"skip,attr"`
	Optional        bool   `xml:"optional,attr"`
	Nullable        bool   `xml:"nullable,attr"`

	TransferOwnership
	AnyType
	Doc *Doc
}

type Parameters struct {
	XMLName           xml.Name           `xml:"http://www.gtk.org/introspection/core/1.0 parameters"`
	InstanceParameter *InstanceParameter `xml:"http://www.gtk.org/introspection/core/1.0 instance-parameter"`
	Parameters        []*Parameter       `xml:"http://www.gtk.org/introspection/core/1.0 parameter"`
}

type Prerequisite struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 prerequisite"`
	Name    string   `xml:"name,attr"`
}

type Property struct{}

type Record struct {
	XMLName              xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 record"`
	Name                 string   `xml:"name,attr"`
	CType                string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	GLibTypeName         string   `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType          string   `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	CSymbolPrefix        string   `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`
	GLibIsGTypeStructFor string   `xml:"http://www.gtk.org/introspection/glib/1.0 is-gtype-struct-for,attr"`
	Disguised            bool     `xml:"disguised,attr"`
	Foreign              bool     `xml:"foreign,attr"`

	Fields       []*Field       `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Functions    []*Function    `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Unions       []*Union       `xml:"http://www.gtk.org/introspection/core/1.0 union"`
	Methods      []*Method      `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	Constructors []*Constructor `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	Properties   []*Property    `xml:"http://www.gtk.org/introspection/core/1.0 property"`

	InfoAttrs
	InfoElements
}

// Find implements Searchable.
func (r Record) Find(typ string) any {
	for _, field := range r.Fields {
		if field.Name == typ {
			return field
		}
	}
	for _, function := range r.Functions {
		if function.Name == typ {
			return function
		}
	}
	for _, method := range r.Methods {
		if method.Name == typ {
			return method
		}
	}
	for _, constructor := range r.Constructors {
		if constructor.Name == typ {
			return constructor
		}
	}
	for _, union := range r.Unions {
		if union.Name == typ {
			return union
		}
	}
	// TODO: properties
	// for _, property := range r.Properties {
	// 	if property.Name == typ {
	// 		return property
	// 	}
	// }
	return nil
}

type ReturnValue struct {
	XMLName        xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 return-value"`
	Scope          string   `xml:"scope,attr"`
	Closure        *int     `xml:"closure,attr"`
	Destroy        *int     `xml:"destroy,attr"`
	Introspectable bool     `xml:"introspectable,attr"`
	Nullable       bool     `xml:"nullable,attr"`
	Skip           bool     `xml:"skip,attr"`
	AllowNone      bool     `xml:"allow-none,attr"`
	TransferOwnership
	DocElements
	AnyType
}

type Signal struct {
	XMLName   xml.Name   `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`
	Name      string     `xml:"name,attr"`
	Detailed  bool       `xml:"detailed,attr"`
	When      SignalWhen `xml:"when,attr"`
	Action    bool       `xml:"action,attr"`
	NoHooks   bool       `xml:"no-hooks,attr"`
	NoRecurse bool       `xml:"no-recurse,attr"`
	InfoElements
	Parameters  *Parameters  `xml:"http://www.gtk.org/introspection/core/1.0 parameters"`
	ReturnValue *ReturnValue `xml:"http://www.gtk.org/introspection/core/1.0 return-value"`
}

type SignalWhen string

const (
	SignalWhenFirst   = "first"
	SignalWhenLast    = "last"
	SignalWhenCleanup = "cleanup"
)

type SourcePosition struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 source-position"`
	Filename string   `xml:"filename,attr"`
	Line     int      `xml:"line,attr"`
	Column   int      `xml:"column,attr"`
}

type TransferOwnership struct {
	TransferOwnership string `xml:"transfer-ownership,attr"`
}

type Type struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 type"`

	Name           string `xml:"name,attr"`
	CType          string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	Introspectable *bool  `xml:"introspectable,attr"`

	DocElements
	// Types is the type's inner types. e.g. generic types of container types
	Types []*Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`
}

func (typ Type) IsIntrospectable() bool {
	return typ.Introspectable == nil || *typ.Introspectable
}

type Union struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 union"`

	Name          string `xml:"name,attr"` // Go case
	CType         string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CSymbolPrefix string `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`
	GLibTypeName  string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType   string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`

	InfoAttrs
	InfoElements

	Fields       []*Field       `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Constructors []*Constructor `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	Methods      []*Method      `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	Functions    []*Function    `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Records      []*Record      `xml:"http://www.gtk.org/introspection/core/1.0 record"`
}

type VarArgs struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 varargs"`
}

type VirtualMethod struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 virtual-method"`

	Invoker string `xml:"invoker,attr"`
	*CallableAttrs
}
