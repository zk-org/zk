package fts5

import "strings"

// ConvertQuery transforms a Google-like query into a SQLite FTS5 one.
func ConvertQuery(query string) string {
	out := ""

	// List of tokens which won't be automatically quoted in the output query.
	passthroughTokens := map[string]bool{
		"AND": true,
		"OR":  true,
		"NOT": true,
	}

	// Whitespaces and parentheses are term separators outside explicit quotes.
	termSeparators := map[rune]bool{
		' ':  true,
		'\t': true,
		'\n': true,
		'(':  true,
		')':  true,
	}

	// Indicates whether the current term was explicitely quoted in the query.
	inQuote := false
	// Current term being read.
	term := ""

	// Finishes the current term and write it to the output after quoting it.
	closeTerm := func() {
		if term == "" {
			return
		}

		if !inQuote && passthroughTokens[term] {
			out += term
		} else {
			// If the term has a wildcard suffix, it is a prefix token. We make
			// sure that the * is not quoted or it will be ignored by the FTS5
			// tokenizer.
			isPrefixToken := !inQuote && strings.HasSuffix(term, "*")
			if isPrefixToken {
				term = strings.TrimSuffix(term, "*")
			}
			out += `"` + term + `"`
			if isPrefixToken {
				out += "*"
			}
		}

		term = ""
	}

	for _, c := range query {
		switch {
		// Explicit quotes.
		case c == '"':
			if inQuote { // We are already in a quoted term? Then it's a closing quote.
				closeTerm()
			}
			inQuote = !inQuote

		// Passthrough for ^ and * when they are at the start of a term, to allow:
		//   ^foo -> ^"foo"
		//   "foo"* -> "foo"*
		case term == "" && (c == '^' || c == '*'):
			out += string(c)

		// Passthrough for FTS5's column filters, e.g.
		//   col:foo -> col:"foo"
		case !inQuote && c == ':':
			out += term + string(c)
			term = ""

		// - is an alias to NOT, but only at the start of a term, to allow
		// compound words such as "well-known"
		case c == '-' && term == "":
			out += " NOT "

		// | is an alias to OR.
		case !inQuote && c == '|':
			closeTerm()
			out += " OR "

		// FTS5's + is ignored because it doesn't bring much to the syntax,
		// compared to explicit quotes.
		case !inQuote && c == '+' && term == "":
			break

		// Term separators outside explicit quotes terminates the current term.
		case !inQuote && termSeparators[c]:
			closeTerm()
			out += string(c)

		default:
			term += string(c)
		}
	}

	closeTerm()
	return out
}
