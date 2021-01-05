package sqlite

import (
	"fmt"
	"testing"
	"time"

	"github.com/mickael-menu/zk/core/file"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/assert"
)

func TestNoteDAOIndexed(t *testing.T) {
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		expected := []file.Metadata{
			{
				Path:     file.Path{Dir: "", Filename: "f39c8.md", Abs: "/test/f39c8.md"},
				Modified: date("2020-01-20T08:52:42.321024+01:00"),
			},
			{
				Path:     file.Path{Dir: "", Filename: "index.md", Abs: "/test/index.md"},
				Modified: date("2019-12-04T12:17:21.720747+01:00"),
			},
			{
				Path:     file.Path{Dir: "log", Filename: "2021-01-03.md", Abs: "/test/log/2021-01-03.md"},
				Modified: date("2020-11-22T16:27:45.734454655+01:00"),
			},
			{
				Path:     file.Path{Dir: "log", Filename: "2021-01-04.md", Abs: "/test/log/2021-01-04.md"},
				Modified: date("2020-11-29T08:20:18.138907236+01:00"),
			},
			{
				Path:     file.Path{Dir: "ref/test", Filename: "a.md", Abs: "/test/ref/test/a.md"},
				Modified: date("2019-11-20T20:34:06.120375+01:00"),
			},
			{
				Path:     file.Path{Dir: "ref/test", Filename: "b.md", Abs: "/test/ref/test/b.md"},
				Modified: date("2019-11-20T20:34:06.120375+01:00"),
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
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		err := dao.Add(note.Metadata{
			Path: file.Path{
				Dir:      "log",
				Filename: "added.md",
				Abs:      "/test/log/added.md",
			},
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Created:   date("2019-11-19T15:33:31.467036963+01:00"),
			Modified:  date("2020-01-16T16:04:59.396405+01:00"),
			Checksum:  "check",
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `filename = "added.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Dir:       "log",
			Filename:  "added.md",
			Title:     "Added note",
			Body:      "Note body",
			WordCount: 2,
			Checksum:  "check",
			Created:   date("2019-11-19T15:33:31.467036963+01:00"),
			Modified:  date("2020-01-16T16:04:59.396405+01:00"),
		})
	})
}

// Check that we can't add a duplicate note with an existing path.
func TestNoteDAOAddExistingNote(t *testing.T) {
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		err := dao.Add(note.Metadata{
			Path: file.Path{
				Dir:      "ref/test",
				Filename: "a.md",
			},
		})
		assert.Err(t, err, "ref/test/a.md: can't add note to the index: UNIQUE constraint failed: notes.filename, notes.dir")
	})
}

func TestNoteDAOUpdate(t *testing.T) {
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		err := dao.Update(note.Metadata{
			Path: file.Path{
				Dir:      "ref/test",
				Filename: "a.md",
				Abs:      "/test/log/added.md",
			},
			Title:     "Updated note",
			Body:      "Updated body",
			WordCount: 42,
			Created:   date("2020-11-22T16:49:47.309530098+01:00"),
			Modified:  date("2020-11-22T16:49:47.309769915+01:00"),
			Checksum:  "updated checksum",
		})
		assert.Nil(t, err)

		row, err := queryNoteRow(tx, `dir = "ref/test" AND filename = "a.md"`)
		assert.Nil(t, err)
		assert.Equal(t, row, noteRow{
			Dir:       "ref/test",
			Filename:  "a.md",
			Title:     "Updated note",
			Body:      "Updated body",
			WordCount: 42,
			Checksum:  "updated checksum",
			Created:   date("2019-11-20T20:32:56.107028961+01:00"),
			Modified:  date("2020-11-22T16:49:47.309769915+01:00"),
		})
	})
}

func TestNoteDAOUpdateUnknown(t *testing.T) {
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		err := dao.Update(note.Metadata{
			Path: file.Path{
				Dir:      "unknown",
				Filename: "unknown.md",
				Abs:      "/test/unknown/unknown.md",
			},
		})
		assert.Err(t, err, "unknown/unknown.md: failed to update note index: note not found in the index")
	})
}

func TestNoteDAORemove(t *testing.T) {
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		err := dao.Remove(file.Path{
			Dir:      "ref/test",
			Filename: "a.md",
			Abs:      "/test/ref/test/a.md",
		})
		assert.Nil(t, err)
	})
}

func TestNoteDAORemoveUnknown(t *testing.T) {
	testTransaction(t, func(tx Transaction) {
		dao := NewNoteDAO(tx, "/test", &util.NullLogger)

		err := dao.Remove(file.Path{
			Dir:      "unknown",
			Filename: "unknown.md",
			Abs:      "/test/unknown/unknown.md",
		})
		assert.Err(t, err, "unknown/unknown.md: failed to remove note index: note not found in the index")
	})
}

type noteRow struct {
	Dir, Filename, Title, Body, Checksum string
	WordCount                            int
	Created, Modified                    time.Time
}

func queryNoteRow(tx Transaction, where string) (noteRow, error) {
	var row noteRow
	err := tx.QueryRow(fmt.Sprintf(`
		SELECT dir, filename, title, body, word_count, checksum, created, modified
		  FROM notes
		 WHERE %v
	`, where)).Scan(&row.Dir, &row.Filename, &row.Title, &row.Body, &row.WordCount, &row.Checksum, &row.Created, &row.Modified)
	return row, err
}

func date(s string) time.Time {
	date, _ := time.Parse(time.RFC3339, s)
	return date
}
