package sqlite

import "database/sql"

// Inspired by https://pseudomuto.com/2018/01/clean-sql-transactions-in-golang/

// Transaction is an interface that models the standard transaction in
// database/sql.
//
// To ensure TxFn funcs cannot commit or rollback a transaction (which is
// handled by `WithTransaction`), those methods are not included here.
type Transaction interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// TxFn is a function that will be called with an initialized Transaction
// object that can be used for executing statements and queries against a
// database.
type TxFn func(tx Transaction) error

// WithTransaction creates a new transaction and handles rollback/commit based
// on the error object returned by the TxFn closure.
func (db *DB) WithTransaction(fn TxFn) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			// A panic occurred, rollback and repanic.
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}
