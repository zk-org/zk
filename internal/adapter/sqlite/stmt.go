package sqlite

import (
	"database/sql"
	"fmt"
	"sync"
)

// Preparer is the subset of a transaction or connection able to prepare SQL
// statements.
type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

// LazyStmt is a wrapper around a sql.Stmt which will be evaluated on first use.
type LazyStmt struct {
	query  string
	create func() (*sql.Stmt, error)
	stmt   *sql.Stmt
	err    error
	once   sync.Once
}

// NewLazyStmt creates a new lazy statement bound to the given Preparer.
func NewLazyStmt(p Preparer, query string) *LazyStmt {
	return &LazyStmt{
		query:  query,
		create: func() (*sql.Stmt, error) { return p.Prepare(query) },
	}
}

func (s *LazyStmt) Stmt() (*sql.Stmt, error) {
	s.once.Do(func() {
		s.stmt, s.err = s.create()
	})
	return s.stmt, s.wrapErr(s.err)
}

func (s *LazyStmt) Exec(args ...any) (sql.Result, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	res, err := stmt.Exec(args...)
	return res, s.wrapErr(err)
}

func (s *LazyStmt) Query(args ...any) (*sql.Rows, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(args...)
	return rows, s.wrapErr(err)
}

func (s *LazyStmt) QueryRow(args ...any) (*sql.Row, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	return stmt.QueryRow(args...), nil
}

func (s *LazyStmt) wrapErr(err error) error {
	if err != nil {
		return fmt.Errorf("database query: %s: %w", s.query, err)
	}
	return nil
}
