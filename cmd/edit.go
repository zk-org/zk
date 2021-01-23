package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
)

// Edit opens notes matching a set of criteria with the user editor.
type Edit struct {
	Filtering `embed`
	Sorting   `embed`
	Force     bool `help:"Don't confirm before editing many notes at the same time" short:"f"`
}

func (cmd *Edit) Run(container *Container) error {
	zk, err := zk.Open(".")
	if err != nil {
		return err
	}

	opts, err := NewFinderOpts(zk, cmd.Filtering, cmd.Sorting)
	if err != nil {
		return errors.Wrapf(err, "incorrect criteria")
	}

	db, err := container.Database(zk.DBPath())
	if err != nil {
		return err
	}

	var notes []note.Match
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		notes, err = container.NoteFinder(tx).Find(*opts)
		return err
	})
	if err != nil {
		return err
	}

	count := len(notes)

	if count > 0 {
		if !cmd.Force && count > 2 {
			if !container.TTY.Confirm(
				fmt.Sprintf("Are you sure you want to open %v notes in the editor?", count),
				"Open all the notes",
				"Don't open any note",
			) {
				return nil
			}
		}
		paths := make([]string, 0)
		for _, note := range notes {
			absPath := filepath.Join(zk.Path, note.Path)
			paths = append(paths, absPath)
		}

		note.Edit(zk, paths...)
	}

	return err
}
