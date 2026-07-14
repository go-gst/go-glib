package typesystem

import (
	"fmt"
	"strings"

	"github.com/go-gst/go-glib/gir"
)

type Registry struct {
	Repositories []*Repository
}

// FromRepositories loads all repositories into the registry and resolves all type references.
//
// the repositories must be loaded in the correct order to be able to resolve all types
func FromRepositories(cfg Config, repos gir.Repositories) *Registry {
	r := &Registry{
		Repositories: make([]*Repository, 0, len(repos)),
	}

	withIncludes := resolveNamespaceIncludes(repos)

	for _, repoTmp := range withIncludes {
		repo := &Repository{
			Filename: repoTmp.filename,
		}

		for _, nsTmp := range repoTmp.namespaces {

			ns := r.newNamespace(cfg, nsTmp)

			if ns == nil {
				continue
			}

			repo.Namespaces = append(repo.Namespaces, ns)
		}

		r.Repositories = append(r.Repositories, repo)
	}

	return r
}

// FindNamespaceByName finds the given namespace, e.g. "Gtk-4"
func (r *Registry) FindNamespaceByName(name string) *Namespace {

	parts := strings.Split(name, "-")

	if len(parts) != 2 {
		panic(fmt.Errorf("namespace name %s is not in the format <name>-<version>", name))
	}
	nsName := parts[0]
	nsMajor := parts[1]

	v, err := gir.ParseVersion(nsMajor)

	if err != nil {
		return nil
	}

	return r.findNamespace(versionedName{

		name:    nsName,
		version: v,
	})
}

func (r *Registry) findNamespace(v versionedName) *Namespace {
	for _, repo := range r.Repositories {
		for _, ns := range repo.Namespaces {
			if ns.v == v {
				return ns
			}
		}
	}

	return nil
}

func (r *Registry) Postprocess(post []PostProcessor) {
	for _, f := range post {
		if err := f(r); err != nil {
			panic(err)
		}
	}
}
