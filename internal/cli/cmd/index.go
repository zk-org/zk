package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/paths"
)

// Index indexes the content of all the notes in the notebook.
type Index struct {
	Force   bool
	Verbose bool
	Quiet   bool
	// Automatic is set when the index is refreshed before a command, as
	// opposed to an explicit `zk index`. It is not a command line flag.
	Automatic bool
}

// IndexSpec describes the `zk index` command.
func IndexSpec(container *cli.Container) *spec.Command {
	cmd := Index{}

	return &spec.Command{
		Name:  "index",
		Short: "Index the notes to be searchable.",
		Long:  "You usually do not need to run `zk index` manually, as notes are indexed automatically when needed.",
		Group: GroupNotebook,
		// Indexing the notebook first would hide the statistics this command
		// is run for.
		SkipAutoIndex: true,
		Flags: []spec.Flag{
			{Name: "force", Short: "f", Target: &cmd.Force,
				Help: "Force indexing all the notes."},
			{Name: "verbose", Short: "v", Target: &cmd.Verbose,
				Help: "Print detailed information about the indexing process."},
			{Name: "quiet", Short: "q", Target: &cmd.Quiet,
				Help: "Do not print statistics nor progress."},
		},
		Run: func() error {
			if cmd.Verbose && cmd.Quiet {
				return errors.New("--verbose and --quiet can't be used together")
			}
			return cmd.Run(container)
		},
	}
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
			progressbar.OptionSetWriter(container.Stderr),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSpinnerType(14),
		)
	}

	opts := core.NoteIndexOpts{
		Force:     cmd.Force,
		Verbose:   cmd.Verbose,
		Automatic: cmd.Automatic,
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
		fmt.Fprintln(container.Stdout, stats)
	}

	return nil
}
