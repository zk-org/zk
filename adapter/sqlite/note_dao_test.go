package sqlite

import (
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/mickael-menu/zk/util/test"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestNoteDAOIndexed(t *testing.T) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		expected := []paths.Metadata{
			{
				Path:     "f39c8.md",
				Modified: time.Date(2020, 1, 20, 8, 52, 42, 0, time.Local),
			},
			{
				Path:     "index.md",
				Modified: time.Date(2019, 12, 4, 12, 17, 21, 0, time.Local),
			},
			{
				Path:     "log/2021-01-03.md",
				Modified: time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
			},
			{
				Path:     "log/2021-01-04.md",
				Modified: time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
			},
			{
				Path:     "log/2021-02-04.md",
				Modified: time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
			},
			{
				Path:     "ref/test/a.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.Local),
			},
			{
				Path:     "ref/test/b.md",
				Modified: time.Date(2019, 11, 20, 20, 34, 6, 0, time.Local),
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
		err := dao.Add(note.Metadata{
			Path:      "log/added.md",
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Created:   test.Date("2019-11-19T15:33:31+01:00"),
			Modified:  test.Date("2020-01-16T16:04:59+01:00"),
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
			Created:   test.Date("2019-11-19T15:33:31+01:00"),
			Modified:  test.Date("2020-01-16T16:04:59+01:00"),
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
			Created:   test.Date("2020-11-22T16:49:47+01:00"),
			Modified:  test.Date("2020-11-22T16:49:47+01:00"),
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
			Created:   test.Date("2019-11-20T20:32:56+01:00"),
			Modified:  test.Date("2020-11-22T16:49:47+01:00"),
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
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "ref/test/b.md",
				Title:     "A nested note",
				Body:      "This one is in a sub sub directory",
				WordCount: 8,
				Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.Local),
				Modified:  time.Date(2019, 11, 20, 20, 34, 6, 0, time.Local),
				Checksum:  "yvwbae",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "f39c8.md",
				Title:     "An interesting note",
				Body:      "Its content will surprise you",
				WordCount: 5,
				Created:   time.Date(2020, 1, 19, 10, 58, 41, 0, time.Local),
				Modified:  time.Date(2020, 1, 20, 8, 52, 42, 0, time.Local),
				Checksum:  "irkwyc",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "ref/test/a.md",
				Title:     "Another nested note",
				Body:      "It shall appear before b.md",
				WordCount: 5,
				Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.Local),
				Modified:  time.Date(2019, 11, 20, 20, 34, 6, 0, time.Local),
				Checksum:  "iecywst",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-02-04.md",
				Title:     "February 4, 2021",
				Body:      "A third daily note",
				WordCount: 4,
				Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Modified:  time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Checksum:  "earkte",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "index.md",
				Title:     "Index",
				Body:      "Index of the Zettelkasten",
				WordCount: 4,
				Created:   time.Date(2019, 12, 4, 11, 59, 11, 0, time.Local),
				Modified:  time.Date(2019, 12, 4, 12, 17, 21, 0, time.Local),
				Checksum:  "iaefhv",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-01-03.md",
				Title:     "January 3, 2021",
				Body:      "A daily note",
				WordCount: 3,
				Created:   time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
				Modified:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
				Checksum:  "qwfpgj",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-01-04.md",
				Title:     "January 4, 2021",
				Body:      "A second daily note",
				WordCount: 4,
				Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Modified:  time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Checksum:  "arstde",
			},
		},
	})
}

func TestNoteDAOFindMatch(t *testing.T) {
	expected := []note.Match{
		{
			Snippet: "<zk:match>Index</zk:match> of the Zettelkasten",
			Metadata: note.Metadata{
				Path:      "index.md",
				Title:     "Index",
				Body:      "Index of the Zettelkasten",
				WordCount: 4,
				Created:   time.Date(2019, 12, 4, 11, 59, 11, 0, time.Local),
				Modified:  time.Date(2019, 12, 4, 12, 17, 21, 0, time.Local),
				Checksum:  "iaefhv",
			},
		},
		{
			Snippet: "A <zk:match>daily</zk:match> note",
			Metadata: note.Metadata{
				Path:      "log/2021-01-03.md",
				Title:     "January 3, 2021",
				Body:      "A daily note",
				WordCount: 3,
				Created:   time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
				Modified:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
				Checksum:  "qwfpgj",
			},
		},
		{
			Snippet: "A second <zk:match>daily</zk:match> note",
			Metadata: note.Metadata{
				Path:      "log/2021-01-04.md",
				Title:     "January 4, 2021",
				Body:      "A second daily note",
				WordCount: 4,
				Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Modified:  time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Checksum:  "arstde",
			},
		},
		{
			Snippet: "A third <zk:match>daily</zk:match> note",
			Metadata: note.Metadata{
				Path:      "log/2021-02-04.md",
				Title:     "February 4, 2021",
				Body:      "A third daily note",
				WordCount: 4,
				Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Modified:  time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Checksum:  "earkte",
			},
		},
	}

	testNoteDAOFind(t, expected, note.MatchFilter("daily | index"))
}

func TestNoteDAOFindInPath(t *testing.T) {
	expected := []note.Match{
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-01-03.md",
				Title:     "January 3, 2021",
				Body:      "A daily note",
				WordCount: 3,
				Created:   time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
				Modified:  time.Date(2020, 11, 22, 16, 27, 45, 0, time.Local),
				Checksum:  "qwfpgj",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "log/2021-01-04.md",
				Title:     "January 4, 2021",
				Body:      "A second daily note",
				WordCount: 4,
				Created:   time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Modified:  time.Date(2020, 11, 29, 8, 20, 18, 0, time.Local),
				Checksum:  "arstde",
			},
		},
	}

	testNoteDAOFind(t, expected, note.PathFilter([]string{"log/2021-01-*"}))
}

func TestNoteDAOFindInMultiplePath(t *testing.T) {
	expected := []note.Match{
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "ref/test/b.md",
				Title:     "A nested note",
				Body:      "This one is in a sub sub directory",
				WordCount: 8,
				Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.Local),
				Modified:  time.Date(2019, 11, 20, 20, 34, 6, 0, time.Local),
				Checksum:  "yvwbae",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "ref/test/a.md",
				Title:     "Another nested note",
				Body:      "It shall appear before b.md",
				WordCount: 5,
				Created:   time.Date(2019, 11, 20, 20, 32, 56, 0, time.Local),
				Modified:  time.Date(2019, 11, 20, 20, 34, 6, 0, time.Local),
				Checksum:  "iecywst",
			},
		},
		{
			Snippet: "",
			Metadata: note.Metadata{
				Path:      "index.md",
				Title:     "Index",
				Body:      "Index of the Zettelkasten",
				WordCount: 4,
				Created:   time.Date(2019, 12, 4, 11, 59, 11, 0, time.Local),
				Modified:  time.Date(2019, 12, 4, 12, 17, 21, 0, time.Local),
				Checksum:  "iaefhv",
			},
		},
	}

	testNoteDAOFind(t, expected, note.PathFilter([]string{"ref", "index.md"}))
}

func testNoteDAOFind(t *testing.T, expected []note.Match, filters ...note.Filter) {
	testNoteDAO(t, func(tx Transaction, dao *NoteDAO) {
		actual := make([]note.Match, 0)
		err := dao.Find(func(m note.Match) error {
			actual = append(actual, m)
			return nil
		}, filters...)
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
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
