package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/paths"
	"github.com/schollz/progressbar/v3"
)

// Index indexes the content of all the notes in the notebook.
type Index struct {
	Force   bool `short:"f" help:"Force indexing all the notes."`
	Verbose bool `short:"v" xor:"print" help:"Print detailed information about the indexing process."`
	Quiet   bool `short:"q" xor:"print" help:"Do not print statistics nor progress."`
}

func (cmd *Index) Help() string {
	return "You usually do not need to run `zk index` manually, as notes are indexed automatically when needed."
}

func (cmd *Index) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	return cmd.RunWithNotebook(container, notebook)
}

func (cmd *Index) RunWithNotebook(container *cli.Container, notebook *core.Notebook) error {
	showProgress := container.Terminal.IsInteractive()

	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSpinnerType(14),
		)
	}

	opts := core.NoteIndexOpts{
		Force:   cmd.Force,
		Verbose: cmd.Verbose,
	}

	stats, err := notebook.IndexWithCallback(opts, func(change paths.DiffChange) {
		if showProgress {
			bar.Add(1)
			bar.Describe(change.String())
		}
	})

	if showProgress {
		bar.Clear()
	}

	if err != nil {
		return err
	}

	if !cmd.Quiet {
		fmt.Println(stats)
	}

	return nil
}
