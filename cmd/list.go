package cmd

import (
	"fmt"

	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/opt"
)

// List displays notes matching a set of criteria.
type List struct {
	Path   []string `arg optional placeholder:"PATHS"`
	Match  string   `help:"Terms to search for in the notes" placeholder:"TERMS"`
	Format string   `help:"Pretty prints the list using the given format" placeholder:"TEMPLATE"`
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
		notes := sqlite.NewNoteDAO(tx, container.Logger)

		filters := make([]note.Filter, 0)
		if cmd.Match != "" {
			filters = append(filters, note.MatchFilter(cmd.Match))
		}

		return note.List(
			note.ListOpts{
				Format:  opt.NewNotEmptyString(cmd.Format),
				Filters: filters,
			},
			note.ListDeps{
				BasePath:  zk.Path,
				Finder:    notes,
				Templates: container.TemplateLoader(zk.Config.Lang),
			},
			printNote,
		)
	})
}

func printNote(note string) error {
	_, err := fmt.Println(note)
	return err
}
