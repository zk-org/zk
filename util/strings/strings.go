package strings

import "strings"

// Prepend prefixes each lines of a string with the given prefix.
// It can be used to indent or quote (> ) a paragraph, for example.
func Prepend(text string, prefix string) string {
	if text == "" || prefix == "" {
		return text
	}

	lines := strings.SplitAfter(text, "\n")
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(append([]string{""}, lines...), prefix)
}
