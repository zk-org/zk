package sqlite

import (
	"database/sql"
	"sync"
)

// LazyStmt is a wrapper around a sql.Stmt which will be evaluated on first use.
type LazyStmt struct {
	create func() (*sql.Stmt, error)
	stmt   *sql.Stmt
	err    error
	once   sync.Once
}

// NewLazyStmt creates a new lazy statement bound to the given transaction.
func NewLazyStmt(tx *sql.Tx, query string) *LazyStmt {
	return &LazyStmt{
		create: func() (*sql.Stmt, error) { return tx.Prepare(query) },
	}
}

func (s *LazyStmt) Stmt() (*sql.Stmt, error) {
	s.once.Do(func() {
		s.stmt, s.err = s.create()
	})
	return s.stmt, s.err
}

func (s *LazyStmt) Exec(args ...interface{}) (sql.Result, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	return stmt.Exec(args...)
}

func (s *LazyStmt) Query(args ...interface{}) (*sql.Rows, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	return stmt.Query(args...)
}

func (s *LazyStmt) QueryRow(args ...interface{}) (*sql.Row, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	return stmt.QueryRow(args...), nil
}
