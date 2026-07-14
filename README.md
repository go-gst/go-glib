# go-glib

Glib and gobject bindings and bindings generator for go.

These are intended to be used together with https://github.com/go-gst/go-gst

[![godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/go-gst/go-glib)
[![GoReportCard](https://goreportcard.com/badge/github.com/go-gst/go-glib)](https://goreportcard.com/report/github.com/go-gst/go-glib)

## Where are the v1.X.X versions?

In https://github.com/go-gst/go-glib/pull/32 this repo was migrated to using a generator from GIR files. This makes all code in this repo:

* safer
* more aligned with the GStreamer functions
* way easier to maintain

The old code isn't gone, you can always pin your versions on the old commits. The tags have been retracted though, so you may end up seeing some logs from the go toolchain complaining.

There are some migrations needed, as this is a breaking change, but mostly this is simple syntax or function naming. The underlying GStreamer logic does not change from this. See the [go-gst examples](https://github.com/go-gst/go-gst/tree/main/examples) for some reference.
