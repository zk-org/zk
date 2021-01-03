package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mickael-menu/zk/util/errors"
)

// DB holds the connections to a SQLite database.
type DB struct {
	*sql.DB
}

// Open creates a new DB instance for the SQLite database at the given path.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", "file:"+path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the database")
	}
	return &DB{db}, nil
}

// Close terminates the connections to the SQLite database.
func (db *DB) Close() error {
	err := db.Close()
	return errors.Wrap(err, "failed to close the database")
}

// Migrate upgrades the SQL schema of the database.
func (db *DB) Migrate() error {
	wrap := errors.Wrapper("database migration failed")

	tx, err := db.Begin()
	if err != nil {
		return wrap(err)
	}
	defer tx.Rollback()

	var version int
	err = tx.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return wrap(err)
	}

	if version == 0 {
		err = execMultiple(tx, []string{
			`
				CREATE TABLE IF NOT EXISTS notes (
					id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
					filename TEXT NOT NULL,
					dir TEXT NOT NULL,
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
	}
	if err != nil {
		return wrap(err)
	}

	err = tx.Commit()
	if err != nil {
		return wrap(err)
	}

	return nil
}

func execMultiple(tx *sql.Tx, stmts []string) error {
	var err error
	for _, stmt := range stmts {
		if err != nil {
			break
		}
		_, err = tx.Exec(stmt)
	}
	return err
}
