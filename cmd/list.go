package cmd

import (
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
)

// List displays notes matching a set of criteria.
type List struct {
	Query string `arg optional help:"Terms to search for in the notes" placeholder:"TERMS"`
}

func (cmd *List) Run(container *Container) error {
	zk, err := zk.Open(".")
	if err != nil {
		return err
	}

	db, err := container.Database(zk.DBPath())
	if err != nil {
		return err
	}

	return db.WithTransaction(func(tx sqlite.Transaction) error {
		notes := sqlite.NewNoteDAO(tx, zk.Path, container.Logger)

		filters := make([]note.Filter, 0)
		if cmd.Query != "" {
			filters = append(filters, note.QueryFilter(cmd.Query))
		}

		return note.List(notes, filters...)
	})
}
