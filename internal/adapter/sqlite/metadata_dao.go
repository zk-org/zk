package sqlite

import (
	"database/sql"

	"github.com/mickael-menu/zk/internal/util/errors"
)

// Known metadata keys.
var reindexingRequiredKey = "zk.reindexing_required"

// MetadataDAO persists arbitrary key/value pairs in the SQLite database.
type MetadataDAO struct {
	tx Transaction

	// Prepared SQL statements
	getStmt *LazyStmt
	setStmt *LazyStmt
}

// NewMetadataDAO creates a new instance of a DAO working on the given
// database transaction.
func NewMetadataDAO(tx Transaction) *MetadataDAO {
	return &MetadataDAO{
		tx: tx,
		getStmt: tx.PrepareLazy(`
			SELECT key, value FROM metadata WHERE key = ?
		`),
		setStmt: tx.PrepareLazy(`
			INSERT OR REPLACE INTO metadata(key, value)
			VALUES (?, ?)
		`),
	}
}

// Get returns the value for the given key.
func (d *MetadataDAO) Get(key string) (string, error) {
	wrap := errors.Wrapperf("failed to get metadata with key %s", key)

	row, err := d.getStmt.QueryRow(key)
	if err != nil {
		return "", wrap(err)
	}

	var value string
	err = row.Scan(&key, &value)

	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", wrap(err)
	default:
		return value, nil
	}
}

// Set resets the value for the given metadata key.
func (d *MetadataDAO) Set(key string, value string) error {
	_, err := d.setStmt.Exec(key, value)
	return err
}
