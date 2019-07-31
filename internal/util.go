package internal

import (
	"fmt"
	"strings"

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

// EscapeColumnName will escape a column name if using reserved keyword as column name, returning it in
// surrounded backquotes.
func EscapeColumnName(s string) string {
	if _, ok := reservedKeywords[strings.ToUpper(s)]; ok {
		// return surrounded s with backquotes if reserved keyword
		return fmt.Sprintf("`%s`", s)
	}
	// return s if not reserved keyword
	return s
}
