package sqlite

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/paths"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNoteDAOIndexed(t *testing.T) {
	testNoteDAOWithFixtures(t, "", func(tx Transaction, dao *NoteDAO) {
		for _, note := range []core.Note{
			{
				Path:     "a.md",
				Modified: time.Date(2020, 1, 20, 8, 52, 42, 0, time.UTC),
			},
			{
				Path:     "dir1/a.md",
				Modified: time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
			},
			{
				Path:     "b.md",
				Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
			},
			{
				Path:     "dir1/b.md",
				Modified: time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
			},
			{
				Path:     "dir1/dir1/a.md",
				Modified: time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
			},
			{
				Path:     "dir2/a.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
			},
			{
				Path:     "dir1 a space/a.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
			},
			{
				Path:     "Dir3/a.md",
				Modified: time.Date(2019, 11, 12, 20, 34, 6, 0, time.UTC),
			},
		} {
			_, err := dao.Add(note)
			assert.Nil(t, err)
		}

		// We check that the metadata are sorted by the path but not
		// lexicographically. Instead it needs to be sorted on each path
		// component, like filepath.Walk would.
		expected := []paths.Metadata{
			{
				Path:     "Dir3/a.md",
				Modified: time.Date(2019, 11, 12, 20, 34, 6, 0, time.UTC),
			},
			{
				Path:     "a.md",
				Modified: time.Date(2020, 1, 20, 8, 52, 42, 0, time.UTC),
			},
			{
				Path:     "b.md",
				Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
			},
			{
				Path:     "dir1/a.md",
				Modified: time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
			},
			{
				Path:     "dir1/b.md",
				Modified: time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
			},
			{
				Path:     "dir1/dir1/a.md",
				Modified: time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
			},
			{
				Path:     "dir1 a space/a.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
			},
			{
				Path:     "dir2/a.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
			},
		}

		c, err := dao.Indexed()
		assert.Nil(t, err)

		actual := make([]paths.Metadata, 0)
		for a := range c {
			actual = append(actual, a)
		}
		assert.Equal(t, actual, expected)
	})
}

func TestNoteDAOAdd(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Add(core.Note{
			Path:       "log/added.md",
			Title:      "Added note",
			Lead:       "Note",
			Body:       "Note body",
			RawContent: "# Added note\nNote body",
			WordCount:  2,
			Metadata:   map[string]interface{}{"key": "value"},
			Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:   time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
			Checksum:   "check",
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `path = "log/added.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Path:       "log/added.md",
			Title:      "Added note",
			Lead:       "Note",
			Body:       "Note body",
			RawContent: "# Added note\nNote body",
			WordCount:  2,
			Checksum:   "check",
			Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:   time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
			Metadata:   `{"key":"value"}`,
		})
	})
}

// Check that we can't add a duplicate note with an existing path.
func TestNoteDAOAddExistingNote(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Add(core.Note{Path: "ref/test/a.md"})
		assert.Err(t, err, "UNIQUE constraint failed: notes.path")
	})
}

func TestNoteDAOUpdate(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		id, err := dao.Update(core.Note{
			Path:       "ref/test/a.md",
			Title:      "Updated note",
			Lead:       "Updated lead",
			Body:       "Updated body",
			RawContent: "Updated raw content",
			Checksum:   "updated checksum",
			Metadata:   map[string]interface{}{"updated-key": "updated-value"},
			WordCount:  42,
			Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:   time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
		})
		assert.Nil(t, err)
		assert.Equal(t, id, core.NoteID(6))

		row, err := queryNoteRow(tx, `path = "ref/test/a.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Path:       "ref/test/a.md",
			Title:      "Updated note",
			Lead:       "Updated lead",
			Body:       "Updated body",
			RawContent: "Updated raw content",
			Checksum:   "updated checksum",
			WordCount:  42,
			Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:   time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
			Metadata:   `{"updated-key":"updated-value"}`,
		})
	})
}

func TestNoteDAOUpdateUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Update(core.Note{
			Path: "unknown/unknown.md",
		})
		assert.Err(t, err, "note not found in the index")
	})
}

func TestNoteDAORemove(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := queryNoteRow(tx, `path = "ref/test/a.md"`)
		assert.Nil(t, err)

		err = dao.Remove("ref/test/a.md")
		assert.Nil(t, err)

		_, err = queryNoteRow(tx, `path = "ref/test/a.md"`)
		assert.Equal(t, err, sql.ErrNoRows)
	})
}

func TestNoteDAORemoveUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Remove("unknown/unknown.md")
		assert.Err(t, err, "note not found in the index")
	})
}

// Also remove the outbound links, and set the target_id of inbound links to NULL.
func TestNoteDAORemoveCascadeLinks(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		links := queryLinkRows(t, tx, `source_id = 1`)
		assert.Equal(t, len(links) > 0, true)

		links = queryLinkRows(t, tx, `id = 4`)
		assert.Equal(t, *links[0].TargetId, core.NoteID(1))

		err := dao.Remove("log/2021-01-03.md")
		assert.Nil(t, err)

		links = queryLinkRows(t, tx, `source_id = 1`)
		assert.Equal(t, len(links), 0)

		links = queryLinkRows(t, tx, `id = 4`)
		assert.Nil(t, links[0].TargetId)
	})
}

func TestNoteDAOFindIdsByHref(t *testing.T) {
	test := func(href string, allowPartialHref bool, expected []core.NoteID) {
		testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
			actual, err := dao.FindIdsByHref(href, allowPartialHref)
			assert.Nil(t, err)
			assert.Equal(t, actual, expected)
		})
	}

	test("test", false, []core.NoteID{})
	test("test", true, []core.NoteID{6, 5, 8})

	// Filename takes precedence over the rest of the path.
	// See https://github.com/zk-org/zk/issues/111
	test("ref", true, []core.NoteID{8})

}

func TestNoteDAOFindIdsByHrefPrefixBug(t *testing.T) {
	testNoteDAOWithFixtures(t, "", func(tx Transaction, dao *NoteDAO) {
		shorterNote := core.Note{
			Path:     "2024-08-27.md",
			Title:    "Shorter note",
			Body:     "This is the shorter note",
			Modified: time.Date(2024, 8, 27, 10, 0, 0, 0, time.UTC),
		}
		longerNote := core.Note{
			Path:     "2024-08-27_ajct.md",
			Title:    "Longer note with suffix",
			Body:     "This is the longer note with suffix",
			Modified: time.Date(2024, 8, 27, 11, 0, 0, 0, time.UTC),
		}

		shorterId, err := dao.Add(shorterNote)
		assert.Nil(t, err)
		longerId, err := dao.Add(longerNote)
		assert.Nil(t, err)

		// Test partial matching like wiki links would use
		ids, err := dao.FindIdsByHref("2024-08-27", true)
		assert.Nil(t, err)
		if len(ids) == 0 {
			t.Fatal("Should find at least one match")
		}

		t.Logf("Partial: Found %d matches for '2024-08-27': %v", len(ids), ids)
		t.Logf("Shorter note ID: %d, Longer note ID: %d", shorterId, longerId)
		t.Logf("Expected first ID: %d (2024-08-27.md), Actual first ID: %d", shorterId, ids[0])

		if ids[0] != shorterId {
			t.Errorf("Expected exact match '2024-08-27.md' (ID %d) but got ID %d. This demonstrates the prefix matching bug.", shorterId, ids[0])
		}

		// Also test exact matching
		exactIds, err := dao.FindIdsByHref("2024-08-27.md", false)
		assert.Nil(t, err)
		if len(exactIds) == 0 {
			t.Fatal("Should find at least one exact match")
		}

		t.Logf("Exact: Found %d matches for '2024-08-27.md': %v", len(exactIds), exactIds)

		if exactIds[0] != shorterId {
			t.Errorf("Exact matching failed: Expected '2024-08-27.md' (ID %d) but got ID %d. This affects LSP markdown links.", shorterId, exactIds[0])
		}
	})
}

func TestNoteDAOFindIncludingHrefs(t *testing.T) {
	test := func(href string, allowPartialHref bool, expected []string) {
		testNoteDAOFindPaths(t,
			core.NoteFindOpts{
				IncludeHrefs:      []string{href},
				AllowPartialHrefs: allowPartialHref,
			},
			expected,
		)
	}

	test("test", false, []string{})
	test("test", true, []string{"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md"})

	// Filename takes precedence over the rest of the path.
	// See https://github.com/zk-org/zk/issues/111
	test("ref", true, []string{"ref/test/ref.md"})
}

func TestNoteDAOFindExcludingHrefs(t *testing.T) {
	test := func(href string, allowPartialHref bool, expected []string) {
		testNoteDAOFindPaths(t,
			core.NoteFindOpts{
				ExcludeHrefs:      []string{href},
				AllowPartialHrefs: allowPartialHref,
			},
			expected,
		)
	}

	test("test", false, []string{"ref/test/ref.md", "ref/test/b.md",
		"f39c8.md", "ref/test/a.md", "log/2021-01-03.md", "log/2021-02-04.md",
		"index.md", "log/2021-01-04.md"})
	test("test", true, []string{"f39c8.md", "log/2021-01-03.md",
		"log/2021-02-04.md", "index.md", "log/2021-01-04.md"})

	// Filename takes precedence over the rest of the path.
	// See https://github.com/zk-org/zk/issues/111
	test("ref", true, []string{"ref/test/b.md", "f39c8.md", "ref/test/a.md",
		"log/2021-01-03.md", "log/2021-02-04.md", "index.md", "log/2021-01-04.md"})
}

func TestNoteDAOFindMinimalAll(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		notes, err := dao.FindMinimal(core.NoteFindOpts{})
		assert.Nil(t, err)

		assert.Equal(t, notes, []core.MinimalNote{
			{ID: 8, Path: "ref/test/ref.md", Title: "", Metadata: map[string]interface{}{}},
			{ID: 5, Path: "ref/test/b.md", Title: "A nested note", Metadata: map[string]interface{}{}},
			{ID: 4, Path: "f39c8.md", Title: "An interesting note", Metadata: map[string]interface{}{}},
			{ID: 6, Path: "ref/test/a.md", Title: "Another nested note", Metadata: map[string]interface{}{
				"alias": "a.md",
			}},
			{ID: 1, Path: "log/2021-01-03.md", Title: "Daily note", Metadata: map[string]interface{}{
				"author": "Dom",
			}},
			{ID: 7, Path: "log/2021-02-04.md", Title: "February 4, 2021", Metadata: map[string]interface{}{}},
			{ID: 3, Path: "index.md", Title: "Index", Metadata: map[string]interface{}{
				"aliases": []interface{}{"First page"},
			}},
			{ID: 2, Path: "log/2021-01-04.md", Title: "January 4, 2021", Metadata: map[string]interface{}{}},
		})
	})
}

func TestNoteDAOFindMinimalWithFilter(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		notes, err := dao.FindMinimal(core.NoteFindOpts{
			Match:         []string{"daily | index"},
			MatchStrategy: core.MatchStrategyFts,
			Sorters:       []core.NoteSorter{{Field: core.NoteSortWordCount, Ascending: true}},
			Limit:         3,
		})
		assert.Nil(t, err)

		assert.Equal(t, notes, []core.MinimalNote{
			{ID: 1, Path: "log/2021-01-03.md", Title: "Daily note", Metadata: map[string]interface{}{
				"author": "Dom",
			}},
			{ID: 3, Path: "index.md", Title: "Index", Metadata: map[string]interface{}{
				"aliases": []interface{}{"First page"},
			}},
			{ID: 7, Path: "log/2021-02-04.md", Title: "February 4, 2021", Metadata: map[string]interface{}{}},
		})
	})
}

func TestNoteDAOFindAll(t *testing.T) {
	testNoteDAOFindPaths(t, core.NoteFindOpts{}, []string{
		"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-01-03.md",
		"log/2021-02-04.md", "index.md", "log/2021-01-04.md",
	})
}

func TestNoteDAOFindLimit(t *testing.T) {
	testNoteDAOFindPaths(t, core.NoteFindOpts{Limit: 3}, []string{
		"ref/test/ref.md",
		"ref/test/b.md",
		"f39c8.md",
	})
}

func TestNoteDAOFindTag(t *testing.T) {
	test := func(tags []string, expectedPaths []string) {
		testNoteDAOFindPaths(t, core.NoteFindOpts{Tags: tags}, expectedPaths)
	}

	test([]string{"fiction"}, []string{"log/2021-01-03.md"})
	test([]string{" adventure "}, []string{"ref/test/b.md", "log/2021-01-03.md"})
	test([]string{"fiction", "adventure"}, []string{"log/2021-01-03.md"})
	test([]string{"fiction|fantasy"}, []string{"f39c8.md", "log/2021-01-03.md"})
	test([]string{"fiction  |   fantasy"}, []string{"f39c8.md", "log/2021-01-03.md"})
	test([]string{"fiction  OR  fantasy"}, []string{"f39c8.md", "log/2021-01-03.md"})
	test([]string{"fiction | adventure | fantasy"}, []string{"ref/test/b.md", "f39c8.md", "log/2021-01-03.md"})
	test([]string{"fiction | history", "adventure"}, []string{"ref/test/b.md", "log/2021-01-03.md"})
	test([]string{"fiction", "unknown"}, []string{})
	test([]string{"-fiction"}, []string{"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-02-04.md", "index.md", "log/2021-01-04.md"})
	test([]string{"NOT   fiction"}, []string{"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-02-04.md", "index.md", "log/2021-01-04.md"})
	test([]string{"NOTfiction"}, []string{"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-02-04.md", "index.md", "log/2021-01-04.md"})
}

func TestNoteDAOFindMatch(t *testing.T) {
	testNoteDAOFind(t,
		core.NoteFindOpts{
			Match:         []string{"daily | index"},
			MatchStrategy: core.MatchStrategyFts,
		},
		[]core.ContextualNote{
			{
				Note: core.Note{
					ID:         3,
					Path:       "index.md",
					Title:      "Index",
					Lead:       "Index of the Zettelkasten",
					Body:       "Index of the Zettelkasten",
					RawContent: "# Index\nIndex of the Zettelkasten",
					WordCount:  4,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata: map[string]interface{}{
						"aliases": []interface{}{"First page"},
					},
					Created:  time.Date(2019, 12, 4, 11, 59, 11, 0, time.UTC),
					Modified: time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
					Checksum: "iaefhv",
				},
				Snippets: []string{"<zk:match>Index</zk:match> of the Zettelkasten"},
			},
			{
				Note: core.Note{
					ID:         1,
					Path:       "log/2021-01-03.md",
					Title:      "Daily note",
					Lead:       "A daily note",
					Body:       "A daily note\n\nWith lot of content",
					RawContent: "# Daily note\nA note\n\nWith lot of content",
					WordCount:  3,
					Links:      []core.Link{},
					Tags:       []string{"fiction", "adventure"},
					Metadata: map[string]interface{}{
						"author": "Dom",
					},
					Created:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Checksum: "qwfpgj",
				},
				Snippets: []string{"A <zk:match>daily</zk:match> note\n\nWith lot of content"},
			},
			{
				Note: core.Note{
					ID:         7,
					Path:       "log/2021-02-04.md",
					Title:      "February 4, 2021",
					Lead:       "A third daily note",
					Body:       "A third daily note",
					RawContent: "# A third daily note",
					WordCount:  4,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata:   map[string]interface{}{},
					Created:    time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:   time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
					Checksum:   "earkte",
				},
				Snippets: []string{"A third <zk:match>daily</zk:match> note"},
			},
			{
				Note: core.Note{
					ID:         2,
					Path:       "log/2021-01-04.md",
					Title:      "January 4, 2021",
					Lead:       "A second daily note",
					Body:       "A second daily note",
					RawContent: "# A second daily note",
					WordCount:  4,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata:   map[string]interface{}{},
					Created:    time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Checksum:   "arstde",
				},
				Snippets: []string{"A second <zk:match>daily</zk:match> note"},
			},
		},
	)
}

func TestNoteDAOFindMatchWithMultiMatch(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			Match:         []string{"daily | index", "second"},
			MatchStrategy: core.MatchStrategyFts,
			Sorters: []core.NoteSorter{
				{Field: core.NoteSortPath, Ascending: false},
			},
		},
		[]string{
			"log/2021-01-04.md",
		},
	)
}

func TestNoteDAOFindMatchWithSort(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			Match:         []string{"daily | index"},
			MatchStrategy: core.MatchStrategyFts,
			Sorters: []core.NoteSorter{
				{Field: core.NoteSortPath, Ascending: false},
			},
		},
		[]string{
			"log/2021-02-04.md",
			"log/2021-01-04.md",
			"log/2021-01-03.md",
			"index.md",
		},
	)
}

func TestNoteDAOFindExactMatch(t *testing.T) {
	test := func(match string, expected []string) {
		testNoteDAOFindPaths(t,
			core.NoteFindOpts{
				Match:         []string{match},
				MatchStrategy: core.MatchStrategyExact,
			},
			expected,
		)
	}

	// Case insensitive
	test("dailY NOTe", []string{"log/2021-01-03.md", "log/2021-02-04.md", "log/2021-01-04.md"})
	// Special characters
	test(`[exact% ch\ar_acters]`, []string{"ref/test/a.md"})
}

func TestNoteDAOFindMentionRequiresFtsMatchStrategy(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Find(core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyExact,
			Mention:       []string{"mention"},
		})
		assert.Err(t, err, "--mention can only be used with --match-strategy=fts")
	})
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Find(core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyRe,
			Mention:       []string{"mention"},
		})
		assert.Err(t, err, "--mention can only be used with --match-strategy=fts")
	})
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Find(core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			Mention:       []string{"mention"},
		})
		assert.Err(t, err, "could not find notes at: mention")
	})
}

func TestNoteDAOFindInPathAbsoluteFile(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			IncludeHrefs: []string{"log/2021-01-03.md"},
		},
		[]string{"log/2021-01-03.md"},
	)
}

// You can look for files with only their prefix.
func TestNoteDAOFindInPathWithFilePrefix(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			IncludeHrefs: []string{"log/2021-01"},
		},
		[]string{"log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

// For directory, only complete names work, no prefixes.
func TestNoteDAOFindInPathRequiresCompleteDirName(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			IncludeHrefs:      []string{"lo"},
			AllowPartialHrefs: false,
		},
		[]string{},
	)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			IncludeHrefs:      []string{"log"},
			AllowPartialHrefs: false,
		},
		[]string{"log/2021-01-03.md", "log/2021-02-04.md", "log/2021-01-04.md"},
	)
}

// You can look for multiple paths, in which case notes can be in any of them.
func TestNoteDAOFindInMultiplePaths(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			IncludeHrefs: []string{"ref", "index.md"},
		},
		[]string{"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindExcludingPath(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			ExcludeHrefs: []string{"log"},
		},
		[]string{"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindExcludingMultiplePaths(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			ExcludeHrefs: []string{"ref", "log/2021-01"},
		},
		[]string{"f39c8.md", "log/2021-02-04.md", "index.md"},
	)
}

func TestNoteDAOFindMentions(t *testing.T) {
	testNoteDAOFind(t,
		core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			Mention:       []string{"log/2021-01-03.md", "index.md"},
		},
		[]core.ContextualNote{
			{
				Note: core.Note{
					ID:         5,
					Path:       "ref/test/b.md",
					Title:      "A nested note",
					Lead:       "This one is in a sub sub directory",
					Body:       "This one is in a sub sub directory, not the first page",
					RawContent: "# A nested note\nThis one is in a sub sub directory",
					WordCount:  8,
					Links:      []core.Link{},
					Tags:       []string{"adventure", "history", "science"},
					Metadata:   map[string]interface{}{},
					Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
					Modified:   time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
					Checksum:   "yvwbae",
				},
				Snippets: []string{"This one is in a sub sub directory, not the <zk:match>first page</zk:match>"},
			},
			{
				Note: core.Note{
					ID:         7,
					Path:       "log/2021-02-04.md",
					Title:      "February 4, 2021",
					Lead:       "A third daily note",
					Body:       "A third daily note",
					RawContent: "# A third daily note",
					WordCount:  4,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata:   map[string]interface{}{},
					Created:    time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:   time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
					Checksum:   "earkte",
				},
				Snippets: []string{"A third <zk:match>daily note</zk:match>"},
			},
			{
				Note: core.Note{
					ID:         2,
					Path:       "log/2021-01-04.md",
					Title:      "January 4, 2021",
					Lead:       "A second daily note",
					Body:       "A second daily note",
					RawContent: "# A second daily note",
					WordCount:  4,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata:   map[string]interface{}{},
					Created:    time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Checksum:   "arstde",
				},
				Snippets: []string{"A second <zk:match>daily note</zk:match>"},
			},
		},
	)
}

// Common use case: `--mention x --no-link-to x`
func TestNoteDAOFindUnlinkedMentions(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			Mention:       []string{"log/2021-01-03.md", "index.md"},
			LinkTo: &core.LinkFilter{
				Hrefs:  []string{"log/2021-01-03.md", "index.md"},
				Negate: true,
			},
		},
		[]string{"ref/test/b.md", "log/2021-02-04.md"},
	)
}

func TestNoteDAOFindMentionUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		opts := core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			Mention:       []string{"will-not-be-found"},
		}
		_, err := dao.Find(opts)
		assert.Err(t, err, "could not find notes at: will-not-be-found")
	})
}

func TestNoteDAOFindMentionedBy(t *testing.T) {
	testNoteDAOFind(t,
		core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			MentionedBy:   []string{"ref/test/b.md", "log/2021-01-04.md"},
		},
		[]core.ContextualNote{
			{
				Note: core.Note{
					ID:         1,
					Path:       "log/2021-01-03.md",
					Title:      "Daily note",
					Lead:       "A daily note",
					Body:       "A daily note\n\nWith lot of content",
					RawContent: "# Daily note\nA note\n\nWith lot of content",
					WordCount:  3,
					Links:      []core.Link{},
					Tags:       []string{"fiction", "adventure"},
					Metadata: map[string]interface{}{
						"author": "Dom",
					},
					Created:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Checksum: "qwfpgj",
				},
				Snippets: []string{"A second <zk:match>daily note</zk:match>"},
			},
			{
				Note: core.Note{
					ID:         3,
					Path:       "index.md",
					Title:      "Index",
					Lead:       "Index of the Zettelkasten",
					Body:       "Index of the Zettelkasten",
					RawContent: "# Index\nIndex of the Zettelkasten",
					WordCount:  4,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata: map[string]interface{}{
						"aliases": []interface{}{
							"First page",
						},
					},
					Created:  time.Date(2019, 12, 4, 11, 59, 11, 0, time.UTC),
					Modified: time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
					Checksum: "iaefhv",
				},
				Snippets: []string{"This one is in a sub sub directory, not the <zk:match>first page</zk:match>"},
			},
		},
	)
}

// Common use case: `--mentioned-by x --no-linked-by x`
func TestNoteDAOFindUnlinkedMentionedBy(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			MentionedBy:   []string{"ref/test/b.md", "log/2021-01-04.md"},
			LinkedBy: &core.LinkFilter{
				Hrefs:  []string{"ref/test/b.md", "log/2021-01-04.md"},
				Negate: true,
			},
		},
		[]string{"log/2021-01-03.md"},
	)
}

func TestNoteDAOFindMentionedByUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		opts := core.NoteFindOpts{
			MatchStrategy: core.MatchStrategyFts,
			MentionedBy:   []string{"will-not-be-found"},
		}
		_, err := dao.Find(opts)
		assert.Err(t, err, "could not find notes at: will-not-be-found")
	})
}

func TestNoteDAOFindLinkedBy(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkedBy: &core.LinkFilter{
				Hrefs:     []string{"f39c8.md", "log/2021-01-03"},
				Negate:    false,
				Recursive: false,
			},
		},
		[]string{"ref/test/a.md", "log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindLinkedByRecursive(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkedBy: &core.LinkFilter{
				Hrefs:     []string{"log/2021-01-04.md"},
				Negate:    false,
				Recursive: true,
			},
		},
		[]string{"index.md", "f39c8.md", "ref/test/a.md", "log/2021-01-03.md"},
	)
}

func TestNoteDAOFindLinkedByRecursiveWithMaxDistance(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkedBy: &core.LinkFilter{
				Hrefs:       []string{"log/2021-01-04.md"},
				Negate:      false,
				Recursive:   true,
				MaxDistance: 2,
			},
		},
		[]string{"index.md", "f39c8.md"},
	)
}

func TestNoteDAOFindLinkedByWithSnippets(t *testing.T) {
	testNoteDAOFind(t,
		core.NoteFindOpts{
			LinkedBy: &core.LinkFilter{Hrefs: []string{"f39c8.md"}},
		},
		[]core.ContextualNote{
			{
				Note: core.Note{
					ID:         6,
					Path:       "ref/test/a.md",
					Title:      "Another nested note",
					Lead:       "It shall appear before b.md",
					Body:       "It shall appear before b.md",
					RawContent: "#Another nested note\nIt shall appear before b.md\nMatch [exact% ch\\ar_acters]",
					WordCount:  5,
					Links:      []core.Link{},
					Tags:       []string{},
					Metadata: map[string]interface{}{
						"alias": "a.md",
					},
					Created:  time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
					Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
					Checksum: "iecywst",
				},
				Snippets: []string{
					"[[<zk:match>Link from 4 to 6</zk:match>]]",
					"[[<zk:match>Duplicated link</zk:match>]]",
				},
			},
			{
				Note: core.Note{
					ID:         1,
					Path:       "log/2021-01-03.md",
					Title:      "Daily note",
					Lead:       "A daily note",
					Body:       "A daily note\n\nWith lot of content",
					RawContent: "# Daily note\nA note\n\nWith lot of content",
					WordCount:  3,
					Links:      []core.Link{},
					Tags:       []string{"fiction", "adventure"},
					Metadata: map[string]interface{}{
						"author": "Dom",
					},
					Created:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Checksum: "qwfpgj",
				},
				Snippets: []string{
					"[[<zk:match>Another link</zk:match>]]",
				},
			},
		},
	)
}

func TestNoteDAOFindLinkedByUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		opts := core.NoteFindOpts{
			LinkedBy: &core.LinkFilter{
				Hrefs: []string{"will-not-be-found"},
			},
		}
		_, err := dao.Find(opts)
		assert.Err(t, err, "could not find notes at: will-not-be-found")
	})
}

func TestNoteDAOFindNotLinkedBy(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkedBy: &core.LinkFilter{
				Hrefs:     []string{"f39c8.md", "log/2021-01-03"},
				Negate:    true,
				Recursive: false,
			},
		},
		[]string{"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "log/2021-02-04.md", "index.md"},
	)
}

func TestNoteDAOFindLinkTo(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkTo: &core.LinkFilter{
				Hrefs:     []string{"log/2021-01-04", "ref/test/a.md"},
				Negate:    false,
				Recursive: false,
			},
		},
		[]string{"f39c8.md", "log/2021-01-03.md"},
	)
}

func TestNoteDAOFindLinkToRecursive(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkTo: &core.LinkFilter{
				Hrefs:     []string{"log/2021-01-04.md"},
				Negate:    false,
				Recursive: true,
			},
		},
		[]string{"log/2021-01-03.md", "f39c8.md", "index.md"},
	)
}

func TestNoteDAOFindLinkToRecursiveWithMaxDistance(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkTo: &core.LinkFilter{
				Hrefs:       []string{"log/2021-01-04.md"},
				Negate:      false,
				Recursive:   true,
				MaxDistance: 2,
			},
		},
		[]string{"log/2021-01-03.md", "f39c8.md"},
	)
}

func TestNoteDAOFindNotLinkTo(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			LinkTo: &core.LinkFilter{Hrefs: []string{"log/2021-01-04", "ref/test/a.md"}, Negate: true},
		},
		[]string{"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md", "log/2021-02-04.md", "index.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindLinkToUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		opts := core.NoteFindOpts{
			LinkTo: &core.LinkFilter{
				Hrefs: []string{"will-not-be-found"},
			},
		}
		_, err := dao.Find(opts)
		assert.Err(t, err, "could not find notes at: will-not-be-found")
	})
}

func TestNoteDAOFindRelated(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			Related: []string{"log/2021-02-04"},
		},
		[]string{},
	)

	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			Related: []string{"log/2021-01-03.md"},
		},
		[]string{"index.md"},
	)
}

func TestNoteDAOFindOrphan(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{Orphan: true},
		[]string{"ref/test/ref.md", "ref/test/b.md", "log/2021-02-04.md"},
	)
}

func TestNoteDAOFindMissingBacklink(t *testing.T) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{MissingBacklink: true},
		[]string{"f39c8.md", "ref/test/a.md", "log/2021-01-03.md", "index.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindCreatedOn(t *testing.T) {
	start := time.Date(2020, 11, 22, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, 11, 23, 0, 0, 0, 0, time.UTC)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			CreatedStart: &start,
			CreatedEnd:   &end,
		},
		[]string{"log/2021-01-03.md"},
	)
}

func TestNoteDAOFindCreatedBefore(t *testing.T) {
	end := time.Date(2019, 12, 04, 11, 59, 11, 0, time.UTC)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			CreatedEnd: &end,
		},
		[]string{"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md"},
	)
}

func TestNoteDAOFindCreatedAfter(t *testing.T) {
	start := time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			CreatedStart: &start,
		},
		[]string{"log/2021-01-03.md", "log/2021-02-04.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindModifiedOn(t *testing.T) {
	start := time.Date(2020, 01, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, 01, 21, 0, 0, 0, 0, time.UTC)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			ModifiedStart: &start,
			ModifiedEnd:   &end,
		},
		[]string{"f39c8.md"},
	)
}

func TestNoteDAOFindModifiedBefore(t *testing.T) {
	end := time.Date(2020, 01, 20, 8, 52, 42, 0, time.UTC)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			ModifiedEnd: &end,
		},
		[]string{"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindModifiedAfter(t *testing.T) {
	start := time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC)
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			ModifiedStart: &start,
		},
		[]string{"log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindSortCreated(t *testing.T) {
	testNoteDAOFindSort(t, core.NoteSortCreated, true, []string{
		"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md", "index.md", "f39c8.md",
		"log/2021-01-03.md", "log/2021-02-04.md", "log/2021-01-04.md",
	})
	testNoteDAOFindSort(t, core.NoteSortCreated, false, []string{
		"log/2021-02-04.md", "log/2021-01-04.md", "log/2021-01-03.md",
		"f39c8.md", "index.md", "ref/test/ref.md", "ref/test/b.md", "ref/test/a.md",
	})
}

func TestNoteDAOFindSortModified(t *testing.T) {
	testNoteDAOFindSort(t, core.NoteSortModified, true, []string{
		"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md", "index.md", "f39c8.md",
		"log/2021-02-04.md", "log/2021-01-03.md", "log/2021-01-04.md",
	})
	testNoteDAOFindSort(t, core.NoteSortModified, false, []string{
		"log/2021-01-04.md", "log/2021-01-03.md", "log/2021-02-04.md",
		"f39c8.md", "index.md", "ref/test/ref.md", "ref/test/b.md", "ref/test/a.md",
	})
}

func TestNoteDAOFindSortPath(t *testing.T) {
	testNoteDAOFindSort(t, core.NoteSortPath, true, []string{
		"f39c8.md", "index.md", "log/2021-01-03.md", "log/2021-01-04.md",
		"log/2021-02-04.md", "ref/test/a.md", "ref/test/b.md", "ref/test/ref.md",
	})
	testNoteDAOFindSort(t, core.NoteSortPath, false, []string{
		"ref/test/ref.md", "ref/test/b.md", "ref/test/a.md", "log/2021-02-04.md",
		"log/2021-01-04.md", "log/2021-01-03.md", "index.md", "f39c8.md",
	})
}

func TestNoteDAOFindSortTitle(t *testing.T) {
	testNoteDAOFindSort(t, core.NoteSortTitle, true, []string{
		"ref/test/ref.md", "ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-01-03.md",
		"log/2021-02-04.md", "index.md", "log/2021-01-04.md",
	})
	testNoteDAOFindSort(t, core.NoteSortTitle, false, []string{
		"log/2021-01-04.md", "index.md", "log/2021-02-04.md",
		"log/2021-01-03.md", "ref/test/a.md", "f39c8.md", "ref/test/b.md", "ref/test/ref.md",
	})
}

func TestNoteDAOFindSortWordCount(t *testing.T) {
	testNoteDAOFindSort(t, core.NoteSortWordCount, true, []string{
		"log/2021-01-03.md", "log/2021-02-04.md", "index.md",
		"log/2021-01-04.md", "ref/test/ref.md", "f39c8.md", "ref/test/a.md", "ref/test/b.md",
	})
	testNoteDAOFindSort(t, core.NoteSortWordCount, false, []string{
		"ref/test/b.md", "ref/test/ref.md", "f39c8.md", "ref/test/a.md", "log/2021-02-04.md",
		"index.md", "log/2021-01-04.md", "log/2021-01-03.md",
	})
}

func testNoteDAOFindSort(t *testing.T, field core.NoteSortField, ascending bool, expected []string) {
	testNoteDAOFindPaths(t,
		core.NoteFindOpts{
			Sorters: []core.NoteSorter{{Field: field, Ascending: ascending}},
		},
		expected,
	)
}

func testNoteDAOFindPaths(t *testing.T, opts core.NoteFindOpts, expected []string) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		matches, err := dao.Find(opts)
		assert.Nil(t, err)

		actual := make([]string, 0)
		for _, m := range matches {
			actual = append(actual, m.Path)
		}
		assert.Equal(t, actual, expected)
	})
}

func testNoteDAOFind(t *testing.T, opts core.NoteFindOpts, expected []core.ContextualNote) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		actual, err := dao.Find(opts)
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	})
}

func testNoteDAO(t *testing.T, callback func(tx Transaction, dao *NoteDAO)) {
	testTransaction(t, func(tx Transaction) {
		callback(tx, NewNoteDAO(tx, &util.NullLogger))
	})
}

func testNoteDAOWithFixtures(t *testing.T, fixtures string, callback func(tx Transaction, dao *NoteDAO)) {
	testTransactionWithFixtures(t, opt.NewNotEmptyString(fixtures), func(tx Transaction) {
		callback(tx, NewNoteDAO(tx, &util.NullLogger))
	})
}

type noteRow struct {
	Path, Title, Lead, Body, RawContent, Checksum, Metadata string
	WordCount                                               int
	Created, Modified                                       time.Time
}

func queryNoteRow(tx Transaction, where string) (noteRow, error) {
	var row noteRow
	err := tx.QueryRow(fmt.Sprintf(`
		SELECT path, title, lead, body, raw_content, word_count, checksum, created, modified, metadata
		  FROM notes
		 WHERE %v
	`, where)).Scan(&row.Path, &row.Title, &row.Lead, &row.Body, &row.RawContent, &row.WordCount, &row.Checksum, &row.Created, &row.Modified, &row.Metadata)
	return row, err
}

func idPointer(i int64) *core.NoteID {
	id := core.NoteID(i)
	return &id
}
