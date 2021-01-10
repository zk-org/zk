package sqlite

import (
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/paths"
)

func TestNoteDAOIndexed(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		expected := []paths.Metadata{
			{
				Path:     "f39c8.md",
				Modified: date("2020-01-20T08:52:42+01:00"),
			},
			{
				Path:     "index.md",
				Modified: date("2019-12-04T12:17:21+01:00"),
			},
			{
				Path:     "log/2021-01-03.md",
				Modified: date("2020-11-22T16:27:45+01:00"),
			},
			{
				Path:     "log/2021-01-04.md",
				Modified: date("2020-11-29T08:20:18+01:00"),
			},
			{
				Path:     "ref/test/a.md",
				Modified: date("2019-11-20T20:34:06+01:00"),
			},
			{
				Path:     "ref/test/b.md",
				Modified: date("2019-11-20T20:34:06+01:00"),
			},
		}

		c, err := dao.Indexed()
		assert.Nil(t, err)

		i := 0
		for item := range c {
			assert.Equal(t, item, expected[i])
			i++
		}
	})
}

func TestNoteDAOAdd(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Add(note.Metadata{
			Path:      "log/added.md",
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Created:   date("2019-11-19T15:33:31+01:00"),
			Modified:  date("2020-01-16T16:04:59+01:00"),
			Checksum:  "check",
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `path = "log/added.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Path:      "log/added.md",
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Checksum:  "check",
			Created:   date("2019-11-19T15:33:31+01:00"),
			Modified:  date("2020-01-16T16:04:59+01:00"),
		})
	})
}

// Check that we can't add a duplicate note with an existing path.
func TestNoteDAOAddExistingNote(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Add(note.Metadata{Path: "ref/test/a.md"})
		assert.Err(t, err, "ref/test/a.md: can't add note to the index: UNIQUE constraint failed: notes.path")
	})
}

func TestNoteDAOUpdate(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Update(note.Metadata{
			Path:      "ref/test/a.md",
			Title:     "Updated note",
			Body:      "Updated body",
			WordCount: 42,
			Created:   date("2020-11-22T16:49:47+01:00"),
			Modified:  date("2020-11-22T16:49:47+01:00"),
			Checksum:  "updated checksum",
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `path = "ref/test/a.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Path:      "ref/test/a.md",
			Title:     "Updated note",
			Body:      "Updated body",
			WordCount: 42,
			Checksum:  "updated checksum",
			Created:   date("2019-11-20T20:32:56+01:00"),
			Modified:  date("2020-11-22T16:49:47+01:00"),
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

func TestNoteDAORemove(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Remove("ref/test/a.md")
		assert.Nil(t, err)
	})
}

func TestNoteDAORemoveUnknown(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		err := dao.Remove("unknown/unknown.md")
		assert.Err(t, err, "unknown/unknown.md: failed to remove note index: note not found in the index")
	})
}

func TestNoteDAOFindAll(t *testing.T) {
	testNoteDAOFind(t, []note.Match{
		note.Match{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "ref/test/b.md",
				Title:     "A nested note",
				Body:      "This one is in a sub sub directory",
				WordCount: 8,
				Created:   date("2019-11-20T20:32:56+01:00"),
				Modified:  date("2019-11-20T20:34:06+01:00"),
				Checksum:  "yvwbae",
			},
		},
		note.Match{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "f39c8.md",
				Title:     "An interesting note",
				Body:      "Its content will surprise you",
				WordCount: 5,
				Created:   date("2020-01-19T10:58:41+01:00"),
				Modified:  date("2020-01-20T08:52:42+01:00"),
				Checksum:  "irkwyc",
			},
		},
		note.Match{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "ref/test/a.md",
				Title:     "Another nested note",
				Body:      "It shall appear before b.md",
				WordCount: 5,
				Created:   date("2019-11-20T20:32:56+01:00"),
				Modified:  date("2019-11-20T20:34:06+01:00"),
				Checksum:  "iecywst",
			},
		},
		note.Match{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "index.md",
				Title:     "Index",
				Body:      "Index of the Zettelkasten",
				WordCount: 4,
				Created:   date("2019-12-04T11:59:11+01:00"),
				Modified:  date("2019-12-04T12:17:21+01:00"),
				Checksum:  "iaefhv",
			},
		},
		note.Match{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-01-03.md",
				Title:     "January 3, 2021",
				Body:      "A daily note",
				WordCount: 3,
				Created:   date("2020-11-22T16:27:45+01:00"),
				Modified:  date("2020-11-22T16:27:45+01:00"),
				Checksum:  "qwfpgj",
			},
		},
		note.Match{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-01-04.md",
				Title:     "January 4, 2021",
				Body:      "A second daily note",
				WordCount: 4,
				Created:   date("2020-11-29T08:20:18+01:00"),
				Modified:  date("2020-11-29T08:20:18+01:00"),
				Checksum:  "arstde",
			},
		},
	})
}

func testNoteDAOFind(t *testing.T, expected []note.Match) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		actual := make([]note.Match, 0)
		err := dao.Find(func(m note.Match) error {
			actual = append(actual, m)
			return nil
		})
		assert.Nil(t, err)

		popExpected := func() (note.Match, bool) {
			if len(expected) == 0 {
				return note.Match{}, false
			}
			item := expected[0]
			expected = expected[1:]
			return item, true
		}

		for _, act := range actual {
			exp, ok := popExpected()
			if !ok {
				t.Errorf("More matches than expected: %v", actual)
				return
			}
			assert.Equal(t, act, exp)
		}

		if len(expected) > 0 {
			t.Errorf("Missing expected matches: %v", expected)
		}
	})
}

func testNoteDAO(t *testing.T, callback func(tx Transaction, dao *NoteDAO)) {
	testTransaction(t, func(tx Transaction) {
		callback(tx, NewNoteDAO(tx, &util.NullLogger))
	})
}

type noteRow struct {
	Path, Title, Body, Checksum string
	WordCount                   int
	Created, Modified           time.Time
}

func queryNoteRow(tx Transaction, where string) (noteRow, error) {
	var row noteRow
	err := tx.QueryRow(fmt.Sprintf(`
		SELECT path, title, body, word_count, checksum, created, modified
		  FROM notes
		 WHERE %v
	`, where)).Scan(&row.Path, &row.Title, &row.Body, &row.WordCount, &row.Checksum, &row.Created, &row.Modified)
	return row, err
}

func date(s string) time.Time {
	date, _ := time.Parse(time.RFC3339, s)
	return date
}
