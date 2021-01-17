package glib

// #include "glib.go.h"
import "C"
import "sync"

var registerMutex sync.RWMutex

var registeredTypes = make(map[string]Type)
var registeredClasses = make(map[C.gpointer]GoObjectSubclass)
