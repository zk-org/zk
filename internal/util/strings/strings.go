package strings

import (
	"bufio"
	"net/url"
	"regexp"
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

// RemoveBlank keeps only non-empty strings in the source.
func RemoveBlank(strs []string) []string {
	if strs == nil {
		return nil
	}

	res := make([]string, 0)
	for _, val := range strs {
		if strings.TrimSpace(val) != "" {
			res = append(res, val)
		}
	}

	return res
}

// Expand literal escaped whitespace characters in the given string to their
// actual character.
func ExpandWhitespaceLiterals(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}

// Contains returns whether the given slice of strings contains the given
// string.
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// WordAt returns the word found at the given character position.
// Credit https://github.com/aca/neuron-language-server/blob/450a7cff71c14e291ee85ff8a0614fa9d4dd5145/utils.go#L13
func WordAt(str string, index int) string {
	wordIdxs := wordRegex.FindAllStringIndex(str, -1)
	for _, wordIdx := range wordIdxs {
		if wordIdx[0] <= index && index <= wordIdx[1] {
			return str[wordIdx[0]:wordIdx[1]]
		}
	}

	return ""
}

var wordRegex = regexp.MustCompile(`[^ \t\n\f\r,;\[\]\"\']+`)

func CopyList(list []string) []string {
	out := make([]string, len(list))
	copy(out, list)
	return out
}

func ByteIndexToRuneIndex(s string, i int) int {
	res := 0
	for j, _ := range s {
		if j >= i {
			break
		}
		res += 1
	}
	return res
}
