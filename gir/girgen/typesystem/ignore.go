package typesystem

import (
	"regexp"

	"github.com/go-gst/go-glib/gir"
)

type IgnoreFunc func(parent, self string, attrs gir.InfoAttrs, elements gir.InfoElements) bool

func ignoreOr(sf ...IgnoreFunc) IgnoreFunc {
	if len(sf) == 0 {
		return func(parent, self string, attrs gir.InfoAttrs, elements gir.InfoElements) bool {
			return false
		}
	}
	if len(sf) == 1 {
		return sf[0]
	}
	return func(parent, self string, attrs gir.InfoAttrs, elements gir.InfoElements) bool {
		for _, f := range sf {
			if f(parent, self, attrs, elements) {
				return true
			}
		}

		return false
	}
}

func IgnoreMatching(pattern string) IgnoreFunc {
	girpattern := GIRPattern(pattern)

	return func(parent, self string, attrs gir.InfoAttrs, elements gir.InfoElements) bool {
		return girpattern.Matches(parent, self)
	}
}

func IgnoreByRegex(pattern string) IgnoreFunc {
	re := regexp.MustCompile(pattern)
	return func(parent, self string, attrs gir.InfoAttrs, elements gir.InfoElements) bool {
		if parent == "" {
			return re.Match([]byte(self))
		}

		return re.Match([]byte(parent + "." + self))
	}
}
