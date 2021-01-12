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
	Paths  []string `arg optional placeholder:"PATHS"`
	Format string   `help:"Pretty prints the list using the given format" placeholder:"TEMPLATE"`
	Match  string   `help:"Terms to search for in the notes" placeholder:"TERMS"`
	Limit  int      `help:"Limit the number of results" placeholder:"MAX"`
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

		paths := make([]string, 0)
		for _, p := range cmd.Paths {
			path, err := zk.RelPath(p)
			if err == nil {
				paths = append(paths, path)
			}
		}
		if len(paths) > 0 {
			filters = append(filters, note.PathFilter(paths))
		}

		if cmd.Match != "" {
			filters = append(filters, note.MatchFilter(cmd.Match))
		}

		count, err := note.List(
			note.ListOpts{
				Format: opt.NewNotEmptyString(cmd.Format),
				FinderOpts: note.FinderOpts{
					Filters: filters,
					Limit:   cmd.Limit,
				},
			},
			note.ListDeps{
				BasePath:  zk.Path,
				Finder:    notes,
				Templates: container.TemplateLoader(zk.Config.Lang),
			},
			printNote,
		)

		if err == nil {
			fmt.Printf("\nFound %d result(s)\n", count)
		}

		return err
	})
}

func printNote(note string) error {
	_, err := fmt.Println(note)
	return err
}
