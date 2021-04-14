package fts5

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestConvertQuery(t *testing.T) {
	test := func(query, expected string) {
		assert.Equal(t, ConvertQuery(query), expected)
	}

	// Quotes
	test(`foo`, `"foo"`)
	test(`foo bar`, `"foo" "bar"`)
	test(`foo   bar`, `"foo"   "bar"`)
	test(`"foo"`, `"foo"`)
	test(`"foo bar"`, `"foo bar"`)
	test(`"foo bar" qux`, `"foo bar" "qux"`)
	test(`foo "bar qux`, `"foo" "bar qux"`)

	// Conjunction
	test(`foo AND bar`, `"foo" AND "bar"`)
	test(`foo AN bar`, `"foo" "AN" "bar"`)
	test(`foo ANT bar`, `"foo" "ANT" "bar"`)
	test(`foo "AND" bar`, `"foo" "AND" "bar"`)
	test(`"foo AND bar"`, `"foo AND bar"`)

	// Disjunction
	test(`foo OR bar`, `"foo" OR "bar"`)
	test(`foo | bar`, `"foo"  OR  "bar"`)
	test(`foo|bar`, `"foo" OR "bar"`)
	test(`"foo | bar"`, `"foo | bar"`)

	// Negation
	test(`foo NOT bar`, `"foo" NOT "bar"`)
	test(`foo -bar`, `"foo"  NOT "bar"`)
	test(`"foo -bar"`, `"foo -bar"`)
	test(`foo-bar`, `"foo-bar"`)

	// Grouping
	test(`(foo AND bar) OR qux`, `("foo" AND "bar") OR "qux"`)

	// Special characters
	test(`foo/bar`, `"foo/bar"`)
	test(`foo;bar`, `"foo;bar"`)
	test(`foo,bar`, `"foo,bar"`)
	test(`foo&bar`, `"foo&bar"`)
	test(`foo's bar`, `"foo's" "bar"`)

	// Prefix queries
	test(`foo ba*`, `"foo" "ba"*`)
	test(`foo ba* qux`, `"foo" "ba"* "qux"`)
	test(`"foo ba"*`, `"foo ba"*`)
	test(`"foo ba*"`, `"foo ba*"`)
	test(`(foo ba*)`, `("foo" "ba"*)`)
	test(`foo*bar`, `"foo*bar"`)
	test(`"foo*bar"`, `"foo*bar"`)

	// Column filters
	test(`col:foo bar`, `col:"foo" "bar"`)
	test(`foo col:bar`, `"foo" col:"bar"`)
	test(`foo "col:bar"`, `"foo" "col:bar"`)
	test(`":foo"`, `":foo"`)
	test(`-col:foo bar`, ` NOT col:"foo" "bar"`)
	test(`col:(foo bar)`, `col:("foo" "bar")`)

	// First token
	test(`^foo`, `^"foo"`)
	test(`^foo bar`, `^"foo" "bar"`)
	test(`foo ^bar`, `"foo" ^"bar"`)
	test(`^"foo bar"`, `^"foo bar"`)
	test(`"foo ^bar"`, `"foo ^bar"`)
	test(`col:^foo`, `col:^"foo"`)

	// FTS5's + is ignored
	test(`foo + bar`, `"foo"  "bar"`)
	test(`"foo + bar"`, `"foo + bar"`)
	test(`"+foo"`, `"+foo"`)

	// NEAR is not supported
	test(`NEAR(foo, bar, 4)`, `"NEAR"("foo," "bar," "4")`)
}
