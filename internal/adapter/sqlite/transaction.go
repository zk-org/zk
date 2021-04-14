package sqlite

import "database/sql"

// Inspired by https://pseudomuto.com/2018/01/clean-sql-transactions-in-golang/

// Transaction is an interface that models the standard transaction in
// database/sql.
//
// To ensure TxFn funcs cannot commit or rollback a transaction (which is
// handled by `WithTransaction`), those methods are not included here.
type Transaction interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecStmts(stmts []string) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareLazy(query string) *LazyStmt
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// txWrapper wraps a native sql.Tx to fully implement the Transaction interface.
type txWrapper struct {
	*sql.Tx
}

func (tx *txWrapper) PrepareLazy(query string) *LazyStmt {
	return NewLazyStmt(tx.Tx, query)
}

func (tx *txWrapper) ExecStmts(stmts []string) error {
	var err error
	for _, stmt := range stmts {
		if err != nil {
			break
		}
		_, err = tx.Exec(stmt)
	}
	return err
}

// A Txfn is a function that will be called with an initialized Transaction
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

	err = fn(&txWrapper{tx})
	return err
}
