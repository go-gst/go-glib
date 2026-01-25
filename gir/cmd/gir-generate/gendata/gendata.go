// Package gendata contains data used to generate GTK4 bindings for Go. It
// exists primarily to be used externally.
package gendata

import (
	"slices"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/cmd/gir-generate/genmain"
	"github.com/go-gst/go-glib/gir/girgen/generators"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
	girfiles_goglib "github.com/go-gst/go-glib/girs"
)

const Module = "github.com/go-gst/go-glib/pkg"

var Main = genmain.Data{
	Module:        Module,
	GirFiles:      girfiles_goglib.GirFiles,
	Preprocessors: Preprocessors,

	Documentation: generators.NewGtkGodocGenerator,

	Config: typesystem.Config{
		GIRReplacements: map[string]string{
			"GType": "GObject.Type", // GType is often referred to as a global type instead of GObject scoped
		},
		Namespaces: map[string]typesystem.NamespaceConfig{
			"GLib-2": {
				MinVersion: "2.80",
				ManualTypes: []typesystem.Type{
					&typesystem.Container{
						GirName: "List",
						C:       "GList",
						CGo:     "C.GList",
						MakeGoType: func(innerTypes []typesystem.CouldBeForeign[typesystem.Type]) string {
							return "[]" + innerTypes[0].NamespacedGoType(1)
						},
						NumInnerTypes:        1,
						FromGlibFullFunction: "UnsafeListFromGlibFull",
						FromGlibNoneFunction: "UnsafeListFromGlibNone",
					},
					&typesystem.Container{
						GirName: "SList",
						C:       "GSList",
						CGo:     "C.GSList",
						MakeGoType: func(innerTypes []typesystem.CouldBeForeign[typesystem.Type]) string {
							return "[]" + innerTypes[0].NamespacedGoType(1)
						},
						NumInnerTypes:        1,
						FromGlibFullFunction: "UnsafeSListFromGlibFull",
						FromGlibNoneFunction: "UnsafeSListFromGlibNone",
					},
					&typesystem.Callback{
						BaseType: typesystem.BaseType{
							GirName: "DestroyNotify",
							GoTyp:   "DestroyNotify",
							CGoTyp:  "C.GDestroyNotify",
							CTyp:    "GDestroyNotify",
						},
						Parameters: &typesystem.Parameters{
							CReturn: typesystem.NewManualParam("cret", "goret", typesystem.Void, 0),
							GIRParameters: typesystem.ParamList{
								typesystem.NewManualParam("arg0", "goarg0", typesystem.Gpointer, 0),
							},
						},
						TrampolineName: "destroyUserdata",
					},
					// TODO: implement Variant
					// &typesystem.Record{
					// 	BaseType: typesystem.BaseType{
					// 		GirName: "Variant",
					// 		GoTyp:   "Variant",
					// 		CGoTyp:  "C.GVariant",
					// 		CTyp:    "GVariant",
					// 	},
					// 	BaseConversions: typesystem.BaseConversions{
					// 		FromGlibBorrowFunction: "UnsafeVariantFromGlibBorrow",
					// 		FromGlibFullFunction:   "UnsafeVariantFromGlibFull",
					// 		FromGlibNoneFunction:   "UnsafeVariantFromGlibNone",
					// 		ToGlibNoneFunction:     "UnsafeVariantToGlibNone",
					// 		ToGlibFullFunction:     "UnsafeVariantToGlibFull",
					// 	},
					// },
					&typesystem.Record{
						BaseType: typesystem.BaseType{
							GirName: "Error",
							CGoTyp:  "C.GError",
							CTyp:    "GError",
							GoTyp:   "error",
						},
						BaseConversions: typesystem.BaseConversions{
							FromGlibFullFunction: "UnsafeErrorFromGlibFull",
							ToGlibFullFunction:   "UnsafeErrorToGlibFull",
						},
					},
				},
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// not found:
					typesystem.IgnoreMatching("StatBuf"),
					typesystem.IgnoreMatching("access"),
					typesystem.IgnoreMatching("chdir"),
					typesystem.IgnoreMatching("chmod"),
					typesystem.IgnoreMatching("close"),
					typesystem.IgnoreMatching("closefrom"),
					typesystem.IgnoreMatching("creat"),
					typesystem.IgnoreMatching("date_get_week_of_year"),
					typesystem.IgnoreMatching("date_get_weeks_in_year"),
					typesystem.IgnoreMatching("fdwalk_set_cloexec"),
					typesystem.IgnoreMatching("fsync"),
					typesystem.IgnoreMatching("log_get_always_fatal"),
					typesystem.IgnoreMatching("lstat"),
					typesystem.IgnoreMatching("mkdir"),
					typesystem.IgnoreMatching("open"),
					typesystem.IgnoreMatching("remove"),
					typesystem.IgnoreMatching("rename"),
					typesystem.IgnoreMatching("rmdir"),
					typesystem.IgnoreMatching("source_dup_context"),
					typesystem.IgnoreMatching("stat"),
					typesystem.IgnoreMatching("string_copy"),
					typesystem.IgnoreMatching("unlink"),

					typesystem.IgnoreByRegex("Date.*"),
					typesystem.IgnoreMatching("Source"),
					typesystem.IgnoreMatching("TestLogMsg"),
					typesystem.IgnoreMatching("String"),
					typesystem.IgnoreMatching("Thread"),
					typesystem.IgnoreMatching("ThreadPool"),

					typesystem.IgnoreMatching("ref_string_equal"), // not found

					typesystem.IgnoreMatching("HookFlagMask"), // Has a member of the same name

					typesystem.IgnoreMatching("Variant"),           // TODO: implement manually
					typesystem.IgnoreMatching("variant_get_gtype"), // implemented with gvalue in gobject

					typesystem.IgnoreMatching("ucs4_to_utf16"), // returns a pointer instead of an array

					// Nothing "Unix" is going to be available on Windows.
					typesystem.IgnoreByRegex(".*[Uu]nix.*"),
					// Useless
					typesystem.IgnoreMatching("NullifyPointer"),
					typesystem.IgnoreByRegex("[Aa]tomic.*"),
					typesystem.IgnoreByRegex("ATOMIC.*"),
					// Dangerous.
					typesystem.IgnoreMatching("IOChannel.read"),
					typesystem.IgnoreMatching("Bytes.new_take"),
					typesystem.IgnoreMatching("Bytes.new_static"),
					typesystem.IgnoreMatching("Bytes.unref_to_data"),
					typesystem.IgnoreMatching("Bytes.unref_to_array"),

					typesystem.IgnoreMatching("G_WIN32_MSG_HANDLE"),
					typesystem.IgnoreMatching("GLIB_VERSION_MIN_REQUIRED"),

					typesystem.IgnoreMatching("strv_get_type"), // requires gobject

					// see https://gitlab.gnome.org/GNOME/gobject-introspection/-/issues/305#note_981623
					// Container structures that are unused:
					typesystem.IgnoreMatching("Array"),
					typesystem.IgnoreMatching("Queue"),
					typesystem.IgnoreMatching("Tree"),
					typesystem.IgnoreMatching("PtrArray"),
					typesystem.IgnoreMatching("HashTable"),

					// Differs between platforms
					typesystem.IgnoreMatching("Pid"),
				},
			},
			"Gio-2": {
				ManualTypes: []typesystem.Type{
					// TODO: implement a cancellable type that translates to context.Context
					// &typesystem.Record{
					// 	BaseType: typesystem.BaseType{
					// 		GirName: "Cancellable",
					// 		CGoTyp:  "C.GCancellable",
					// 		CTyp:    "GCancellable",
					// 		GoTyp:   "context.Context",

					// 		GoImport: "context",
					// 	},
					// 	BaseConversions: typesystem.BaseConversions{
					// 		FromGlibBorrowFunction: "",
					// 		FromGlibFullFunction:   "",
					// 		FromGlibNoneFunction:   "NewCancellableContext",
					// 		ToGlibNoneFunction:     "UnsafeGCancellableToGlibNone",
					// 		ToGlibFullFunction:     "",
					// 	},
					// },
				},
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// Nothing "Unix" is going to be available on Windows.
					typesystem.IgnoreByRegex(".*[Uu]nix.*"),
					typesystem.IgnoreByRegex(".*Subprocess.*"),

					typesystem.IgnoreByRegex("FileDescriptorBased"),
					typesystem.IgnoreByRegex("SettingsBackend"),
					typesystem.IgnoreByRegex("DesktopAppInfo.*"),
					typesystem.IgnoreByRegex("DBus.*"),
					typesystem.IgnoreByRegex("ThreadedResolver.*"),

					typesystem.IgnoreMatching("ZlibCompressor"),

					typesystem.IgnoreMatching("networking_init"),

					typesystem.IgnoreMatching("Resource"),
					typesystem.IgnoreMatching("resources_has_children"),

					typesystem.IgnoreMatching("DataInputStream.read_byte"), // collides with BufferedInputStream.read_byte
				},
			},
			"GObject-2": {
				MinVersion: "2.80",
				ManualTypes: func() []typesystem.Type { // use an immediately invoked function to create the circular references
					object := &typesystem.Class{
						BaseType: typesystem.BaseType{
							GirName: "Object",
							GoTyp:   "ObjectInstance",
							CTyp:    "GObject",
							CGoTyp:  "C.GObject",
						},
						GoInterfaceName: "Object",
						Doc:             typesystem.Doc{},
						BaseConversions: typesystem.BaseConversions{
							FromGlibBorrowFunction: "UnsafeObjectFromGlibBorrow", // borrow is needed for subclassing
							FromGlibFullFunction:   "UnsafeObjectFromGlibFull",
							FromGlibNoneFunction:   "UnsafeObjectFromGlibNone",
							ToGlibNoneFunction:     "UnsafeObjectToGlibNone",
							ToGlibFullFunction:     "UnsafeObjectToGlibFull",
						},
						GoExtendOverrideStructName: "ObjectOverrides",
						GoUnsafeApplyOverridesName: "UnsafeApplyObjectOverrides",
					}
					objectClass := &typesystem.Record{
						BaseType: typesystem.BaseType{
							GirName: "ObjectClass",
							GoTyp:   "ObjectClass",
							CTyp:    "GObjectClass",
							CGoTyp:  "C.GObjectClass",
						},
						BaseConversions: typesystem.BaseConversions{
							FromGlibBorrowFunction: "UnsafeObjectClassFromGlibBorrow",
							// not transferable, we don't want any methods with this
						},
					}

					object.TypeStruct = objectClass
					objectClass.IsTypeStructFor = object

					return []typesystem.Type{
						&typesystem.Alias{
							BaseType: typesystem.BaseType{
								GirName: "Type",
								CTyp:    "GType",
								CGoTyp:  "C.GType",
								GoTyp:   "Type",
							},
							AliasedType: typesystem.CouldBeForeign[typesystem.Type]{
								Namespace: nil,
								Type:      typesystem.Guint64,
							},
						},
						object,
						objectClass,
						&typesystem.Record{
							BaseType: typesystem.BaseType{
								GirName: "Value",
								GoTyp:   "Value",
								CTyp:    "GValue",
								CGoTyp:  "C.GValue",
							},
							BaseConversions: typesystem.BaseConversions{
								FromGlibBorrowFunction: "ValueFromNative",

								// these should get implemented manually, because "any" would be a better match
								FromGlibFullFunction: "UnsafeValueFromGlibUseAnyInstead",
								FromGlibNoneFunction: "UnsafeValueFromGlibUseAnyInstead",
								ToGlibNoneFunction:   "UnsafeValueToGlibUseAnyInstead",
								ToGlibFullFunction:   "UnsafeValueToGlibUseAnyInstead",
							},
						},
						&typesystem.Record{
							BaseType: typesystem.BaseType{
								GirName: "ParamSpec",
								GoTyp:   "ParamSpec",
								CTyp:    "GParamSpec",
								CGoTyp:  "C.GParamSpec",
							},
							BaseConversions: typesystem.BaseConversions{
								FromGlibBorrowFunction: "UnsafeParamSpecFromGlibBorrow",
								FromGlibFullFunction:   "UnsafeParamSpecFromGlibFull",
								FromGlibNoneFunction:   "UnsafeParamSpecFromGlibNone",
								ToGlibNoneFunction:     "UnsafeParamSpecToGlibNone",
								ToGlibFullFunction:     "UnsafeParamSpecToGlibFull",
							},
						},
					}
				}(),
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// manually implemented, but hidden from the user
					typesystem.IgnoreMatching("ParamSpecClass"),
					typesystem.IgnoreMatching("Closure"),
					typesystem.IgnoreMatching("SignalGroup"),
					typesystem.IgnoreMatching("SignalQuery"),
					typesystem.IgnoreMatching("TypeQuery"),

					// maybe something for later, not needed now:
					typesystem.IgnoreMatching("TypeModule"),
					typesystem.IgnoreMatching("TypeModuleClass"),
					typesystem.IgnoreMatching("ParamSpecPool"),
					typesystem.IgnoreMatching("TypePlugin"),
					typesystem.IgnoreMatching("TypePluginClass"),
					typesystem.IgnoreMatching("ParamSpecTypeInfo"), // needed for registering param types

					typesystem.IgnoreMatching("Binding"), // is this needed?

					// signal handler accumulators, maybe implement them manually?
					typesystem.IgnoreMatching("signal_accumulator_first_wins"),
					typesystem.IgnoreMatching("signal_accumulator_true_handled"),

					// type registration is implemented manually but hidden from the user
					typesystem.IgnoreMatching("type_register_fundamental"),
					typesystem.IgnoreMatching("type_register_static"),

					typesystem.IgnoreMatching("type_check_value_holds"),
					typesystem.IgnoreMatching("type_check_value"),
					typesystem.IgnoreMatching("strdup_value_contents"),

					typesystem.IgnoreMatching("TypeInterface"), // base struct for interfaces, not needed
					typesystem.IgnoreMatching("TypeClass"),     // base struct for classes, not needed
				},
			},
		},
	},
}

// Preprocessors defines a list of preprocessors that the main generator will
// use. It's mostly used for renaming colliding types/identifiers.
var Preprocessors = []gir.Preprocessor{
	// Collision due to case conversions between record and function:
	gir.TypeRenamer("GLib-2.file_test", "test_file"),

	gir.RemoveCIncludes("Gio-2.0.gir", "gio/gdesktopappinfo.h"),
	// These probably shouldn't be built on Windows.
	gir.RemovePkgconfig("Gio-2.0.gir", "gio-unix-2.0"),
	gir.RemoveCIncludes("Gio-2.0.gir", "gio/gfiledescriptorbased.h", `gio/gunix.*\.h`),

	// Fix GAsyncReadyCallback missing the closure bit for the user_data
	// parameter.
	gir.PreprocessorFunc(func(repos gir.Repositories) {
		callback := repos.FindFullType("Gio-2.AsyncReadyCallback").(*gir.Callback)

		userDataIx := slices.IndexFunc(
			callback.Parameters.Parameters,
			func(p *gir.Parameter) bool { return p.Name == "data" },
		)

		userData := callback.Parameters.Parameters[userDataIx]
		userData.Closure = &userDataIx
	}),

	// Collide in other namespaces (e.g. Gio) when implementing TypePlugin and TypeModule
	gir.RenameCallable("GObject-2.TypePlugin.use", "use_plugin"),
	gir.RenameCallable("GObject-2.TypePlugin.unuse", "unuse_plugin"),

	// Collide with GObject.Connect:
	gir.RenameCallable("Gio-2.Socket.connect", "connect_socket"),
	gir.RenameCallable("Gio-2.SocketClient.connect", "connect_socket_client"),
	gir.RenameCallable("Gio-2.SocketConnection.connect", "connect_socket_connection"),
	gir.RenameCallable("Gio-2.Proxy.connect", "connect_proxy"),

	// Less confusing because C.int differs from int in Go.
	gir.RenameCallable("GObject-2.param_spec_int", "param_spec_int32"),
}
