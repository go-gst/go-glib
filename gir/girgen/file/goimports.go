package file

import (
	"cmp"
	"io"
	"slices"
)

// goImports maps the package name to its import alias
type goImports map[singleImport]struct{}

type singleImport struct {
	pkg   string
	alias string

	isStd bool
}

func (gi goImports) formatted() io.Reader {
	if len(gi) == 0 {
		return empty
	}

	parts := make([]io.Reader, 0, len(gi)+2)

	parts = append(parts, str("import (\n"))

	std, other := gi.split()

	parts = append(parts, importReader(std)...)

	if len(std) > 0 && len(other) > 0 {
		parts = append(parts, str("\n"))
	}

	parts = append(parts, importReader(other)...)

	parts = append(parts, str(")\n"))

	return io.MultiReader(parts...)
}

func importReader(imports []singleImport) []io.Reader {
	var parts []io.Reader
	for _, imp := range imports {
		parts = append(parts, str("\t"))

		if imp.alias != "" {
			parts = append(parts,
				str(imp.alias),
				str(" "),
			)
		}

		parts = append(parts,
			str(`"`),
			str(imp.pkg),
			str("\"\n"),
		)
	}

	return parts
}

func (gi goImports) split() ([]singleImport, []singleImport) {
	std := make([]singleImport, 0)
	other := make([]singleImport, 0)

	for pkg := range gi {
		if pkg.isStd {
			std = append(std, pkg)
		} else {
			other = append(other, pkg)
		}
	}

	// sort alphabetically, since map ordering will change:

	slices.SortFunc(std, func(a, b singleImport) int {
		return cmp.Compare(a.pkg, b.pkg)
	})

	slices.SortFunc(other, func(a, b singleImport) int {
		return cmp.Compare(a.pkg, b.pkg)
	})

	return std, other
}
func (gi goImports) add(pkg string, alias string, isStd bool) {
	if pkg == "" {
		return
	}

	si := singleImport{
		pkg:   pkg,
		alias: alias,
		isStd: isStd,
	}

	if _, ok := gi[si]; ok {
		return
	}

	gi[si] = struct{}{}
}
