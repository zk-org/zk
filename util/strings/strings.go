package strings

import (
	"bufio"
	"strings"
)

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

// Pluralize adds an `s` at the end of word if the count is more than 1.
func Pluralize(word string, count int) string {
	if word == "" || (count >= -1 && count <= 1) {
		return word
	} else {
		return word + "s"
	}
}

// SplitLines splits a string by the newlines character in a portable way
// Using only `strings.Split(s, "\n")` doesn't work on Windows.
func SplitLines(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// JoinLines joins each lines of the given string by replacing the newlines by
// a single space.
func JoinLines(s string) string {
	return strings.Join(SplitLines(s), " ")
}
