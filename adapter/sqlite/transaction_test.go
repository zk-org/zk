package sqlite

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/mickael-menu/zk/util/assert"
)

// testTransaction is an utility function used to test a SQLite transaction to
// the DB.
func testTransaction(t *testing.T, test func(tx Transaction)) {
	db, err := OpenInMemory()
	assert.Nil(t, err)
	err = db.Migrate()
	assert.Nil(t, err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db.db),
		testfixtures.Dialect("sqlite"),
		testfixtures.Directory("fixtures"),
		// Necessary to work with an in-memory database.
		testfixtures.DangerousSkipTestDatabaseCheck(),
	)
	assert.Nil(t, err)
	err = fixtures.Load()
	assert.Nil(t, err)

	err = db.WithTransaction(func(tx Transaction) error {
		test(tx)
		return nil
	})
	assert.Nil(t, err)
}
