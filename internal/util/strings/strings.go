package strings

import (
	"bufio"
	"net/url"
	"strconv"
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

// JoinInt64 joins a list of int64 into a single string with the given
// delimiter.
func JoinInt64(ints []int64, delimiter string) string {
	strs := make([]string, 0)
	for _, i := range ints {
		strs = append(strs, strconv.FormatInt(i, 10))
	}
	return strings.Join(strs, delimiter)
}

// IsURL returns whether the given string is a valid URL.
func IsURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// RemoveDuplicates keeps only unique strings in the source.
func RemoveDuplicates(strings []string) []string {
	if strings == nil {
		return nil
	}

	check := make(map[string]bool)
	res := make([]string, 0)
	for _, val := range strings {
		if _, ok := check[val]; ok {
			continue
		}
		check[val] = true
		res = append(res, val)
	}

	return res
}

// InList returns whether the string is part of the given list of strings.
func InList(strings []string, s string) bool {
	for _, c := range strings {
		if c == s {
			return true
		}
	}
	return false
}

// Expand literal escaped whitespace characters in the given string to their
// actual character.
func ExpandWhitespaceLiterals(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}
