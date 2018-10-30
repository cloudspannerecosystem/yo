package internal

import (
	"github.com/gedex/inflector"
	"github.com/knq/snaker"
)

// reverseIndexRune finds the last rune r in s, returning -1 if not present.
func reverseIndexRune(s string, r rune) int {
	if s == "" {
		return -1
	}

	rs := []rune(s)
	for i := len(rs) - 1; i >= 0; i-- {
		if rs[i] == r {
			return i
		}
	}

	return -1
}

// SinguralizeIdentifier will singularize a identifier, returning it in
// CamelCase.
func SingularizeIdentifier(s string) string {
	if i := reverseIndexRune(s, '_'); i != -1 {
		s = s[:i] + "_" + inflector.Singularize(s[i+1:])
	} else {
		s = inflector.Singularize(s)
	}

	// return snaker.SnakeToCamelIdentifier(s)
	return snaker.ForceCamelIdentifier(s)
}
