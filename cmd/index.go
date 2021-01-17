package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/schollz/progressbar/v3"
)

// Index indexes the content of all the notes in the slip box.
type Index struct {
	Directory string `arg optional type:"path" default:"." help:"Directory containing the notes to index"`
	Force     bool   `help:"Force indexing all the notes" short:"f"`
	Quiet     bool   `help:"Do not print statistics nor progress" short:"q"`
}

func (cmd *Index) Run(container *Container) error {
	zk, err := zk.Open(".")
	if err != nil {
		return err
	}

	dir, err := zk.RequireDirAt(cmd.Directory)
	if err != nil {
		return err
	}

	db, err := container.Database(zk.DBPath())
	if err != nil {
		return err
	}

	var bar *progressbar.ProgressBar
	if !cmd.Quiet {
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSpinnerType(14),
		)
	}

	var stats note.IndexingStats
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		notes := sqlite.NewNoteDAO(tx, container.Logger)

		stats, err = note.Index(
			*dir,
			cmd.Force,
			container.Parser(),
			notes,
			container.Logger,
			func(change paths.DiffChange) {
				if bar != nil {
					bar.Add(1)
					bar.Describe(change.String())
				}
			},
		)
		return err
	})

	if bar != nil {
		bar.Clear()
	}

	if err == nil && !cmd.Quiet {
		fmt.Println(stats)
	}

	return err
}
