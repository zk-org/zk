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
