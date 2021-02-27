package note

import (
	"testing"

	"github.com/mickael-menu/zk/util/test/assert"
)

func TestSorterFromString(t *testing.T) {
	test := func(str string, expectedField SortField, expectedAscending bool) {
		actual, err := SorterFromString(str)
		assert.Nil(t, err)
		assert.Equal(t, actual, Sorter{Field: expectedField, Ascending: expectedAscending})
	}

	test("c", SortCreated, false)
	test("c+", SortCreated, true)
	test("created", SortCreated, false)
	test("created-", SortCreated, false)
	test("created+", SortCreated, true)

	test("m", SortModified, false)
	test("modified", SortModified, false)
	test("modified+", SortModified, true)

	test("p", SortPath, true)
	test("path", SortPath, true)
	test("path-", SortPath, false)

	test("t", SortTitle, true)
	test("title", SortTitle, true)
	test("title-", SortTitle, false)

	test("r", SortRandom, true)
	test("random", SortRandom, true)
	test("random-", SortRandom, false)

	test("wc", SortWordCount, true)
	test("word-count", SortWordCount, true)
	test("word-count-", SortWordCount, false)

	_, err := SorterFromString("foobar")
	assert.Err(t, err, "foobar: unknown sorting term")
}

func TestSortersFromStrings(t *testing.T) {
	test := func(strs []string, expected []Sorter) {
		actual, err := SortersFromStrings(strs)
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	}

	test([]string{}, []Sorter{})

	test([]string{"created"}, []Sorter{
		{Field: SortCreated, Ascending: false},
	})

	// It is parsed in reverse order to be able to override sort criteria set
	// in aliases.
	test([]string{"c+", "title", "random"}, []Sorter{
		{Field: SortRandom, Ascending: true},
		{Field: SortTitle, Ascending: true},
		{Field: SortCreated, Ascending: true},
	})

	_, err := SortersFromStrings([]string{"c", "foobar"})
	assert.Err(t, err, "foobar: unknown sorting term")
}
