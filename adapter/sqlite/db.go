package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mickael-menu/zk/util/errors"
)

// DB holds the connections to a SQLite database.
type DB struct {
	db *sql.DB
}

// Open creates a new DB instance for the SQLite database at the given path.
func Open(path string) (*DB, error) {
	return open("file:" + path)
}

// OpenInMemory creates a new in-memory DB instance.
func OpenInMemory() (*DB, error) {
	return open(":memory:")
}

func open(uri string) (*DB, error) {
	db, err := sql.Open("sqlite3", uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the database")
	}
	return &DB{db}, nil
}

// Close terminates the connections to the SQLite database.
func (db *DB) Close() error {
	err := db.db.Close()
	return errors.Wrap(err, "failed to close the database")
}

// Migrate upgrades the SQL schema of the database.
func (db *DB) Migrate() error {
	err := db.WithTransaction(func(tx Transaction) error {
		var version int
		err := tx.QueryRow("PRAGMA user_version").Scan(&version)
		if err != nil {
			return err
		}

		if version == 0 {
			err = tx.ExecStmts([]string{
				`
					CREATE TABLE IF NOT EXISTS notes (
						id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
						dir TEXT NOT NULL,
						filename TEXT NOT NULL,
						title TEXT DEFAULT('') NOT NULL,
						body TEXT DEFAULT('') NOT NULL,
						word_count INTEGER DEFAULT(0) NOT NULL,
						checksum TEXT NOT NULL,
						created DATETIME DEFAULT(CURRENT_TIMESTAMP) NOT NULL,
						modified DATETIME DEFAULT(CURRENT_TIMESTAMP) NOT NULL,
						UNIQUE(filename, dir)
					)
				`,
				`CREATE INDEX IF NOT EXISTS notes_checksum_idx ON notes(checksum)`,
				`
					CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
						title, body,
						content = notes,
						content_rowid = id,
						tokenize = 'porter unicode61 remove_diacritics 1'
					)
				`,
				// Triggers to keep the FTS index up to date.
				`
					CREATE TRIGGER IF NOT EXISTS notes_ai AFTER INSERT ON notes BEGIN
						INSERT INTO notes_fts(rowid, title, body) VALUES (new.id, new.title, new.body);
					END
				`,
				`
					CREATE TRIGGER IF NOT EXISTS notes_ad AFTER DELETE ON notes BEGIN
						INSERT INTO notes_fts(notes_fts, rowid, title, body) VALUES('delete', old.id, old.title, old.body);
					END
				`,
				`
					CREATE TRIGGER IF NOT EXISTS notes_au AFTER UPDATE ON notes BEGIN
						INSERT INTO notes_fts(notes_fts, rowid, title, body) VALUES('delete', old.id, old.title, old.body);
						INSERT INTO notes_fts(rowid, title, body) VALUES (new.id, new.title, new.body);
					END
				`,
				`PRAGMA user_version = 1`,
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	return errors.Wrap(err, "database migration failed")
}
