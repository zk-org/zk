package sqlite

import "strings"

// escapeLikeTerm returns the given term after escaping any LIKE-significant
// characters with the given escapeChar.
// This is meant to be used with the ESCAPE keyword:
// https://www.sqlite.org/lang_expr.html
func escapeLikeTerm(term string, escapeChar rune) string {
	escape := func(term string, char string) string {
		return strings.ReplaceAll(term, char, string(escapeChar)+char)
	}
	return escape(escape(escape(term, string(escapeChar)), "%"), "_")
}
