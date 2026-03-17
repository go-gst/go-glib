package typesystem

type CastablePrimitive struct {
	BaseType
}

// canBeCasted implements CastableType.
func (c *CastablePrimitive) canBeCasted() {}

var _ CastableType = (*CastablePrimitive)(nil)

func prim(girName, cType, cGoType, goType string) *CastablePrimitive {
	return &CastablePrimitive{
		BaseType: BaseType{
			GirName: girName,
			CTyp:    cType,
			CGoTyp:  cGoType,
			GoTyp:   goType,
		},
	}
}

// BooleanPrimitive describes a boolean. C booleans differ from go booleans and must be converted differently
type BooleanPrimitive struct {
	BaseType
}

var Gboolean = &BooleanPrimitive{
	BaseType: BaseType{
		GirName: "gboolean",
		CTyp:    "gboolean",
		CGoTyp:  "C.gboolean",
		GoTyp:   "bool",
	},
}

// StringPrimitive describes a string, which is builtin in go but a char array in C
type StringPrimitive struct {
	GirName string
}

var _ Type = (*StringPrimitive)(nil)

var Utf8 = &StringPrimitive{
	GirName: "utf8",
}

var Filename = &StringPrimitive{
	GirName: "filename",
}

// minPointersRequired implements minPointerConstrainedType.
func (a *StringPrimitive) minPointersRequired() int {
	return 1
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *StringPrimitive) maxPointersAllowed() int {
	return 1
}

// GIRName implements Type.
func (a *StringPrimitive) GIRName() string {
	return a.GirName
}

// GoType implements Type.
func (a *StringPrimitive) GoType(_ int) string {
	return "string"
}

// CGoType implements Type.
func (a *StringPrimitive) CGoType(_ int) string {
	return "*C.gchar"
}

// CType implements Type.
func (a *StringPrimitive) CType(_ int) string {
	return "gchar*"
}

// GoTypeRequiredImport implements Type.
func (a *StringPrimitive) GoTypeRequiredImport() (alias string, module string) {
	return "", ""
}

var (
	Guint         = prim("guint", "guint", "C.guint", "uint")
	Guint8        = prim("guint8", "guint8", "C.guint8", "uint8")
	Guint16       = prim("guint16", "guint16", "C.guint16", "uint16")
	Guint32       = prim("guint32", "guint32", "C.guint32", "uint32")
	Guint64       = prim("guint64", "guint64", "C.guint64", "uint64")
	Gint          = prim("gint", "gint", "C.gint", "int32") // C int is 32 bit
	Gint8         = prim("gint8", "gint8", "C.gint8", "int8")
	Gint16        = prim("gint16", "gint16", "C.gint16", "int16")
	Gint32        = prim("gint32", "gint32", "C.gint32", "int32")
	Gint64        = prim("gint64", "gint64", "C.gint64", "int64")
	Gshort        = prim("gshort", "gshort", "C.gshort", "int16")
	Gushort       = prim("gushort", "gushort", "C.gushort", "uint16")
	Gsize         = prim("gsize", "gsize", "C.gsize", "uint")
	Gssize        = prim("gssize", "gssize", "C.gssize", "int")
	Gchar         = prim("gchar", "gchar", "C.char", "byte")
	Guchar        = prim("guchar", "guchar", "C.guchar", "byte")
	Gunichar      = prim("gunichar", "gunichar", "C.gunichar", "uint32")
	Gfloat        = prim("gfloat", "gfloat", "C.gfloat", "float32")
	Gdouble       = prim("gdouble", "gdouble", "C.gdouble", "float64")
	Gpointer      = prim("gpointer", "gpointer", "C.gpointer", "unsafe.Pointer")
	Gconstpointer = prim("gconstpointer", "gconstpointer", "C.gconstpointer", "unsafe.Pointer")
	Gintptr       = prim("gintptr", "gintptr", "C.gintptr", "uintptr")
	Guintptr      = prim("guintptr", "guintptr", "C.guintptr", "uintptr")
	Glong         = prim("glong", "glong", "C.glong", "int32")
	Gulong        = prim("gulong", "gulong", "C.gulong", "uint32")
	Time_t        = prim("time_t", "time_t", "C.time_t", "uint64") // TODO: check go type
	Pid_t         = prim("pid_t", "pid_t", "C.pid_t", "int")       // process ids
	Ino_t         = prim("ino_t", "ino_t", "C.ino_t", "uint")      // file serial ids
	Uid_t         = prim("uid_t", "uid_t", "C.uid_t", "uint")      // user ids, may be signed on some platforms
	Gid_t         = prim("gid_t", "gid_t", "C.gid_t", "uint")      // group ids, may be signed on some platforms
)

var Primitives = []Type{
	Guint,
	Guint8,
	Guint16,
	Guint32,
	Guint64,

	Gint,
	Gint8,
	Gint16,
	Gint32,
	Gint64,

	Gshort,
	Gushort,

	Gsize,
	Gssize,

	Gchar,
	Guchar,
	Gunichar,

	Gboolean,

	Gfloat,
	Gdouble,

	Utf8,
	Filename,

	Gintptr,
	Guintptr,
	Gpointer,
	Gconstpointer,

	Glong,
	Gulong,

	Time_t,

	Pid_t,
	Ino_t,
	Uid_t,
	Gid_t,

	Void,
}

type VoidType struct {
	BaseType
}

var Void = &VoidType{
	BaseType: BaseType{
		GirName: "none",
		GoTyp:   typeInvalid,
		CGoTyp:  "C.void",
		CTyp:    "void",
	},
}

// findBuiltinPrimitiveByCType is needed because the GIR names are not always
// unique, e.g. gpointer and gconstpointer are both "gpointer" in GIR
func findBuiltinPrimitiveByCType(ctype string) Type {
	for _, p := range Primitives {
		if p.CType(0) == ctype {
			return p
		}
	}

	return nil
}
func findBuiltinPrimitiveByGIRName(girname string) Type {
	for _, p := range Primitives {
		if p.GIRName() == girname {
			return p
		}
	}

	return nil
}

var IncompatibleCTypes = []string{
	"long double", // may be more precise than float64, so we do not have a go equivalent
	"tm",          // requires time.h
	"va_list",
}

func ctypeIsIncompatible(typ string) bool {
	for _, ign := range IncompatibleCTypes {
		if ign == typ {
			return true
		}
	}

	return false
}
