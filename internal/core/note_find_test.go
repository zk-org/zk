package core

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNoteSorterFromString(t *testing.T) {
	test := func(str string, expectedField NoteSortField, expectedAscending bool) {
		actual, err := NoteSorterFromString(str)
		assert.Nil(t, err)
		assert.Equal(t, actual, NoteSorter{Field: expectedField, Ascending: expectedAscending})
	}

	test("c", NoteSortCreated, false)
	test("c+", NoteSortCreated, true)
	test("created", NoteSortCreated, false)
	test("created-", NoteSortCreated, false)
	test("created+", NoteSortCreated, true)

	test("m", NoteSortModified, false)
	test("modified", NoteSortModified, false)
	test("modified+", NoteSortModified, true)

	test("p", NoteSortPath, true)
	test("path", NoteSortPath, true)
	test("path-", NoteSortPath, false)

	test("t", NoteSortTitle, true)
	test("title", NoteSortTitle, true)
	test("title-", NoteSortTitle, false)

	test("r", NoteSortRandom, true)
	test("random", NoteSortRandom, true)
	test("random-", NoteSortRandom, false)

	test("wc", NoteSortWordCount, true)
	test("word-count", NoteSortWordCount, true)
	test("word-count-", NoteSortWordCount, false)

	_, err := NoteSorterFromString("foobar")
	assert.Err(t, err, "foobar: unknown sorting term")
}

func TestSortersFromStrings(t *testing.T) {
	test := func(strs []string, expected []NoteSorter) {
		actual, err := NoteSortersFromStrings(strs)
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	}

	test([]string{}, []NoteSorter{})

	test([]string{"created"}, []NoteSorter{
		{Field: NoteSortCreated, Ascending: false},
	})

	// It is parsed in reverse order to be able to override sort criteria set
	// in aliases.
	test([]string{"c+", "title", "random"}, []NoteSorter{
		{Field: NoteSortRandom, Ascending: true},
		{Field: NoteSortTitle, Ascending: true},
		{Field: NoteSortCreated, Ascending: true},
	})

	_, err := NoteSortersFromStrings([]string{"c", "foobar"})
	assert.Err(t, err, "foobar: unknown sorting term")
}

func TestMatchStrategyFromString(t *testing.T) {
	test := func(str string, expected MatchStrategy) {
		actual, err := MatchStrategyFromString(str)
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	}

	test("f", MatchStrategyFts)
	test("fts", MatchStrategyFts)

	test("r", MatchStrategyRe)
	test("re", MatchStrategyRe)
	test("grep", MatchStrategyRe)

	test("e", MatchStrategyExact)
	test("exact", MatchStrategyExact)

	_, err := MatchStrategyFromString("foobar")
	assert.Err(t, err, "foobar: unknown match strategy\ntry fts (full-text search), re (regular expression) or exact")
}
