package sqlite

import (
	"database/sql"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/fts5"
	"github.com/mickael-menu/zk/util/paths"
)

// NoteDAO persists notes in the SQLite database.
// It implements the core ports note.Indexer and note.Finder.
type NoteDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	indexedStmt *LazyStmt
	addStmt     *LazyStmt
	updateStmt  *LazyStmt
	removeStmt  *LazyStmt
	existsStmt  *LazyStmt
}

func NewNoteDAO(tx Transaction, logger util.Logger) *NoteDAO {
	return &NoteDAO{
		tx:     tx,
		logger: logger,
		indexedStmt: tx.PrepareLazy(`
			SELECT path, modified from notes
			 ORDER BY path ASC
		`),
		addStmt: tx.PrepareLazy(`
			INSERT INTO notes (path, title, body, word_count, checksum, created, modified)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`),
		updateStmt: tx.PrepareLazy(`
			UPDATE notes
			   SET title = ?, body = ?, word_count = ?, checksum = ?, modified = ?
			 WHERE path = ?
		`),
		removeStmt: tx.PrepareLazy(`
			DELETE FROM notes
			 WHERE path = ?
		`),
		existsStmt: tx.PrepareLazy(`
			SELECT EXISTS (SELECT 1 FROM notes WHERE path = ?)
		`),
	}
}

func (d *NoteDAO) Indexed() (<-chan paths.Metadata, error) {
	wrap := errors.Wrapper("failed to get indexed notes")

	rows, err := d.indexedStmt.Query()
	if err != nil {
		return nil, wrap(err)
	}

	c := make(chan paths.Metadata)
	go func() {
		defer close(c)
		defer rows.Close()
		var (
			path     string
			modified time.Time
		)

		for rows.Next() {
			err := rows.Scan(&path, &modified)
			if err != nil {
				d.logger.Err(wrap(err))
			}

			c <- paths.Metadata{
				Path:     path,
				Modified: modified,
			}
		}

		err = rows.Err()
		if err != nil {
			d.logger.Err(wrap(err))
		}
	}()

	return c, nil
}

func (d *NoteDAO) Add(note note.Metadata) error {
	_, err := d.addStmt.Exec(
		note.Path, note.Title, note.Body, note.WordCount, note.Checksum,
		note.Created, note.Modified,
	)
	return errors.Wrapf(err, "%v: can't add note to the index", note.Path)
}

func (d *NoteDAO) Update(note note.Metadata) error {
	wrap := errors.Wrapperf("%v: failed to update note index", note.Path)

	exists, err := d.exists(note.Path)
	if err != nil {
		return wrap(err)
	}
	if !exists {
		return wrap(errors.New("note not found in the index"))
	}

	_, err = d.updateStmt.Exec(
		note.Title, note.Body, note.WordCount, note.Checksum, note.Modified,
		note.Path,
	)
	return errors.Wrapf(err, "%v: failed to update note index", note.Path)
}

func (d *NoteDAO) Remove(path string) error {
	wrap := errors.Wrapperf("%v: failed to remove note index", path)

	exists, err := d.exists(path)
	if err != nil {
		return wrap(err)
	}
	if !exists {
		return wrap(errors.New("note not found in the index"))
	}

	_, err = d.removeStmt.Exec(path)
	return wrap(err)
}

func (d *NoteDAO) exists(path string) (bool, error) {
	row, err := d.existsStmt.QueryRow(path)
	if err != nil {
		return false, err
	}
	var exists bool
	row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (d *NoteDAO) Find(callback func(note.Match) error, filters ...note.Filter) error {
	rows, err := func() (*sql.Rows, error) {
		if len(filters) == 0 {
			return d.tx.Query(`
				SELECT id, path, title, body, word_count, created, modified,
					   checksum, "" as snippet from notes
				 ORDER BY title ASC
			`)
		} else {
			filter := filters[0].(note.MatchFilter)
			return d.tx.Query(`
				SELECT n.id, n.path, n.title, n.body, n.word_count,
				       n.created, n.modified, n.checksum,
					   snippet(notes_fts, -1, '<zk:match>', '</zk:match>', '…', 20) as snippet
				  FROM notes n
				  JOIN notes_fts
					ON n.id = notes_fts.rowid
				 WHERE notes_fts MATCH ?
				 ORDER BY bm25(notes_fts, 1000.0, 500.0, 1.0)
				 --- ORDER BY rank
			`, fts5.ConvertQuery(string(filter)))
		}
	}()

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, wordCount        int
			title, body, snippet string
			path, checksum       string
			created, modified    time.Time
		)

		err := rows.Scan(&id, &path, &title, &body, &wordCount, &created, &modified, &checksum, &snippet)
		if err != nil {
			d.logger.Err(err)
			continue
		}

		callback(note.Match{
			Snippet: snippet,
			Metadata: note.Metadata{
				Path:      path,
				Title:     title,
				Body:      body,
				WordCount: wordCount,
				Created:   created,
				Modified:  modified,
				Checksum:  checksum,
			},
		})
	}

	return nil
}
