package generators

import "github.com/go-gst/go-glib/gir/girgen/typesystem"

type Config struct {
	Namespace           *typesystem.Namespace
	DocGeneratorFactory func(namespaces *typesystem.Namespace, documented typesystem.Documented) DocGenerator
}

func (c *Config) DocGenerator(documented typesystem.Documented) DocGenerator {
	return c.DocGeneratorFactory(c.Namespace, documented)
}
