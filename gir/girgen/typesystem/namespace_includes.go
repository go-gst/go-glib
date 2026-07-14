package typesystem

import (
	"cmp"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"slices"

	"github.com/go-gst/go-glib/gir"
)

// namespaceWithIncludes wraps a gir.Namespace and contains resolved includes
//
// it is used as a preprocessing step before resolving all the types in the namespace
type namespaceWithIncludes struct {
	versionedName versionedName
	// resolveIndex is used to keep track of the order in which the namespaces are resolved
	// lower index means the namespace is resolved earlier, meaning it has less dependencies
	// and must be processed first
	resolveIndex int
	includes     map[string]*namespaceWithIncludes

	repository *gir.Repository
	*gir.Namespace
}

type repoWithIncludes struct {
	filename string
	*gir.Repository

	namespaces []*namespaceWithIncludes
}

type versionedName struct {
	name    string
	version gir.Version
}

func (v versionedName) String() string {
	return fmt.Sprintf("%s-%s", v.name, v.version)
}

func resolveNamespaceIncludes(repos gir.Repositories) []*repoWithIncludes {
	namespacesByName := make(map[versionedName]*namespaceWithIncludes)
	outRepos := make([]*repoWithIncludes, 0, len(repos))

	for filename, repo := range repos {
		includes := repoPrefilledIncludes(repo)

		outRepo := &repoWithIncludes{
			filename:   filename,
			Repository: repo,
		}

		for _, ns := range repo.Namespaces {
			versioned := versionedName{
				name:    ns.Name,
				version: ns.Version,
			}

			namespace := &namespaceWithIncludes{
				versionedName: versioned,
				includes:      includes,

				repository: repo,
				Namespace:  ns,
			}

			namespacesByName[versioned] = namespace
			outRepo.namespaces = append(outRepo.namespaces, namespace)
		}

		outRepos = append(outRepos, outRepo)
	}

	// here we populate the includes, and also the transitively included namespaces
	// we want to have pointers to the actual namespaces, not just the name and version
	// so we resolve the entire include tree

	// keep track of visited namespaces for the include resolution
	visited := make(map[versionedName]struct{})

	var collectIncludes func(n *namespaceWithIncludes) map[string]*namespaceWithIncludes

	// resolvedNamespaces keeps track of how many namespaces we have resolved
	// this is used for the resolveIndex
	resolvedNamespaces := 0

	collectIncludes = func(n *namespaceWithIncludes) map[string]*namespaceWithIncludes {
		if _, ok := visited[n.versionedName]; ok {
			slog.Info("includes already resolved", "name", n.versionedName)
			return n.includes // includes are already resolved
		}

		slog.Info("collecting includes for namespace", "name", n.versionedName)

		visited[n.versionedName] = struct{}{}

		// the includes are still only prefilled, meaning we have the name and version,
		// but not the actual pointer to the namespace

		resolvedIncludes := make(map[string]*namespaceWithIncludes, len(n.includes))

		for _, inc := range n.includes {
			if incNs, ok := namespacesByName[inc.versionedName]; ok {
				slog.Info("included namespace found", "name", n.versionedName, "include", inc.versionedName)

				resolvedIncludes[inc.versionedName.name] = incNs

				transitiveIncludes := collectIncludes(incNs)

				maps.Copy(resolvedIncludes, transitiveIncludes)
			} else {
				log.Printf("warning: include %s not found in repositories", inc.versionedName)
			}
		}

		n.includes = resolvedIncludes

		n.resolveIndex = resolvedNamespaces
		resolvedNamespaces++

		return resolvedIncludes
	}

	for _, namespace := range namespacesByName {
		collectIncludes(namespace)
	}

	// to make further processing easier we sort the repositories in a way that moves the base dependencies
	// to the front. We tracked the resolveIndex so we know which namespaces were resolved first.

	slices.SortFunc(outRepos, func(a, b *repoWithIncludes) int {
		if len(a.namespaces) != 1 || len(b.namespaces) != 1 {
			panic("expected exactly one namespace per repository for sorting")
		}

		return cmp.Compare(a.namespaces[0].resolveIndex, b.namespaces[0].resolveIndex)
	})

	return outRepos
}

// repoPrefilledIncludes prefills the includes with namespaces name and version, to
// be able to later set it to the actual pointer to the foreign namespace
func repoPrefilledIncludes(r *gir.Repository) map[string]*namespaceWithIncludes {
	m := make(map[string]*namespaceWithIncludes)
	for _, v := range r.Includes {
		m[v.Name] = &namespaceWithIncludes{
			versionedName: versionedName{
				name:    v.Name,
				version: v.Version,
			},
		}
	}
	return m
}
