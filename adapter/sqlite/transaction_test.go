package sqlite

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/test/assert"
)

// testTransaction is an utility function used to test a SQLite transaction to
// the DB, which loads the default set of DB fixtures.
func testTransaction(t *testing.T, test func(tx Transaction)) {
	testTransactionWithFixtures(t, opt.NewString("default"), test)
}

// testTransaction is an utility function used to test a SQLite transaction to
// an empty DB.
func testTransactionWithoutFixtures(t *testing.T, test func(tx Transaction)) {
	testTransactionWithFixtures(t, opt.NullString, test)
}

// testTransactionWithFixtures is an utility function used to test a SQLite transaction to
// the DB, which loads the given set of DB fixtures.
func testTransactionWithFixtures(t *testing.T, fixturesDir opt.String, test func(tx Transaction)) {
	db, err := OpenInMemory()
	assert.Nil(t, err)
	_, err = db.Migrate()
	assert.Nil(t, err)

	if !fixturesDir.IsNull() {
		fixtures, err := testfixtures.New(
			testfixtures.Database(db.db),
			testfixtures.Dialect("sqlite"),
			testfixtures.Directory("fixtures/"+fixturesDir.String()),
			// Necessary to work with an in-memory database.
			testfixtures.DangerousSkipTestDatabaseCheck(),
		)
		assert.Nil(t, err)
		err = fixtures.Load()
		assert.Nil(t, err)
	}

	err = db.WithTransaction(func(tx Transaction) error {
		test(tx)
		return nil
	})
	assert.Nil(t, err)
}

func assertExistOrNot(t *testing.T, tx Transaction, shouldExist bool, sql string, args ...interface{}) {
	if shouldExist {
		assertExist(t, tx, sql, args...)
	} else {
		assertNotExist(t, tx, sql, args...)
	}
}

func assertExist(t *testing.T, tx Transaction, sql string, args ...interface{}) {
	if !exists(t, tx, sql, args...) {
		t.Errorf("SQL query did not return any result: %s, with arguments %v", sql, args)
	}
}

func assertNotExist(t *testing.T, tx Transaction, sql string, args ...interface{}) {
	if exists(t, tx, sql, args...) {
		t.Errorf("SQL query returned a result: %s, with arguments %v", sql, args)
	}
}

func exists(t *testing.T, tx Transaction, sql string, args ...interface{}) bool {
	var exists int
	err := tx.QueryRow("SELECT EXISTS ("+sql+")", args...).Scan(&exists)
	assert.Nil(t, err)
	return exists == 1
}
