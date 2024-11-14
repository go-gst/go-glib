package glib

import "runtime/pprof"

var gObjectProfile *pprof.Profile
var closureProfile *pprof.Profile

func init() {
	objects := "go-glib-reffed-objects"
	gObjectProfile = pprof.Lookup(objects)
	if gObjectProfile == nil {
		gObjectProfile = pprof.NewProfile(objects)
	}

	closures := "go-glib-active-closures"
	closureProfile = pprof.Lookup(closures)
	if closureProfile == nil {
		closureProfile = pprof.NewProfile(closures)
	}

}
