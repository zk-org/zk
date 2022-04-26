package sqlite

import (
	"testing"

	"github.com/mickael-menu/zk/internal/util/fixtures"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestOpen(t *testing.T) {
	_, err := Open(fixtures.Path("sample.db"))
	assert.Nil(t, err)
}

func TestClose(t *testing.T) {
	db, err := Open(fixtures.Path("sample.db"))
	assert.Nil(t, err)
	err = db.Close()
	assert.Nil(t, err)
}

func TestMigrateFrom0(t *testing.T) {
	db, err := OpenInMemory()
	assert.Nil(t, err)

	err = db.WithTransaction(func(tx Transaction) error {
		var version int
		err := tx.QueryRow("PRAGMA user_version").Scan(&version)
		assert.Nil(t, err)
		assert.Equal(t, version, 7)

		_, err = tx.Exec(`
			INSERT INTO notes (path, sortable_path, title, body, word_count, checksum)
			VALUES ("ref/tx1.md", "reftx1.md", "A reference", "Content", 1, "qwfpg")
		`)
		assert.Nil(t, err)

		return nil
	})
	assert.Nil(t, err)
}
