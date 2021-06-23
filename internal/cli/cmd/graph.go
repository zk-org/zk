package cmd

import (
	"fmt"
	"os"

	"github.com/mickael-menu/zk/internal/adapter/fzf"
	"github.com/mickael-menu/zk/internal/cli"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/strings"
)

// Graph produces a directed graph of the notes matching a set of criteria.
type Graph struct {
	Format string `group:format short:f                        help:"Format of the graph among: json." enum:"json" default:"json"`
	Quiet  bool   `group:format short:q help:"Do not print the total number of notes found."`
	cli.Filtering
}

func (cmd *Graph) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	format, err := notebook.NewNoteFormatter("{{json .}}")
	if err != nil {
		return err
	}

	findOpts, err := cmd.Filtering.NewNoteFindOpts(notebook)
	if err != nil {
		return errors.Wrapf(err, "incorrect criteria")
	}

	notes, err := notebook.FindNotes(findOpts)
	if err != nil {
		return err
	}

	filter := container.NewNoteFilter(fzf.NoteFilterOpts{
		Interactive:  cmd.Interactive,
		AlwaysFilter: false,
		NotebookDir:  notebook.Path,
	})

	notes, err = filter.Apply(notes)
	if err != nil {
		if err == fzf.ErrCancelled {
			return nil
		}
		return err
	}

	count := len(notes)
	if count > 0 {
		for i, note := range notes {
			if i > 0 {
				fmt.Println()
			}

			ft, err := format(note)
			if err != nil {
				return err
			}
			fmt.Print(ft)
		}
	}

	if err == nil && !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\n\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return err
}
