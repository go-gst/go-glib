package main

import (
	"github.com/go-gst/go-glib/gir/cmd/gir-generate/gendata"
	"github.com/go-gst/go-glib/gir/cmd/gir-generate/genmain"
)

func main() {
	genmain.Run(gendata.Main)
}
