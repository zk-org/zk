package strings

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestPrepend(t *testing.T) {
	test := func(text string, prefix string, expected string) {
		assert.Equal(t, Prepend(text, prefix), expected)
	}

	test("", "> ", "")
	test("One line", "> ", "> One line")
	test("One line\nTwo lines", "> ", "> One line\n> Two lines")
	test("One line\nTwo lines\nThree lines", "> ", "> One line\n> Two lines\n> Three lines")
	test("Newline\n", "> ", "> Newline\n")
}

func TestPluralize(t *testing.T) {
	test := func(word string, count int, expected string) {
		assert.Equal(t, Pluralize(word, count), expected)
	}

	test("", 1, "")
	test("", 2, "")
	test("word", -2, "words")
	test("word", -1, "word")
	test("word", 0, "word")
	test("word", 1, "word")
	test("word", 2, "words")
	test("word", 1000, "words")
}

func TestSplitLines(t *testing.T) {
	test := func(text string, expected ...string) {
		assert.Equal(t, SplitLines(text), expected)
	}

	test("")
	test("One line", "One line")
	test("One line\nTwo lines", "One line", "Two lines")
	test("One line\nTwo lines\n\nThree lines", "One line", "Two lines", "", "Three lines")
}

func TestJoinLines(t *testing.T) {
	test := func(text string, expected string) {
		assert.Equal(t, JoinLines(text), expected)
	}

	test("", "")
	test("One line", "One line")
	test("One line\nTwo lines", "One line Two lines")
	test("One line\nTwo lines\n\nThree lines", "One line Two lines  Three lines")
	test("One line\nTwo lines\n Three lines", "One line Two lines  Three lines")
}

func TestJoinInt64(t *testing.T) {
	test := func(ints []int64, expected string) {
		assert.Equal(t, JoinInt64(ints, ","), expected)
	}

	test([]int64{}, "")
	test([]int64{1}, "1")
	test([]int64{1, 2}, "1,2")
	test([]int64{1, 2, 3}, "1,2,3")
}

func TestIsURL(t *testing.T) {
	test := func(text string, expected bool) {
		assert.Equal(t, IsURL(text), expected)
	}

	test("", false)
	test("example.com/", false)
	test("path", false)
	test("http://example.com", true)
	test("https://example.com/dir", true)
	test("http://example.com/dir", true)
	test("ftp://example.com/", true)
}

func TestRemoveDuplicates(t *testing.T) {
	test := func(items []string, expected []string) {
		assert.Equal(t, RemoveDuplicates(items), expected)
	}

	test([]string{}, []string{})
	test([]string{"One"}, []string{"One"})
	test([]string{"One", "Two"}, []string{"One", "Two"})
	test([]string{"One", "Two", "One"}, []string{"One", "Two"})
	test([]string{"Two", "One", "Two", "One"}, []string{"Two", "One"})
	test([]string{"One", "Two", "OneTwo"}, []string{"One", "Two", "OneTwo"})
}

func TestRemoveBlank(t *testing.T) {
	test := func(items []string, expected []string) {
		assert.Equal(t, RemoveBlank(items), expected)
	}

	test([]string{}, []string{})
	test([]string{"One"}, []string{"One"})
	test([]string{"One", "Two"}, []string{"One", "Two"})
	test([]string{"One", "Two", ""}, []string{"One", "Two"})
	test([]string{"Two", "One", " "}, []string{"Two", "One"})
	test([]string{"One", "Two", "	  "}, []string{"One", "Two"})
}

func TestExpandWhitespaceLiterals(t *testing.T) {
	test := func(s string, expected string) {
		assert.Equal(t, ExpandWhitespaceLiterals(s), expected)
	}

	test(`nothing`, "nothing")
	test(`newline\ntab\t`, "newline\ntab\t")
}

func TestContains(t *testing.T) {
	test := func(items []string, s string, expected bool) {
		assert.Equal(t, Contains(items, s), expected)
	}

	test([]string{}, "", false)
	test([]string{}, "none", false)
	test([]string{"one"}, "none", false)
	test([]string{"one"}, "one", true)
	test([]string{"one", "two"}, "one", true)
	test([]string{"one", "two"}, "three", false)
}

func TestWordAt(t *testing.T) {
	test := func(s string, pos int, expected string) {
		assert.Equal(t, WordAt(s, pos), expected)
	}

	test("", 0, "")
	test("		  ", 2, "")
	test("word", 2, "word")
	test("		word	", 4, "word")
	test("one two three", 4, "two")
	test("one two three", 5, "two")
	test("one two three", 7, "two")
	test("one two-third three", 5, "two-third")
	test("one two,three", 5, "two")
	test("one two;three", 5, "two")
	test("one [two] three", 5, "two")
	test("one \"two\" three", 5, "two")
	test("one 'two' three", 5, "two")
	test("one\ntwo\nthree", 5, "two")
	test("one\ttwo\tthree", 5, "two")
	test("one @:~two three", 5, "@:~two")
}
