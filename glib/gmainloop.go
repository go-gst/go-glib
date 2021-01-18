package glib

/*
#include "glib.go.h"
*/
import "C"
import "runtime"

// MainLoop is a go representation of a GMainLoop. It can be used to block execution
// while a pipeline is running, and also allows for event sources and signals to be used
// across gstreamer objects.
type MainLoop struct {
	ptr *C.GMainLoop
}

func wrapMainLoop(loop *C.GMainLoop) *MainLoop {
	return &MainLoop{ptr: loop}
}

// NewMainLoop creates a new GMainLoop. If ctx is nil then the default context is used.
// If isRunning is true the loop will automatically start, however, this function will not
// block. To block on the loop itself you will still need to call MainLoop.Run().
//
// A MainLoop is required when wishing to handle signals to/from elements asynchronously.
// Otherwise you will need to iterate on the DefaultMainContext (or an external created one)
// manually.
func NewMainLoop(ctx *MainContext, isRunning bool) *MainLoop {
	var gCtx *C.GMainContext
	if ctx != nil {
		gCtx = ctx.native()
	}
	loop := C.g_main_loop_new(gCtx, gbool(isRunning))
	ml := wrapMainLoop(loop)
	runtime.SetFinalizer(ml, (*MainLoop).Unref)
	return ml
}

// Instance returns the underlying GMainLoop instance.
func (m *MainLoop) Instance() *C.GMainLoop { return m.ptr }

// Ref increases the ref count on the main loop by one. It returns the original main loop
// for convenience in return functions.
func (m *MainLoop) Ref() *MainLoop {
	return wrapMainLoop(C.g_main_loop_ref(m.Instance()))
}

// Unref decreases the reference count on a GMainLoop object by one. If the result is zero,
// it frees the loop and all associated memory.
func (m *MainLoop) Unref() { C.g_main_loop_unref(m.Instance()) }

// Run a main loop until Quit() is called on the loop. If this is called from the thread of
// the loop's GMainContext, it will process events from the loop, otherwise it will simply wait.
func (m *MainLoop) Run() { C.g_main_loop_run(m.Instance()) }

// RunError is an alias to Run() except it returns nil as soon as the main loop quits. This is for
// convenience when wanting to use `return mainLoop.RunError()` at the end of a function that
// expects an error.
func (m *MainLoop) RunError() error {
	m.Run()
	return nil
}

// Quit stops a MainLoop from running. Any calls to Run() for the loop will return. Note that
// sources that have already been dispatched when Quit() is called will still be executed.
func (m *MainLoop) Quit() { C.g_main_loop_quit(m.Instance()) }

// IsRunning returns true if this main loop is currently running.
func (m *MainLoop) IsRunning() bool { return gobool(C.g_main_loop_is_running(m.Instance())) }

// GetContext returns the GMainContext for this loop.
func (m *MainLoop) GetContext() *MainContext {
	c := C.g_main_loop_get_context(m.Instance())
	return (*MainContext)(c)
}
