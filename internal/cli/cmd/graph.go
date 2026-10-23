package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/strings"
)

// Graph produces a directed graph of the notes matching a set of criteria.
type Graph struct {
	Format string
	Quiet  bool
	cli.Filtering
}

// GraphSpec describes the `zk graph` command.
func GraphSpec(container *cli.Container) *spec.Command {
	cmd := Graph{}
	arg := cmd.PathArg()
	arg.Complete = completeNotePaths(container)

	flags := []spec.Flag{
		{Name: "format", Short: "f", Group: spec.GroupFormat, Target: &cmd.Format,
			Help:     "Format of the graph among: json.",
			Required: true, OneOf: []string{"json"}, Complete: spec.Values("json")},
		{Name: "quiet", Short: "q", Group: spec.GroupFormat, Target: &cmd.Quiet,
			Help: "Do not print the total number of notes found."},
	}

	return &spec.Command{
		Name:  "graph",
		Short: "Produce a graph of the notes matching the given criteria.",
		Group: GroupNotes,
		Args:  []spec.Arg{arg},
		Flags: append(flags, filteringFlags(container, &cmd.Filtering)...),
		Run:   func() error { return cmd.Run(container) },
	}
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

	findOpts, err := cmd.NewNoteFindOpts(notebook)
	if err != nil {
		return fmt.Errorf("incorrect criteria: %w", err)
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
		Sorted:       len(findOpts.Sorters) > 0,
	})

	notes, err = filter.Apply(notes)
	if err != nil {
		if err == fzf.ErrCancelled {
			return nil
		}
		return err
	}

	fmt.Fprint(container.Stdout, "{\n  \"notes\": [\n")
	for i, note := range notes {
		if i > 0 {
			fmt.Fprint(container.Stdout, ",\n")
		}
		ft, err := format(note)
		if err != nil {
			return err
		}
		fmt.Fprintf(container.Stdout, "    %s", ft)
	}

	fmt.Fprint(container.Stdout, "\n  ],\n  \"links\": [\n")
	for i, link := range links {
		if i > 0 {
			fmt.Fprint(container.Stdout, ",\n")
		}
		ft, err := json.Marshal(link)
		if err != nil {
			return err
		}
		fmt.Fprintf(container.Stdout, "    %s", string(ft))
	}

	fmt.Fprint(container.Stdout, "\n  ]\n}\n")

	// FIXME; bad error check
	if err == nil && !cmd.Quiet {
		count := len(notes)
		fmt.Fprintf(container.Stderr, "\n\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return err
}
