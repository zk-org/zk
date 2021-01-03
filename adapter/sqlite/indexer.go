package sqlite

import (
	"database/sql"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/core/file"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util"
)

// NoteIndexer retrieves and stores notes indexation in the SQLite database.
// It implements the Core port note.Indexer.
type NoteIndexer struct {
	tx     *sql.Tx
	root   string
	logger util.Logger

	// Prepared SQL statements
	indexedStmt *sql.Stmt
	addStmt     *sql.Stmt
	updateStmt  *sql.Stmt
	removeStmt  *sql.Stmt
}

func NewNoteIndexer(tx *sql.Tx, root string, logger util.Logger) (*NoteIndexer, error) {
	indexedStmt, err := tx.Prepare(`
		SELECT filename, dir, modified from notes
		 ORDER BY dir, filename ASC
	`)
	if err != nil {
		return nil, err
	}

	addStmt, err := tx.Prepare(`
		INSERT INTO notes (filename, dir, title, body, word_count, checksum, created, modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}

	updateStmt, err := tx.Prepare(`
		UPDATE notes
		   SET title = ?, body = ?, word_count = ?, checksum = ?, modified = ?
		 WHERE filename = ? AND dir = ?
	`)
	if err != nil {
		return nil, err
	}

	removeStmt, err := tx.Prepare(`
		DELETE FROM notes
		 WHERE filename = ? AND dir = ?
	`)
	if err != nil {
		return nil, err
	}

	return &NoteIndexer{
		tx:          tx,
		root:        root,
		logger:      logger,
		indexedStmt: indexedStmt,
		addStmt:     addStmt,
		updateStmt:  updateStmt,
		removeStmt:  removeStmt,
	}, nil
}

func (i *NoteIndexer) Indexed() (<-chan file.Metadata, error) {
	rows, err := i.indexedStmt.Query()
	if err != nil {
		return nil, err
	}

	c := make(chan file.Metadata)
	go func() {
		defer close(c)
		defer rows.Close()
		var (
			filename, dir string
			modified      time.Time
		)

		for rows.Next() {
			err := rows.Scan(&filename, &dir, &modified)
			if err != nil {
				i.logger.Err(err)
			}

			c <- file.Metadata{
				Path:     file.Path{Dir: dir, Filename: filename, Abs: filepath.Join(i.root, dir, filename)},
				Modified: modified,
			}
		}

		err = rows.Err()
		if err != nil {
			i.logger.Err(err)
		}
	}()

	return c, nil
}

func (i *NoteIndexer) Add(metadata note.Metadata) error {
	_, err := i.addStmt.Exec(
		metadata.Path.Filename, metadata.Path.Dir, metadata.Title,
		metadata.Body, metadata.WordCount, metadata.Checksum,
		metadata.Created, metadata.Modified,
	)
	return err
}

func (i *NoteIndexer) Update(metadata note.Metadata) error {
	_, err := i.updateStmt.Exec(
		metadata.Title, metadata.Body, metadata.WordCount,
		metadata.Checksum, metadata.Modified,
		metadata.Path.Filename, metadata.Path.Dir,
	)
	return err
}

func (i *NoteIndexer) Remove(path file.Path) error {
	_, err := i.updateStmt.Exec(path.Filename, path.Dir)
	return err
}
