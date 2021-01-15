package strings

import (
	"testing"

	"github.com/mickael-menu/zk/util/test/assert"
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
