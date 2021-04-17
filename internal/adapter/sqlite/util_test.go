package sqlite

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestEscapeLikeTerm(t *testing.T) {
	test := func(term string, escapeChar rune, expected string) {
		assert.Equal(t, escapeLikeTerm(term, escapeChar), expected)
	}

	test("foo bar", '@', "foo bar")
	test("foo%bar_with@", '@', "foo@%bar@_with@@")
	test(`foo%bar_with\`, '\\', `foo\%bar\_with\\`)
}
