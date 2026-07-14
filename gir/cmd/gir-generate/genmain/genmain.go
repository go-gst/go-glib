package genmain

import (
	"flag"
	"fmt"
	"log"
	"maps"

	"github.com/go-gst/go-glib/gir"
	"github.com/go-gst/go-glib/gir/girgen/file"
	"github.com/go-gst/go-glib/gir/girgen/generators"
	"github.com/go-gst/go-glib/gir/girgen/typesystem"
)

var (
	Output  string
	Verbose bool
	ListPkg bool
	CgoLink bool
)

func init() {
	flag.StringVar(&Output, "o", "", "output directory to mkdir in")
}

// ParseFlag calls flag.Parse() and initializes external global options.
func ParseFlag() {
	flag.Parse()

	if !ListPkg && Output == "" {
		log.Fatalln("Missing -o output directory.")
	}
}

type Package struct {
	// Name is the pkg-config name.
	Name string
	// Namespaces is the possible namespaces within it. Refer to
	// ./cmd/gir_namespaces.
	Namespaces []string
}

// Data contains generation data that genmain uses to generate.
type Data struct {
	// Module is the Go Module name that the generator is running for. An
	// example is "github.com/go-gst/go-glib/pkg".
	Module string
	// GirFiles contains a map of Gir file names to their contents.
	GirFiles gir.RawFiles

	// Preprocessors defines a list of preprocessors that the main generator
	// will use. It's mostly used for renaming colliding types/identifiers.
	Preprocessors []gir.Preprocessor

	// Config is the typesystem.Config that will be used for resolving all types.
	Config typesystem.Config

	// Postprocessors will run on the resolved typesystem before the files are written
	Postprocessors []typesystem.PostProcessor

	// Documentation provides a constructor for documentation generators. If this is nil it defaults to
	// [generators.NewInlineGoDocGenerator].
	Documentation func(*typesystem.Namespace, typesystem.Documented) generators.DocGenerator
}

// Run runs the application. The given [Data] are overlaid and the last one is used as the
// source for generating the code
func Run(datas ...Data) {
	ParseFlag()

	if len(datas) == 0 {
		log.Fatalln("No data provided to run the generator.")
	}

	log.Println("loading packages...")

	// generateData is the last data in the list, which is used to generate the code.
	generateData := datas[len(datas)-1]

	if generateData.Documentation == nil {
		generateData.Documentation = generators.NewInlineGoDocGenerator
	}

	allRepos := make(gir.Repositories)

	importBaseURIs := map[string]string{}

	mergedTSConfig := typesystem.Config{}
	var allPreprocessors []gir.Preprocessor
	var allPostProcessors []typesystem.PostProcessor

	for _, d := range datas {
		repos, err := gir.ParseAll(d.GirFiles)

		if err != nil {
			log.Fatalln("failed to parse GIR files:", err)
		}

		maps.Copy(allRepos, repos)

		for _, r := range repos {
			for _, ns := range r.Namespaces {
				importBaseURIs[fmt.Sprintf("%s-%d", ns.Name, ns.Version.Major)] = d.Module
			}
		}

		mergedTSConfig = mergedTSConfig.Combine(d.Config)
		allPreprocessors = append(allPreprocessors, d.Preprocessors...)
		allPostProcessors = append(allPostProcessors, d.Postprocessors...)
	}

	err := CleanGeneratedFiles(Output)

	if err != nil {
		log.Fatalln("failed to clean output directory:", err)
	}

	gir.ApplyPreprocessors(allRepos, allPreprocessors)

	ts := typesystem.FromRepositories(mergedTSConfig, allRepos)

	ts.Postprocess(allPostProcessors)

	var namespacesToGenerate []generators.Generator

	// the typesystem contains all repositories, but we only want to generate
	// those that are in the generateData.GirFiles
	for _, repo := range ts.Repositories {
		for filename := range generateData.GirFiles {
			if filename == repo.Filename {
				for _, ns := range repo.Namespaces {
					genconfig := &generators.Config{
						DocGeneratorFactory: generateData.Documentation,
						Namespace:           ns,
					}

					namespacesToGenerate = append(
						namespacesToGenerate,
						generators.NewNamespaceGenerator(genconfig),
					)
				}

				break
			}
		}
	}

	for _, g := range namespacesToGenerate {
		w := file.NewPackage(Output, importBaseURIs)

		g.Generate(w)

		// in theory we can add a pre commit stage here, but I don't know if this is useful

		err := w.Commit()

		if err != nil {
			panic(err)
		}
	}
}
