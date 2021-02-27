package icu

// EscapePattern adds backslash escapes to protect any characters that would
// match as ICU pattern metacharacters.
//
// http://userguide.icu-project.org/strings/regexp
func EscapePattern(s string) string {
	out := ""

	for _, c := range s {
		switch c {
		case '\\', '.', '^', '$', '(', ')', '[', ']', '{', '}', '|', '*', '+', '?':
			out += `\`
		}
		out += string(c)
	}

	return out
}
