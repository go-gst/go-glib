// Copyright (c) 2013-2014 Conformal Systems <info@conformal.com>
//
// This file originated from: http://opensource.conformal.com/
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

// Package glib provides Go bindings for GLib 2.  Supports version 2.36
// and later.
package glib

// #cgo pkg-config: gio-2.0 glib-2.0 gobject-2.0
// #include <gio/gio.h>
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"unsafe"
)

/*
 * Type conversions
 */

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
func gobool(b C.gboolean) bool {
	if b != 0 {
		return true
	}
	return false
}

/*
 * Unexported vars
 */

type closureContext struct {
	rf       reflect.Value
	userData reflect.Value
}

var (
	nilPtrErr = errors.New("cgo returned unexpected nil pointer")

	closures = struct {
		sync.RWMutex
		m map[*C.GClosure]closureContext
	}{
		m: make(map[*C.GClosure]closureContext),
	}

	signals = struct {
		sync.RWMutex
		m map[SignalHandle]*C.GClosure
	}{
		m: make(map[SignalHandle]*C.GClosure),
	}
)

/*
 * Constants
 */

// Type is a representation of GLib's GType.
type Type uint

const (
	TYPE_INVALID   Type = C.G_TYPE_INVALID
	TYPE_NONE      Type = C.G_TYPE_NONE
	TYPE_INTERFACE Type = C.G_TYPE_INTERFACE
	TYPE_CHAR      Type = C.G_TYPE_CHAR
	TYPE_UCHAR     Type = C.G_TYPE_UCHAR
	TYPE_BOOLEAN   Type = C.G_TYPE_BOOLEAN
	TYPE_INT       Type = C.G_TYPE_INT
	TYPE_UINT      Type = C.G_TYPE_UINT
	TYPE_LONG      Type = C.G_TYPE_LONG
	TYPE_ULONG     Type = C.G_TYPE_ULONG
	TYPE_INT64     Type = C.G_TYPE_INT64
	TYPE_UINT64    Type = C.G_TYPE_UINT64
	TYPE_ENUM      Type = C.G_TYPE_ENUM
	TYPE_FLAGS     Type = C.G_TYPE_FLAGS
	TYPE_FLOAT     Type = C.G_TYPE_FLOAT
	TYPE_DOUBLE    Type = C.G_TYPE_DOUBLE
	TYPE_STRING    Type = C.G_TYPE_STRING
	TYPE_POINTER   Type = C.G_TYPE_POINTER
	TYPE_BOXED     Type = C.G_TYPE_BOXED
	TYPE_PARAM     Type = C.G_TYPE_PARAM
	TYPE_OBJECT    Type = C.G_TYPE_OBJECT
	TYPE_VARIANT   Type = C.G_TYPE_VARIANT
)

// IsValue checks whether the passed in type can be used for g_value_init().
func (t Type) IsValue() bool {
	return gobool(C._g_type_is_value(C.GType(t)))
}

// Name is a wrapper around g_type_name().
func (t Type) Name() string {
	return C.GoString((*C.char)(C.g_type_name(C.GType(t))))
}

// Depth is a wrapper around g_type_depth().
func (t Type) Depth() uint {
	return uint(C.g_type_depth(C.GType(t)))
}

// Parent is a wrapper around g_type_parent().
func (t Type) Parent() Type {
	return Type(C.g_type_parent(C.GType(t)))
}

// IsA is a wrapper around g_type_is_a().
func (t Type) IsA(isAType Type) bool {
	return gobool(C.g_type_is_a(C.GType(t), C.GType(isAType)))
}

// TypeFromName is a wrapper around g_type_from_name
func TypeFromName(typeName string) Type {
	cstr := (*C.gchar)(C.CString(typeName))
	defer C.free(unsafe.Pointer(cstr))
	return Type(C.g_type_from_name(cstr))
}

//TypeNextBase is a wrapper around g_type_next_base
func TypeNextBase(leafType, rootType Type) Type {
	return Type(C.g_type_next_base(C.GType(leafType), C.GType(rootType)))
}

// SettingsBindFlags is a representation of GLib's GSettingsBindFlags.
type SettingsBindFlags int

const (
	SETTINGS_BIND_DEFAULT        SettingsBindFlags = C.G_SETTINGS_BIND_DEFAULT
	SETTINGS_BIND_GET            SettingsBindFlags = C.G_SETTINGS_BIND_GET
	SETTINGS_BIND_SET            SettingsBindFlags = C.G_SETTINGS_BIND_SET
	SETTINGS_BIND_NO_SENSITIVITY SettingsBindFlags = C.G_SETTINGS_BIND_NO_SENSITIVITY
	SETTINGS_BIND_GET_NO_CHANGES SettingsBindFlags = C.G_SETTINGS_BIND_GET_NO_CHANGES
	SETTINGS_BIND_INVERT_BOOLEAN SettingsBindFlags = C.G_SETTINGS_BIND_INVERT_BOOLEAN
)

// UserDirectory is a representation of GLib's GUserDirectory.
type UserDirectory int

const (
	USER_DIRECTORY_DESKTOP      UserDirectory = C.G_USER_DIRECTORY_DESKTOP
	USER_DIRECTORY_DOCUMENTS    UserDirectory = C.G_USER_DIRECTORY_DOCUMENTS
	USER_DIRECTORY_DOWNLOAD     UserDirectory = C.G_USER_DIRECTORY_DOWNLOAD
	USER_DIRECTORY_MUSIC        UserDirectory = C.G_USER_DIRECTORY_MUSIC
	USER_DIRECTORY_PICTURES     UserDirectory = C.G_USER_DIRECTORY_PICTURES
	USER_DIRECTORY_PUBLIC_SHARE UserDirectory = C.G_USER_DIRECTORY_PUBLIC_SHARE
	USER_DIRECTORY_TEMPLATES    UserDirectory = C.G_USER_DIRECTORY_TEMPLATES
	USER_DIRECTORY_VIDEOS       UserDirectory = C.G_USER_DIRECTORY_VIDEOS
)

const USER_N_DIRECTORIES int = C.G_USER_N_DIRECTORIES

/*
 * GApplicationFlags
 */

type ApplicationFlags int

const (
	APPLICATION_FLAGS_NONE           ApplicationFlags = C.G_APPLICATION_FLAGS_NONE
	APPLICATION_IS_SERVICE           ApplicationFlags = C.G_APPLICATION_IS_SERVICE
	APPLICATION_HANDLES_OPEN         ApplicationFlags = C.G_APPLICATION_HANDLES_OPEN
	APPLICATION_HANDLES_COMMAND_LINE ApplicationFlags = C.G_APPLICATION_HANDLES_COMMAND_LINE
	APPLICATION_SEND_ENVIRONMENT     ApplicationFlags = C.G_APPLICATION_SEND_ENVIRONMENT
	APPLICATION_NON_UNIQUE           ApplicationFlags = C.G_APPLICATION_NON_UNIQUE
)

// goMarshal is called by the GLib runtime when a closure needs to be invoked.
// The closure will be invoked with as many arguments as it can take, from 0 to
// the full amount provided by the call. If the closure asks for more parameters
// than there are to give, a warning is printed to stderr and the closure is
// not run.
//
//export goMarshal
func goMarshal(closure *C.GClosure, retValue *C.GValue,
	nParams C.guint, params *C.GValue,
	invocationHint C.gpointer, marshalData *C.GValue) {

	// Get the context associated with this callback closure.
	closures.RLock()
	cc := closures.m[closure]
	closures.RUnlock()

	// Get number of parameters passed in.  If user data was saved with the
	// closure context, increment the total number of parameters.
	nGLibParams := int(nParams)
	nTotalParams := nGLibParams
	if cc.userData.IsValid() {
		nTotalParams++
	}

	// Get number of parameters from the callback closure.  If this exceeds
	// the total number of marshaled parameters, a warning will be printed
	// to stderr, and the callback will not be run.
	nCbParams := cc.rf.Type().NumIn()
	if nCbParams > nTotalParams {
		fmt.Fprintf(os.Stderr,
			"too many closure args: have %d, max allowed %d\n",
			nCbParams, nTotalParams)
		return
	}

	// Create a slice of reflect.Values as arguments to call the function.
	gValues := gValueSlice(params, nCbParams)
	args := make([]reflect.Value, 0, nCbParams)

	// Fill beginning of args, up to the minimum of the total number of callback
	// parameters and parameters from the glib runtime.
	for i := 0; i < nCbParams && i < nGLibParams; i++ {
		v := &Value{&gValues[i]}
		val, err := v.GoValue()
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"no suitable Go value for arg %d: %v\n", i, err)
			return
		}
		// Parameters that are descendants of GObject come wrapped in another GObject.
		// For C applications, the default marshaller (g_cclosure_marshal_VOID__VOID in
		// gmarshal.c in the GTK glib library) 'peeks' into the enclosing object and
		// passes the wrapped object to the handler. Use the *Object.goValue function
		// to emulate that for Go signal handlers.
		switch objVal := val.(type) {
		case *Object:
			innerVal, err := objVal.goValue()
			if err != nil {
				// print warning and leave val unchanged to preserve old
				// behavior
				fmt.Fprintf(os.Stderr,
					"warning: no suitable Go value from object for arg %d: %v\n", i, err)
			} else {
				val = innerVal
			}
		}
		rv := reflect.ValueOf(val)
		args = append(args, rv.Convert(cc.rf.Type().In(i)))
	}

	// If non-nil user data was passed in and not all args have been set,
	// get and set the reflect.Value directly from the GValue.
	if cc.userData.IsValid() && len(args) < cap(args) {
		args = append(args, cc.userData.Convert(cc.rf.Type().In(nCbParams-1)))
	}

	// Call closure with args. If the callback returns one or more
	// values, save the GValue equivalent of the first.
	rv := cc.rf.Call(args)
	if retValue != nil && len(rv) > 0 {
		if g, err := GValue(rv[0].Interface()); err != nil {
			fmt.Fprintf(os.Stderr,
				"cannot save callback return value: %v", err)
		} else {
			C.g_value_copy(g.native(), retValue)
		}
	}
}

// gValueSlice converts a C array of GValues to a Go slice.
func gValueSlice(values *C.GValue, nValues int) (slice []C.GValue) {
	header := (*reflect.SliceHeader)((unsafe.Pointer(&slice)))
	header.Cap = nValues
	header.Len = nValues
	header.Data = uintptr(unsafe.Pointer(values))
	return
}

/*
 * Main event loop
 */

type SourceHandle uint

// IdleAdd adds an idle source to the default main event loop
// context.  After running once, the source func will be removed
// from the main event loop, unless f returns a single bool true.
//
// This function will cause a panic when f eventually runs if the
// types of args do not match those of f.
func IdleAdd(f interface{}, args ...interface{}) (SourceHandle, error) {
	// f must be a func with no parameters.
	rf := reflect.ValueOf(f)
	if rf.Type().Kind() != reflect.Func {
		return 0, errors.New("f is not a function")
	}

	// Create an idle source func to be added to the main loop context.
	idleSrc := C.g_idle_source_new()
	if idleSrc == nil {
		return 0, nilPtrErr
	}
	return sourceAttach(idleSrc, rf, args...)
}

// TimeoutAdd adds an timeout source to the default main event loop
// context.  After running once, the source func will be removed
// from the main event loop, unless f returns a single bool true.
//
// This function will cause a panic when f eventually runs if the
// types of args do not match those of f.
// timeout is in milliseconds
func TimeoutAdd(timeout uint, f interface{}, args ...interface{}) (SourceHandle, error) {
	// f must be a func with no parameters.
	rf := reflect.ValueOf(f)
	if rf.Type().Kind() != reflect.Func {
		return 0, errors.New("f is not a function")
	}

	// Create a timeout source func to be added to the main loop context.
	timeoutSrc := C.g_timeout_source_new(C.guint(timeout))
	if timeoutSrc == nil {
		return 0, nilPtrErr
	}

	return sourceAttach(timeoutSrc, rf, args...)
}

// sourceAttach attaches a source to the default main loop context.
func sourceAttach(src *C.struct__GSource, rf reflect.Value, args ...interface{}) (SourceHandle, error) {
	if src == nil {
		return 0, nilPtrErr
	}

	// rf must be a func with no parameters.
	if rf.Type().Kind() != reflect.Func {
		C.g_source_destroy(src)
		return 0, errors.New("rf is not a function")
	}

	// Create a new GClosure from f that invalidates itself when
	// f returns false.  The error is ignored here, as this will
	// always be a function.
	var closure *C.GClosure
	closure, _ = ClosureNew(rf.Interface(), args...)

	// Remove closure context when closure is finalized.
	C._g_closure_add_finalize_notifier(closure)

	// Set closure to run as a callback when the idle source runs.
	C.g_source_set_closure(src, closure)

	// Attach the idle source func to the default main event loop
	// context.
	cid := C.g_source_attach(src, nil)
	return SourceHandle(cid), nil
}

// Destroy is a wrapper around g_source_destroy()
func (v *Source) Destroy() {
	C.g_source_destroy(v.native())
}

// IsDestroyed is a wrapper around g_source_is_destroyed()
func (v *Source) IsDestroyed() bool {
	return gobool(C.g_source_is_destroyed(v.native()))
}

// Unref is a wrapper around g_source_unref()
func (v *Source) Unref() {
	C.g_source_unref(v.native())
}

// Ref is a wrapper around g_source_ref()
func (v *Source) Ref() *Source {
	c := C.g_source_ref(v.native())
	if c == nil {
		return nil
	}
	return (*Source)(c)
}

// SourceRemove is a wrapper around g_source_remove()
func SourceRemove(src SourceHandle) bool {
	return gobool(C.g_source_remove(C.guint(src)))
}

/*
 * Miscellaneous Utility Functions
 */

// GetHomeDir is a wrapper around g_get_home_dir().
func GetHomeDir() string {
	c := C.g_get_home_dir()
	return C.GoString((*C.char)(c))
}

// GetUserCacheDir is a wrapper around g_get_user_cache_dir().
func GetUserCacheDir() string {
	c := C.g_get_user_cache_dir()
	return C.GoString((*C.char)(c))
}

// GetUserDataDir is a wrapper around g_get_user_data_dir().
func GetUserDataDir() string {
	c := C.g_get_user_data_dir()
	return C.GoString((*C.char)(c))
}

// GetUserConfigDir is a wrapper around g_get_user_config_dir().
func GetUserConfigDir() string {
	c := C.g_get_user_config_dir()
	return C.GoString((*C.char)(c))
}

// GetUserRuntimeDir is a wrapper around g_get_user_runtime_dir().
func GetUserRuntimeDir() string {
	c := C.g_get_user_runtime_dir()
	return C.GoString((*C.char)(c))
}

// GetUserSpecialDir is a wrapper around g_get_user_special_dir().  A
// non-nil error is returned in the case that g_get_user_special_dir()
// returns NULL to differentiate between NULL and an empty string.
func GetUserSpecialDir(directory UserDirectory) (string, error) {
	c := C.g_get_user_special_dir(C.GUserDirectory(directory))
	if c == nil {
		return "", nilPtrErr
	}
	return C.GoString((*C.char)(c)), nil
}

/*
 * GObject
 */

// IObject is an interface type implemented by Object and all types which embed
// an Object.  It is meant to be used as a type for function arguments which
// require GObjects or any subclasses thereof.
type IObject interface {
	toGObject() *C.GObject
	toObject() *Object
}

/*
 * GInitiallyUnowned
 */

// InitiallyUnowned is a representation of GLib's GInitiallyUnowned.
type InitiallyUnowned struct {
	// This must be a pointer so copies of the ref-sinked object
	// do not outlive the original object, causing an unref
	// finalizer to prematurely run.
	*Object
}

// Native returns a pointer to the underlying GObject.  This is implemented
// here rather than calling Native on the embedded Object to prevent a nil
// pointer dereference.
func (v *InitiallyUnowned) Native() uintptr {
	if v == nil || v.Object == nil {
		return uintptr(unsafe.Pointer(nil))
	}
	return v.Object.Native()
}

type Signal struct {
	name     string
	signalId C.guint
}

func SignalNew(s string) (*Signal, error) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	signalId := C._g_signal_new((*C.gchar)(cstr))

	if signalId == 0 {
		return nil, fmt.Errorf("invalid signal name: %s", s)
	}

	return &Signal{
		name:     s,
		signalId: signalId,
	}, nil
}

func (s *Signal) String() string {
	return s.name
}

type Quark uint32

// GetApplicationName is a wrapper around g_get_application_name().
func GetApplicationName() string {
	c := C.g_get_application_name()

	return C.GoString((*C.char)(c))
}

// SetApplicationName is a wrapper around g_set_application_name().
func SetApplicationName(name string) {
	cstr := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(cstr))

	C.g_set_application_name(cstr)
}

// InitI18n initializes the i18n subsystem.
func InitI18n(domain string, dir string) {
	domainStr := C.CString(domain)
	defer C.free(unsafe.Pointer(domainStr))

	dirStr := C.CString(dir)
	defer C.free(unsafe.Pointer(dirStr))

	C.init_i18n(domainStr, dirStr)
}

// Local localizes a string using gettext
func Local(input string) string {
	cstr := C.CString(input)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(C.localize(cstr))
}
