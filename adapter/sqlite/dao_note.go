package sqlite

import (
	"database/sql"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/core/file"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
)

// NoteDAO persists notes in the SQLite database.
// It implements the core ports note.Indexer and note.Finder.
type NoteDAO struct {
	tx     Transaction
	root   string
	logger util.Logger

	// Prepared SQL statements
	indexedStmt *LazyStmt
	addStmt     *LazyStmt
	updateStmt  *LazyStmt
	removeStmt  *LazyStmt
	existsStmt  *LazyStmt
}

func NewNoteDAO(tx Transaction, root string, logger util.Logger) *NoteDAO {
	return &NoteDAO{
		tx:     tx,
		root:   root,
		logger: logger,
		indexedStmt: tx.PrepareLazy(`
			SELECT dir, filename, modified from notes
			 ORDER BY dir, filename ASC
		`),
		addStmt: tx.PrepareLazy(`
			INSERT INTO notes (dir, filename, title, body, word_count, checksum, created, modified)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`),
		updateStmt: tx.PrepareLazy(`
			UPDATE notes
			   SET title = ?, body = ?, word_count = ?, checksum = ?, modified = ?
			 WHERE dir = ? AND filename = ?
		`),
		removeStmt: tx.PrepareLazy(`
			DELETE FROM notes
			 WHERE dir = ? AND filename = ?
		`),
		existsStmt: tx.PrepareLazy(`
			SELECT EXISTS (SELECT 1 FROM notes WHERE dir = ? AND filename = ?)
		`),
	}
}

func (d *NoteDAO) Indexed() (<-chan file.Metadata, error) {
	wrap := errors.Wrapper("failed to get indexed notes")

	rows, err := d.indexedStmt.Query()
	if err != nil {
		return nil, wrap(err)
	}

	c := make(chan file.Metadata)
	go func() {
		defer close(c)
		defer rows.Close()
		var (
			dir, filename string
			modified      time.Time
		)

		for rows.Next() {
			err := rows.Scan(&dir, &filename, &modified)
			if err != nil {
				d.logger.Err(wrap(err))
			}

			c <- file.Metadata{
				Path:     file.Path{Dir: dir, Filename: filename, Abs: filepath.Join(d.root, dir, filename)},
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
		note.Path.Dir, note.Path.Filename, note.Title,
		note.Body, note.WordCount, note.Checksum,
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
		note.Title, note.Body, note.WordCount,
		note.Checksum, note.Modified,
		note.Path.Dir, note.Path.Filename,
	)
	return errors.Wrapf(err, "%v: failed to update note index", note.Path)
}

func (d *NoteDAO) Remove(path file.Path) error {
	wrap := errors.Wrapperf("%v: failed to remove note index", path)

	exists, err := d.exists(path)
	if err != nil {
		return wrap(err)
	}
	if !exists {
		return wrap(errors.New("note not found in the index"))
	}

	_, err = d.removeStmt.Exec(path.Dir, path.Filename)
	return wrap(err)
}

func (d *NoteDAO) exists(path file.Path) (bool, error) {
	row, err := d.existsStmt.QueryRow(path.Dir, path.Filename)
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
				SELECT id, dir, filename, title, body, word_count, created, modified,
					   checksum, "" as snippet from notes
				 ORDER BY title ASC
			`)
		} else {
			filter := filters[0].(note.QueryFilter)
			return d.tx.Query(`
				SELECT n.id, n.dir, n.filename, n.title, n.body, n.word_count,
				       n.created, n.modified, n.checksum,
					   snippet(notes_fts, -1, '\033[31m', '\033[0m', '…', 20) as snippet
				  FROM notes n
				  JOIN notes_fts
					ON n.id = notes_fts.rowid
				 WHERE notes_fts MATCH ?
				 ORDER BY bm25(notes_fts, 1000.0, 1.0)
				 --- ORDER BY rank
			`, filter)
		}
	}()

	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, wordCount           int
			title, body, snippet    string
			dir, filename, checksum string
			created, modified       time.Time
		)

		err := rows.Scan(&id, &dir, &filename, &title, &body, &wordCount, &created, &modified, &checksum, &snippet)
		if err != nil {
			d.logger.Err(err)
			continue
		}

		callback(note.Match{
			ID:      id,
			Snippet: snippet,
			Metadata: note.Metadata{
				Path: file.Path{
					Dir:      dir,
					Filename: filename,
					Abs:      filepath.Join(d.root, dir, filename),
				},
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
