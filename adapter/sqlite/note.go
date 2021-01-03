package sqlite

import (
	"database/sql"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/core/file"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
)

// NoteDAO persists notes in the SQLite database.
// It implements the core ports note.Indexer and note.Finder.
type NoteDAO struct {
	tx     Transaction
	root   string
	logger util.Logger

	// Prepared SQL statements
	indexedStmt *sql.Stmt
	addStmt     *sql.Stmt
	updateStmt  *sql.Stmt
	removeStmt  *sql.Stmt
}

func NewNoteDAO(tx Transaction, root string, logger util.Logger) (*NoteDAO, error) {
	indexedStmt, err := tx.Prepare(`
		SELECT dir, filename, modified from notes
		 ORDER BY dir, filename ASC
	`)
	if err != nil {
		return nil, err
	}

	addStmt, err := tx.Prepare(`
		INSERT INTO notes (dir, filename, title, body, word_count, checksum, created, modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}

	updateStmt, err := tx.Prepare(`
		UPDATE notes
		   SET title = ?, body = ?, word_count = ?, checksum = ?, modified = ?
		 WHERE dir = ? AND filename = ?`)
	if err != nil {
		return nil, err
	}

	removeStmt, err := tx.Prepare(`
		DELETE FROM notes
		 WHERE dir = ? AND filename = ?
	`)
	if err != nil {
		return nil, err
	}

	return &NoteDAO{
		tx:          tx,
		root:        root,
		logger:      logger,
		indexedStmt: indexedStmt,
		addStmt:     addStmt,
		updateStmt:  updateStmt,
		removeStmt:  removeStmt,
	}, nil
}

func (d *NoteDAO) Indexed() (<-chan file.Metadata, error) {
	rows, err := d.indexedStmt.Query()
	if err != nil {
		return nil, err
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
				d.logger.Err(err)
			}

			c <- file.Metadata{
				Path:     file.Path{Dir: dir, Filename: filename, Abs: filepath.Join(d.root, dir, filename)},
				Modified: modified,
			}
		}

		err = rows.Err()
		if err != nil {
			d.logger.Err(err)
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
	return err
}

func (d *NoteDAO) Update(note note.Metadata) error {
	_, err := d.updateStmt.Exec(
		note.Title, note.Body, note.WordCount,
		note.Checksum, note.Modified,
		note.Path.Dir, note.Path.Filename,
	)
	return err
}

func (d *NoteDAO) Remove(path file.Path) error {
	_, err := d.updateStmt.Exec(path.Dir, path.Filename)
	return err
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
