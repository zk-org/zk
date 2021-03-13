package sqlite

import (
	"database/sql"
	"sync"

	"github.com/mickael-menu/zk/util/errors"
)

// LazyStmt is a wrapper around a sql.Stmt which will be evaluated on first use.
type LazyStmt struct {
	query  string
	create func() (*sql.Stmt, error)
	stmt   *sql.Stmt
	err    error
	once   sync.Once
}

// NewLazyStmt creates a new lazy statement bound to the given transaction.
func NewLazyStmt(tx *sql.Tx, query string) *LazyStmt {
	return &LazyStmt{
		query:  query,
		create: func() (*sql.Stmt, error) { return tx.Prepare(query) },
	}
}

func (s *LazyStmt) Stmt() (*sql.Stmt, error) {
	s.once.Do(func() {
		s.stmt, s.err = s.create()
	})
	return s.stmt, s.wrapErr(s.err)
}

func (s *LazyStmt) Exec(args ...interface{}) (sql.Result, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	res, err := stmt.Exec(args...)
	return res, s.wrapErr(err)
}

func (s *LazyStmt) Query(args ...interface{}) (*sql.Rows, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(args...)
	return rows, s.wrapErr(err)
}

func (s *LazyStmt) QueryRow(args ...interface{}) (*sql.Row, error) {
	stmt, err := s.Stmt()
	if err != nil {
		return nil, err
	}
	return stmt.QueryRow(args...), nil
}

func (s *LazyStmt) wrapErr(err error) error {
	return errors.Wrapf(err, "database query: %s", s.query)
}
