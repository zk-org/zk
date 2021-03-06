package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/schollz/progressbar/v3"
)

// Index indexes the content of all the notes in the notebook.
type Index struct {
	Force bool `short:"f" help:"Force indexing all the notes."`
	Quiet bool `short:"q" help:"Do not print statistics nor progress."`
}

func (cmd *Index) Help() string {
	return "You usually don't need to run `zk index` manually, as notes are indexed automatically before each zk invocation."
}

func (cmd *Index) Run(container *Container) error {
	zk, err := container.OpenZk()
	if err != nil {
		return err
	}

	db, err := container.Database(zk.DBPath())
	if err != nil {
		return err
	}

	var bar = progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSpinnerType(14),
	)

	var stats note.IndexingStats
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		stats, err = note.Index(
			zk,
			cmd.Force,
			container.Parser(),
			container.NoteIndexer(tx),
			container.Logger,
			func(change paths.DiffChange) {
				bar.Add(1)
				bar.Describe(change.String())
			},
		)
		return err
	})
	bar.Clear()

	if err == nil && !cmd.Quiet {
		fmt.Println(stats)
	}

	return err
}
