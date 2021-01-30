package sqlite

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestNoteDAOIndexed(t *testing.T) {
	testNoteDAOWithoutFixtures(t, func(tx Transaction, dao *NoteDAO) {
		for _, note := range []note.Metadata{
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
		_, err := dao.Add(note.Metadata{
			Path:       "log/added.md",
			Title:      "Added note",
			Lead:       "Note",
			Body:       "Note body",
			RawContent: "# Added note\nNote body",
			WordCount:  2,
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
		})
	})
}

func TestNoteDAOAddWithLinks(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		id, err := dao.Add(note.Metadata{
			Path: "log/added.md",
			Links: []note.Link{
				{
					Title: "Same dir",
					Href:  "log/2021-01-04",
					Rels:  []string{"rel-1", "rel-2"},
				},
				{
					Title:   "Relative",
					Href:    "f39c8",
					Snippet: "[Relative](f39c8) link",
				},
				{
					Title: "Second is added",
					Href:  "f39c8",
					Rels:  []string{"second"},
				},
				{
					Title: "Unknown",
					Href:  "unknown",
				},
				{
					Title:    "URL",
					Href:     "http://example.com",
					External: true,
					Snippet:  "External [URL](http://example.com)",
				},
			},
		})
		assert.Nil(t, err)

		rows := queryLinkRows(t, tx, fmt.Sprintf("source_id = %d", id))
		assert.Equal(t, rows, []linkRow{
			{
				SourceId: id,
				TargetId: intPointer(2),
				Title:    "Same dir",
				Href:     "log/2021-01-04",
				Rels:     "\x01rel-1\x01rel-2\x01",
			},
			{
				SourceId: id,
				TargetId: intPointer(4),
				Title:    "Relative",
				Href:     "f39c8",
				Rels:     "",
				Snippet:  "[Relative](f39c8) link",
			},
			{
				SourceId: id,
				TargetId: intPointer(4),
				Title:    "Second is added",
				Href:     "f39c8",
				Rels:     "\x01second\x01",
			},
			{
				SourceId: id,
				TargetId: nil,
				Title:    "Unknown",
				Href:     "unknown",
				Rels:     "",
			},
			{
				SourceId: id,
				TargetId: nil,
				Title:    "URL",
				Href:     "http://example.com",
				External: true,
				Rels:     "",
				Snippet:  "External [URL](http://example.com)",
			},
		})
	})
}

func TestNoteDAOAddFillsLinksMissingTargetId(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		id, err := dao.Add(note.Metadata{
			Path: "missing_target.md",
		})
		assert.Nil(t, err)

		rows := queryLinkRows(t, tx, fmt.Sprintf("target_id = %d", id))
		assert.Equal(t, rows, []linkRow{
			{
				SourceId: 3,
				TargetId: &id,
				Title:    "Missing target",
				Href:     "missing",
				Snippet:  "There's a Missing target",
			},
		})
	})
}

// Check that we can't add a duplicate note with an existing path.
func TestNoteDAOAddExistingNote(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		_, err := dao.Add(note.Metadata{Path: "ref/test/a.md"})
		assert.Err(t, err, "ref/test/a.md: can't add note to the index: UNIQUE constraint failed: notes.path")
	})
}

func TestNoteDAOUpdate(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Update(note.Metadata{
			Path:       "ref/test/a.md",
			Title:      "Updated note",
			Lead:       "Updated lead",
			Body:       "Updated body",
			RawContent: "Updated raw content",
			Checksum:   "updated checksum",
			WordCount:  42,
			Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
			Modified:   time.Date(2020, 11, 22, 16, 49, 47, 0, time.UTC),
		})
		assert.Nil(t, err)

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
		})
	})
}

func TestNoteDAOUpdateUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Update(note.Metadata{
			Path: "unknown/unknown.md",
		})
		assert.Err(t, err, "unknown/unknown.md: failed to update note index: note not found in the index")
	})
}

func TestNoteDAOUpdateWithLinks(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		links := queryLinkRows(t, tx, "source_id = 1")
		assert.Equal(t, links, []linkRow{
			{
				SourceId: 1,
				TargetId: intPointer(2),
				Title:    "An internal link",
				Href:     "log/2021-01-04.md",
				Snippet:  "[[An internal link]]",
			},
			{
				SourceId: 1,
				TargetId: nil,
				Title:    "An external link",
				Href:     "https://domain.com",
				External: true,
				Snippet:  "[[An external link]]",
			},
		})

		err := dao.Update(note.Metadata{
			Path: "log/2021-01-03.md",
			Links: []note.Link{
				{
					Title:    "A new link",
					Href:     "index",
					External: false,
					Rels:     []string{"rel"},
					Snippet:  "[[A new link]]",
				},
				{
					Title:    "An external link",
					Href:     "https://domain.com",
					External: true,
					Snippet:  "[[An external link]]",
				},
			},
		})
		assert.Nil(t, err)

		links = queryLinkRows(t, tx, "source_id = 1")
		assert.Equal(t, links, []linkRow{
			{
				SourceId: 1,
				TargetId: intPointer(3),
				Title:    "A new link",
				Href:     "index",
				Rels:     "\x01rel\x01",
				Snippet:  "[[A new link]]",
			},
			{
				SourceId: 1,
				TargetId: nil,
				Title:    "An external link",
				Href:     "https://domain.com",
				External: true,
				Snippet:  "[[An external link]]",
			},
		})
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
		assert.Err(t, err, "unknown/unknown.md: failed to remove note index: note not found in the index")
	})
}

// Also remove the outbound links, and set the target_id of inbound links to NULL.
func TestNoteDAORemoveCascadeLinks(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		links := queryLinkRows(t, tx, `source_id = 1`)
		assert.Equal(t, len(links) > 0, true)

		links = queryLinkRows(t, tx, `id = 4`)
		assert.Equal(t, *links[0].TargetId, int64(1))

		err := dao.Remove("log/2021-01-03.md")
		assert.Nil(t, err)

		links = queryLinkRows(t, tx, `source_id = 1`)
		assert.Equal(t, len(links), 0)

		links = queryLinkRows(t, tx, `id = 4`)
		assert.Nil(t, links[0].TargetId)
	})
}

func TestNoteDAOFindAll(t *testing.T) {
	testNoteDAOFindPaths(t, note.FinderOpts{}, []string{
		"ref/test/b.md",
		"f39c8.md",
		"ref/test/a.md",
		"log/2021-02-04.md",
		"index.md",
		"log/2021-01-03.md",
		"log/2021-01-04.md",
	})
}

func TestNoteDAOFindLimit(t *testing.T) {
	testNoteDAOFindPaths(t, note.FinderOpts{Limit: 2}, []string{
		"ref/test/b.md",
		"f39c8.md",
	})
}

func TestNoteDAOFindMatch(t *testing.T) {
	testNoteDAOFind(t,
		note.FinderOpts{
			Filters: []note.Filter{note.MatchFilter("daily | index")},
		},
		[]note.Match{
			{
				Metadata: note.Metadata{
					Path:       "index.md",
					Title:      "Index",
					Lead:       "Index of the Zettelkasten",
					Body:       "Index of the Zettelkasten",
					RawContent: "# Index\nIndex of the Zettelkasten",
					WordCount:  4,
					Created:    time.Date(2019, 12, 4, 11, 59, 11, 0, time.UTC),
					Modified:   time.Date(2019, 12, 4, 12, 17, 21, 0, time.UTC),
					Checksum:   "iaefhv",
				},
				Snippets: []string{"<zk:match>Index</zk:match> of the Zettelkasten"},
			},
			{
				Metadata: note.Metadata{
					Path:       "log/2021-02-04.md",
					Title:      "February 4, 2021",
					Lead:       "A third daily note",
					Body:       "A third daily note",
					RawContent: "# A third daily note",
					WordCount:  4,
					Created:    time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:   time.Date(2020, 11, 10, 8, 20, 18, 0, time.UTC),
					Checksum:   "earkte",
				},
				Snippets: []string{"A third <zk:match>daily</zk:match> note"},
			},
			{
				Metadata: note.Metadata{
					Path:       "log/2021-01-04.md",
					Title:      "January 4, 2021",
					Lead:       "A second daily note",
					Body:       "A second daily note",
					RawContent: "# A second daily note",
					WordCount:  4,
					Created:    time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Modified:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.UTC),
					Checksum:   "arstde",
				},
				Snippets: []string{"A second <zk:match>daily</zk:match> note"},
			},
			{
				Metadata: note.Metadata{
					Path:       "log/2021-01-03.md",
					Title:      "January 3, 2021",
					Lead:       "A daily note",
					Body:       "A daily note\n\nWith lot of content",
					RawContent: "# A daily note\nA daily note\n\nWith lot of content",
					WordCount:  3,
					Created:    time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Modified:   time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Checksum:   "qwfpgj",
				},
				Snippets: []string{"A <zk:match>daily</zk:match> note\n\nWith lot of content"},
			},
		},
	)
}

func TestNoteDAOFindInPath(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.PathFilter([]string{"log/2021-01-*"})},
		},
		[]string{"log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindInMultiplePaths(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.PathFilter([]string{"ref", "index.md"})},
		},
		[]string{"ref/test/b.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindExcludingPath(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.ExcludePathFilter([]string{"log"})},
		},
		[]string{"ref/test/b.md", "f39c8.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindExcludingMultiplePaths(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.ExcludePathFilter([]string{"ref", "log/2021-01-*"})},
		},
		[]string{"f39c8.md", "log/2021-02-04.md", "index.md"},
	)
}

func TestNoteDAOFindLinkedBy(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.LinkedByFilter{Paths: []string{"f39c8.md", "log/2021-01-03"}, Negate: false}},
		},
		[]string{"ref/test/a.md", "log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindLinkedByWithSnippets(t *testing.T) {
	testNoteDAOFind(t,
		note.FinderOpts{
			Filters: []note.Filter{note.LinkedByFilter{Paths: []string{"f39c8.md"}}},
		},
		[]note.Match{
			{
				Metadata: note.Metadata{
					Path:       "ref/test/a.md",
					Title:      "Another nested note",
					Lead:       "It shall appear before b.md",
					Body:       "It shall appear before b.md",
					RawContent: "#Another nested note\nIt shall appear before b.md",
					WordCount:  5,
					Links:      nil,
					Created:    time.Date(2019, 11, 20, 20, 32, 56, 0, time.UTC),
					Modified:   time.Date(2019, 11, 20, 20, 34, 6, 0, time.UTC),
					Checksum:   "iecywst",
				},
				Snippets: []string{
					"[[<zk:match>Link from 4 to 6</zk:match>]]",
					"[[<zk:match>Duplicated link</zk:match>]]",
				},
			},
			{
				Metadata: note.Metadata{
					Path:       "log/2021-01-03.md",
					Title:      "January 3, 2021",
					Lead:       "A daily note",
					Body:       "A daily note\n\nWith lot of content",
					RawContent: "# A daily note\nA daily note\n\nWith lot of content",
					WordCount:  3,
					Links:      nil,
					Created:    time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Modified:   time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Checksum:   "qwfpgj",
				},
				Snippets: []string{
					"[[<zk:match>Another link</zk:match>]]",
				},
			},
		},
	)
}

func TestNoteDAOFindNotLinkedBy(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.LinkedByFilter{Paths: []string{"f39c8.md", "log/2021-01-03"}, Negate: true}},
		},
		[]string{"ref/test/b.md", "f39c8.md", "log/2021-02-04.md", "index.md"},
	)
}

func TestNoteDAOFindLinkingTo(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.LinkingToFilter{Paths: []string{"log/2021-01-04", "ref/test/a.md"}, Negate: false}},
		},
		[]string{"f39c8.md", "log/2021-01-03.md"},
	)
}

func TestNoteDAOFindNotLinkingTo(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.LinkingToFilter{Paths: []string{"log/2021-01-04", "ref/test/a.md"}, Negate: true}},
		},
		[]string{"ref/test/b.md", "ref/test/a.md", "log/2021-02-04.md", "index.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindOrphan(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{note.OrphanFilter{}},
		},
		[]string{"ref/test/b.md", "f39c8.md", "log/2021-02-04.md", "index.md"},
	)
}

func TestNoteDAOFindCreatedOn(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 11, 22, 10, 12, 45, 0, time.UTC),
					Field:     note.DateCreated,
					Direction: note.DateOn,
				},
			},
		},
		[]string{"log/2021-01-03.md"},
	)
}

func TestNoteDAOFindCreatedBefore(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2019, 12, 04, 11, 59, 11, 0, time.UTC),
					Field:     note.DateCreated,
					Direction: note.DateBefore,
				},
			},
		},
		[]string{"ref/test/b.md", "ref/test/a.md"},
	)
}

func TestNoteDAOFindCreatedAfter(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Field:     note.DateCreated,
					Direction: note.DateAfter,
				},
			},
		},
		[]string{"log/2021-02-04.md", "log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindModifiedOn(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 01, 20, 10, 12, 45, 0, time.UTC),
					Field:     note.DateModified,
					Direction: note.DateOn,
				},
			},
		},
		[]string{"f39c8.md"},
	)
}

func TestNoteDAOFindModifiedBefore(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 01, 20, 8, 52, 42, 0, time.UTC),
					Field:     note.DateModified,
					Direction: note.DateBefore,
				},
			},
		},
		[]string{"ref/test/b.md", "ref/test/a.md", "index.md"},
	)
}

func TestNoteDAOFindModifiedAfter(t *testing.T) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Filters: []note.Filter{
				note.DateFilter{
					Date:      time.Date(2020, 11, 22, 16, 27, 45, 0, time.UTC),
					Field:     note.DateModified,
					Direction: note.DateAfter,
				},
			},
		},
		[]string{"log/2021-01-03.md", "log/2021-01-04.md"},
	)
}

func TestNoteDAOFindSortCreated(t *testing.T) {
	testNoteDAOFindSort(t, note.SortCreated, true, []string{
		"ref/test/b.md", "ref/test/a.md", "index.md", "f39c8.md",
		"log/2021-01-03.md", "log/2021-02-04.md", "log/2021-01-04.md",
	})
	testNoteDAOFindSort(t, note.SortCreated, false, []string{
		"log/2021-02-04.md", "log/2021-01-04.md", "log/2021-01-03.md",
		"f39c8.md", "index.md", "ref/test/b.md", "ref/test/a.md",
	})
}

func TestNoteDAOFindSortModified(t *testing.T) {
	testNoteDAOFindSort(t, note.SortModified, true, []string{
		"ref/test/b.md", "ref/test/a.md", "index.md", "f39c8.md",
		"log/2021-02-04.md", "log/2021-01-03.md", "log/2021-01-04.md",
	})
	testNoteDAOFindSort(t, note.SortModified, false, []string{
		"log/2021-01-04.md", "log/2021-01-03.md", "log/2021-02-04.md",
		"f39c8.md", "index.md", "ref/test/b.md", "ref/test/a.md",
	})
}

func TestNoteDAOFindSortPath(t *testing.T) {
	testNoteDAOFindSort(t, note.SortPath, true, []string{
		"f39c8.md", "index.md", "log/2021-01-03.md", "log/2021-01-04.md",
		"log/2021-02-04.md", "ref/test/a.md", "ref/test/b.md",
	})
	testNoteDAOFindSort(t, note.SortPath, false, []string{
		"ref/test/b.md", "ref/test/a.md", "log/2021-02-04.md",
		"log/2021-01-04.md", "log/2021-01-03.md", "index.md", "f39c8.md",
	})
}

func TestNoteDAOFindSortTitle(t *testing.T) {
	testNoteDAOFindSort(t, note.SortTitle, true, []string{
		"ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-02-04.md",
		"index.md", "log/2021-01-03.md", "log/2021-01-04.md",
	})
	testNoteDAOFindSort(t, note.SortTitle, false, []string{
		"log/2021-01-04.md", "log/2021-01-03.md", "index.md",
		"log/2021-02-04.md", "ref/test/a.md", "f39c8.md", "ref/test/b.md",
	})
}

func TestNoteDAOFindSortWordCount(t *testing.T) {
	testNoteDAOFindSort(t, note.SortWordCount, true, []string{
		"log/2021-01-03.md", "log/2021-02-04.md", "index.md",
		"log/2021-01-04.md", "f39c8.md", "ref/test/a.md", "ref/test/b.md",
	})
	testNoteDAOFindSort(t, note.SortWordCount, false, []string{
		"ref/test/b.md", "f39c8.md", "ref/test/a.md", "log/2021-02-04.md",
		"index.md", "log/2021-01-04.md", "log/2021-01-03.md",
	})
}

func testNoteDAOFindSort(t *testing.T, field note.SortField, ascending bool, expected []string) {
	testNoteDAOFindPaths(t,
		note.FinderOpts{
			Sorters: []note.Sorter{{Field: field, Ascending: ascending}},
		},
		expected,
	)
}

func testNoteDAOFindPaths(t *testing.T, opts note.FinderOpts, expected []string) {
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

func testNoteDAOFind(t *testing.T, opts note.FinderOpts, expected []note.Match) {
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

func testNoteDAOWithoutFixtures(t *testing.T, callback func(tx Transaction, dao *NoteDAO)) {
	testTransactionWithoutFixtures(t, func(tx Transaction) {
		callback(tx, NewNoteDAO(tx, &util.NullLogger))
	})
}

type noteRow struct {
	Path, Title, Lead, Body, RawContent, Checksum string
	WordCount                                     int
	Created, Modified                             time.Time
}

func queryNoteRow(tx Transaction, where string) (noteRow, error) {
	var row noteRow
	err := tx.QueryRow(fmt.Sprintf(`
		SELECT path, title, lead, body, raw_content, word_count, checksum, created, modified
		  FROM notes
		 WHERE %v
	`, where)).Scan(&row.Path, &row.Title, &row.Lead, &row.Body, &row.RawContent, &row.WordCount, &row.Checksum, &row.Created, &row.Modified)
	return row, err
}

type linkRow struct {
	SourceId                   int64
	TargetId                   *int64
	Href, Title, Rels, Snippet string
	External                   bool
}

func queryLinkRows(t *testing.T, tx Transaction, where string) []linkRow {
	links := make([]linkRow, 0)

	rows, err := tx.Query(fmt.Sprintf(`
		SELECT source_id, target_id, title, href, external, rels, snippet
		  FROM links
		 WHERE %v
		 ORDER BY id
	`, where))
	assert.Nil(t, err)

	for rows.Next() {
		var row linkRow
		err = rows.Scan(&row.SourceId, &row.TargetId, &row.Title, &row.Href, &row.External, &row.Rels, &row.Snippet)
		assert.Nil(t, err)
		links = append(links, row)
	}
	rows.Close()
	assert.Nil(t, rows.Err())

	return links
}

func intPointer(i int64) *int64 {
	return &i
}
