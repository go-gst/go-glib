package gir

import (
	"log"
	"regexp"
)

// Preprocessor describes something that can preprocess anything in the given
// list of repositories. This is useful for renaming functions, classes or
// anything else.
type Preprocessor interface {
	// Preprocess goes over the given list of repos, changing what's necessary.
	Preprocess(repos Repositories)
}

// ApplyPreprocessors applies the given list of preprocessors onto the given
// list of GIR repositories.
func ApplyPreprocessors(repos Repositories, preprocs []Preprocessor) {
	for _, preproc := range preprocs {
		preproc.Preprocess(repos)
	}
}

// PreprocessorFunc is a helper function to satisfy the Preprocessor interface.
type PreprocessorFunc func(Repositories)

// Preprocess calls f.
func (f PreprocessorFunc) Preprocess(repos Repositories) {
	f(repos)
}

// RemovePkgconfig removes the given pkgconfig from the GIR file with the given name.
func RemovePkgconfig(filename string, pkgconfig string) Preprocessor {
	return PreprocessorFunc(func(repos Repositories) {
		repo := repos.Repository(filename)

		if repo == nil {
			log.Fatalf("RemovePkgconfig: repository %q not found", filename)
		}

		old := repo.Packages

		repo.Packages = repo.Packages[:0]

		for _, pkg := range old {
			if pkg.Name != pkgconfig {
				repo.Packages = append(repo.Packages, pkg)
			}
		}
	})
}

func RemoveCIncludes(filename string, regexes ...string) Preprocessor {
	return PreprocessorFunc(func(repos Repositories) {
		repo := repos.Repository(filename)

		if repo == nil {
			log.Fatalf("RemoveCIncludes: repository %q not found", filename)
		}

		old := repo.CIncludes

		repo.CIncludes = repo.CIncludes[:0]

	nextinclude:
		for _, inc := range old {
			include := inc.Name
			for _, regex := range regexes {
				if matched, _ := regexp.MatchString(regex, include); matched {
					continue nextinclude
				}
			}
			repo.CIncludes = append(repo.CIncludes, inc)
		}
	})
}

// MapMembers allows you to freely change any property of the members of the given enum or bitfield type.
// it can be used to fix faulty girs, but also to change the behavior of the generator.
func MapMembers(enumOrBitfieldType string, fn func(member *Member)) Preprocessor {
	return PreprocessorFunc(func(repos Repositories) {
		result := repos.FindFullType(enumOrBitfieldType)
		if result == nil {
			log.Panicf("GIR enum or bitfield %q not found", enumOrBitfieldType)
			return
		}

		switch v := result.(type) {
		case *Enum:
			for _, member := range v.Members {
				fn(member)
			}
		case *Bitfield:
			for _, member := range v.Members {
				fn(member)
			}
		default:
			log.Panicf("GIR type %T is not enum or bitfield", result)
		}
	})
}

type typeRenamer struct {
	from, to string
}

// TypeRenamer creates a new filter matcher that renames a type. The given GIR
// type must contain the versioned namespace, like "Gtk3.Widget" but the given
// name must not. The GIR type is absolutely matched, similarly to
// AbsoluteFilter.
func TypeRenamer(girType, newName string) Preprocessor {
	return typeRenamer{
		from: girType,
		to:   newName,
	}
}

func (ren typeRenamer) Preprocess(repos Repositories) {
	result := repos.FindFullType(ren.from)
	if result == nil {
		log.Panicf("GIR type %q not found", ren.from)
		return
	}

	// Set the new name:
	switch v := result.(type) {
	case *Class:
		v.Name = ren.to
	case *Interface:
		v.Name = ren.to
	case *Record:
		v.Name = ren.to
	case *Enum:
		v.Name = ren.to
	case *Bitfield:
		v.Name = ren.to
	case *Union:
		v.Name = ren.to
	case *Function:
		v.Name = ren.to
	case *Callback:
		v.Name = ren.to
	case *Alias:
		v.Name = ren.to
	case *Constant:
		v.Name = ren.to
	case *Annotation:
		v.Name = ren.to
	default:
		log.Panicf("GIR type %T is not a type that can be renamed", result)
	}
}

type modifyCallable struct {
	girType string
	modFunc func(*CallableAttrs)
}

// MustIntrospect forces the given type to be introspectable.
func MustIntrospect(girType string) Preprocessor {
	return ModifyCallable(girType, func(c *CallableAttrs) {
		t := new(bool)
		*t = true
		c.Introspectable = t
	})
}

// ModifyCallable is a preprocessor that modifies an existing callable. It only
// does Function or Callback.
func ModifyCallable(girType string, f func(c *CallableAttrs)) Preprocessor {
	return modifyCallable{
		girType: girType,
		modFunc: f,
	}
}

// RenameCallable renames a callable using ModifyCallable.
func RenameCallable(girType, newName string) Preprocessor {
	return ModifyCallable(girType, func(c *CallableAttrs) {
		c.Name = newName
	})
}

// ModifyParamDirections wraps ModifyCallable to conveniently override the
// parameters' directions.
func ModifyParamDirections(girType string, dirOverrides map[string]string) Preprocessor {
	return ModifyCallable(girType, func(c *CallableAttrs) {
		for name, dir := range dirOverrides {
			param := c.FindParameter(name)
			if param == nil {
				log.Panicf("cannot find parameter %s for %s", name, girType)
			}
			param.Direction = dir
		}
	})
}

func (m modifyCallable) Preprocess(repos Repositories) {
	result := repos.FindFullType(m.girType)
	if result == nil {
		log.Panicf("GIR type %q not found", m.girType)
		return
	}

	switch v := result.(type) {
	case *Constructor:
		m.modFunc(v.CallableAttrs)
		return
	case *Method:
		m.modFunc(v.CallableAttrs)
		return
	case *VirtualMethod:
		m.modFunc(v.CallableAttrs)
		return
	case *Function:
		m.modFunc(v.CallableAttrs)
		return
	case *Callback:
		m.modFunc(v.CallableAttrs)
		return
	}

	log.Panicf("GIR type %q has no callable", m.girType)
}

var signalMatcherRe = regexp.MustCompile(`(.*)\.(.*)::(.*)`)

// ModifySignal is like ModifyCallable, except it only works on signals from
// classes and interfaces. The GIR type must be "package.class::signal-name".
func ModifySignal(girType string, f func(c *Signal)) Preprocessor {
	parts := signalMatcherRe.FindStringSubmatch(girType)
	if len(parts) != 4 {
		log.Panicf("GIR signal type %q invalid", girType)
	}

	return PreprocessorFunc(func(repos Repositories) {
		result := repos.FindFullType(parts[1] + "." + parts[2])
		if result == nil {
			log.Panicf("GIR type %q not found", girType)
			return
		}

		switch v := result.(type) {
		case *Class:
			for _, signal := range v.Signals {
				if signal.Name == parts[3] {
					f(signal)
					return
				}
			}
		case *Interface:
			for _, signal := range v.Signals {
				if signal.Name == parts[3] {
					f(signal)
					return
				}
			}
		}
	})
}
