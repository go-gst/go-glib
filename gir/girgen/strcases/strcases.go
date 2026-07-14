// Package strcases provides helper functions to convert between string cases,
// such as Pascal Case, snake_case and Go's Mixed Caps, along with various
// special cases.
package strcases

import (
	"log/slog"
	"strings"
	"unicode"
	"unicode/utf8"

	_ "embed"
)

var snakeToPascalSpecialWords = map[string]string{
	// general acromnyms:
	"api":  "API",
	"id":   "ID",
	"ids":  "IDs",
	"uri":  "URI",
	"json": "JSON",
	"ok":   "OK",
	"eof":  "EOF",
	"io":   "IO",
	// encodings:
	"utf8":  "UTF8",
	"utf16": "UTF16",
	"ascii": "ASCII",
	"ucs4":  "UCS4",
	// unicode normalization forms:
	"nfc":  "NFC",
	"nfd":  "NFD",
	"nfkc": "NFKC",
	"nfkd": "NFKD",
	// special casing:
	"foreach": "ForEach",
	// hashes:
	"md5":    "MD5",
	"sha1":   "SHA1",
	"sha256": "SHA256",
	"sha384": "SHA384",
	"sha512": "SHA512",
	// gnome names:
	"dbus":      "DBus",
	"gsettings": "GSettings",
	"gtype":     "GType",
	"vfs":       "VFS",
	// gstreamer names:
	"eos": "EOS",
}

// isLower returns true if the string is all lower-cased.
func isLower(s string) bool {
	return strings.IndexFunc(s, unicode.IsUpper) == -1
}

// firstToUpper returns the first letter in upper-case.
func firstToUpper(s string) string {
	if len(s) == 0 {
		return s
	}

	r, sz := utf8.DecodeRuneInString(s)
	if sz > 0 && r != utf8.RuneError {
		return string(unicode.ToUpper(r)) + s[sz:]
	}

	// impossible for non empty UTF-8 string
	panic("firstToUpper: invalid string " + s)
}

// ParamNameToGo turns snake_case to camelCase and makes sure it does not collide with
// Go keywords or built-in types. No special replacements are done
func ParamNameToGo(p string) string {
	snakeWords := snakeWords(p)

	var camelWords []string
	for i, word := range snakeWords {
		if word == "" {
			continue
		}

		word = strings.ToLower(word)

		if i == 0 {
			camelWords = append(camelWords, word)
			continue
		}

		camelWords = append(camelWords, firstToUpper(word))
	}
	camelString := strings.Join(camelWords, "")

	return noGoReserved(camelString)
}

// SnakeToGo converts snake case to Go's special case. If Pascal is true, then
// the first letter is capitalized.
func SnakeToGo(pascal bool, snakeString string) string {
	if !isLower(snakeString) {
		slog.Warn("SnakeToGo: snake case string is not all lower-case", "snakeString", snakeString, "pascal", pascal)
	}

	snakeWords := snakeWords(snakeString)

	var pascalWords []string
	for i, word := range snakeWords {
		if word == "" {
			continue
		}

		word := strings.ToLower(word)

		if i == 0 && !pascal {
			pascalWords = append(pascalWords, word)
			continue
		}

		if special, ok := snakeToPascalSpecialWords[word]; ok {
			pascalWords = append(pascalWords, special)
			continue
		}

		pascalWords = append(pascalWords, firstToUpper(word))
	}
	pascalString := strings.Join(pascalWords, "")

	return noGoReserved(pascalString)
}

// KebabToGo converts kebab case to Go's special case. See SnakeToGo.
func KebabToGo(pascal bool, kebabString string) string {
	return SnakeToGo(pascal, strings.ReplaceAll(kebabString, "-", "_"))
}

// GoKeywords includes Go keywords. This is primarily to prevent collisions with
// meaningful Go words.
var GoKeywords = map[string]string{
	// Keywords.
	"break":       "",
	"default":     "",
	"func":        "fn",
	"interface":   "iface",
	"select":      "sel",
	"case":        "",
	"defer":       "",
	"go":          "",
	"map":         "",
	"struct":      "",
	"chan":        "ch",
	"else":        "",
	"goto":        "",
	"package":     "pkg",
	"switch":      "",
	"const":       "",
	"fallthrough": "",
	"if":          "",
	"range":       "",
	"type":        "typ",
	"continue":    "",
	"for":         "",
	"import":      "",
	"return":      "ret",
	"var":         "",

	// words that may collide with go stdlib packages
	"context": "_context",
	"strings": "_strings",
	"fmt":     "_fmt",
}

// GoBuiltinTypes contains Go built-in types.
var GoBuiltinTypes = map[string]string{
	// Types.
	"bool":       "",
	"byte":       "",
	"complex128": "cmplx",
	"complex64":  "cmplx",
	"error":      "err",
	"float32":    "",
	"float64":    "",
	"int":        "",
	"int16":      "",
	"int32":      "",
	"int64":      "",
	"int8":       "",
	"rune":       "",
	"string":     "str",
	"uint":       "",
	"uint16":     "",
	"uint32":     "",
	"uint64":     "",
	"uint8":      "",
	"uintptr":    "",
}

// ReceiverName returns the first letter in lower-case.
func ReceiverName(p string) string {
	if len(p) == 0 {
		panic("ReceiverName: empty string")
	}

	r, sz := utf8.DecodeRuneInString(p)
	if sz > 0 && r != utf8.RuneError {
		if r == '_' {
			panic("ReceiverName: invalid string with underscore " + p)
		}

		return string(unicode.ToLower(r))
	}

	panic("ReceiverName: invalid string " + p)
}

// Unexport takes a string and returns it with the first letter in lower-case. It also checks for
// reserved Go keywords and built-in types, returning a modified version if necessary.
func Unexport(s string) string {
	if len(s) == 0 {
		return s
	}

	r, sz := utf8.DecodeRuneInString(s)
	if sz > 0 && r != utf8.RuneError {
		s = string(unicode.ToLower(r)) + s[sz:]

		return noGoReserved(s)
	}

	// impossible for non empty UTF-8 string
	panic("Unexport: invalid string " + s)
}

// CGoField formats the C field name to not be confused with a Go keyword.
// See https://golang.org/cmd/cgo/#hdr-Go_references_to_C.
func CGoField(field string) string {
	_, keyword := GoKeywords[field]
	if keyword {
		return "_" + field
	}
	return field
}

// noGoReserved ensures the snake-case string is never a Go keyword.
func noGoReserved(snake string) string {
	s, isKeyword := GoKeywords[snake]
	if isKeyword {
		if s != "" {
			return s
		}
		return "_" + snake
	}

	s, isType := GoBuiltinTypes[snake]
	if isType {
		if s != "" {
			return s
		}
		return "_" + snake
	}

	return snake
}

// snakeWords splits the snake case string into words.
func snakeWords(snake string) []string {
	return strings.Split(snake, "_")
}
