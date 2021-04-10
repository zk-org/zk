package sqlite

import (
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

// testDB is an utility function to create a database loaded with the default fixtures.
func testDB(t *testing.T) *DB {
	return testDBWithFixtures(t, opt.NewString("default"))
}

// testDB is an utility function to create a database loaded with a set of DB fixtures.
func testDBWithFixtures(t *testing.T, fixturesDir opt.String) *DB {
	db, err := OpenInMemory()
	assert.Nil(t, err)
	_, err = db.Migrate()
	assert.Nil(t, err)

	if !fixturesDir.IsNull() {
		fixtures, err := testfixtures.New(
			testfixtures.Database(db.db),
			testfixtures.Dialect("sqlite"),
			testfixtures.Directory("testdata/"+fixturesDir.String()),
			// Necessary to work with an in-memory database.
			testfixtures.DangerousSkipTestDatabaseCheck(),
		)
		assert.Nil(t, err)
		err = fixtures.Load()
		assert.Nil(t, err)
	}

	return db
}

// testTransaction is an utility function used to test a SQLite transaction to
// the DB, which loads the default set of DB fixtures.
func testTransaction(t *testing.T, test func(tx Transaction)) {
	testTransactionWithFixtures(t, opt.NewString("default"), test)
}

// testTransactionWithFixtures is an utility function used to test a SQLite transaction to
// the DB, which loads the given set of DB fixtures.
func testTransactionWithFixtures(t *testing.T, fixturesDir opt.String, test func(tx Transaction)) {
	err := testDBWithFixtures(t, fixturesDir).WithTransaction(func(tx Transaction) error {
		test(tx)
		return nil
	})
	assert.Nil(t, err)
}

func assertExistOrNot(t *testing.T, db *DB, shouldExist bool, sql string, args ...interface{}) {
	if shouldExist {
		assertExist(t, db, sql, args...)
	} else {
		assertNotExist(t, db, sql, args...)
	}
}

func assertExist(t *testing.T, db *DB, sql string, args ...interface{}) {
	if !exists(t, db, sql, args...) {
		t.Errorf("SQL query did not return any result: %s, with arguments %v", sql, args)
	}
}

func assertNotExist(t *testing.T, db *DB, sql string, args ...interface{}) {
	if exists(t, db, sql, args...) {
		t.Errorf("SQL query returned a result: %s, with arguments %v", sql, args)
	}
}

func exists(t *testing.T, db *DB, sql string, args ...interface{}) bool {
	var exists int
	err := db.db.QueryRow("SELECT EXISTS ("+sql+")", args...).Scan(&exists)
	assert.Nil(t, err)
	return exists == 1
}

// FIXME: Still needed?
func assertExistOrNotTx(t *testing.T, tx Transaction, shouldExist bool, sql string, args ...interface{}) {
	if shouldExist {
		assertExistTx(t, tx, sql, args...)
	} else {
		assertNotExistTx(t, tx, sql, args...)
	}
}

func assertExistTx(t *testing.T, tx Transaction, sql string, args ...interface{}) {
	if !existsTx(t, tx, sql, args...) {
		t.Errorf("SQL query did not return any result: %s, with arguments %v", sql, args)
	}
}

func assertNotExistTx(t *testing.T, tx Transaction, sql string, args ...interface{}) {
	if existsTx(t, tx, sql, args...) {
		t.Errorf("SQL query returned a result: %s, with arguments %v", sql, args)
	}
}

func existsTx(t *testing.T, tx Transaction, sql string, args ...interface{}) bool {
	var exists int
	err := tx.QueryRow("SELECT EXISTS ("+sql+")", args...).Scan(&exists)
	assert.Nil(t, err)
	return exists == 1
}
