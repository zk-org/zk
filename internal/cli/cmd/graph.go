package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/strings"
)

// Graph produces a directed graph of the notes matching a set of criteria.
type Graph struct {
	Format string `group:format short:f                        help:"Format of the graph among: json." enum:"json" required`
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
	noteIDs := []core.NoteID{}
	for _, note := range notes {
		noteIDs = append(noteIDs, note.ID)
	}
	links, err := notebook.FindLinksBetweenNotes(noteIDs)
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

	fmt.Print("{\n  \"notes\": [\n")
	for i, note := range notes {
		if i > 0 {
			fmt.Print(",\n")
		}
		ft, err := format(note)
		if err != nil {
			return err
		}
		fmt.Printf("    %s", ft)
	}

	fmt.Print("\n  ],\n  \"links\": [\n")
	for i, link := range links {
		if i > 0 {
			fmt.Print(",\n")
		}
		ft, err := json.Marshal(link)
		if err != nil {
			return err
		}
		fmt.Printf("    %s", string(ft))
	}

	fmt.Print("\n  ]\n}\n")

	if err == nil && !cmd.Quiet {
		count := len(notes)
		fmt.Fprintf(os.Stderr, "\n\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return err
}
