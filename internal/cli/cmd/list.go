package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/util/strings"
)

// List displays notes matching a set of criteria.
type List struct {
	FormatFlags
	Link bool
	cli.Filtering
}

// ListSpec describes the `zk list` command.
func ListSpec(container *cli.Container) *spec.Command {
	cmd := List{}
	arg := cmd.PathArg()
	arg.Complete = completeNotePaths(container)

	flags := cmd.FormatFlags.Spec(FormatFlagsSpec{
		Noun:       "notes",
		FormatHelp: "Pretty print the list using a custom template or one of the predefined formats: oneline, short, medium, long, full, json, jsonl, uri.",
		Formats:    []string{"oneline", "short", "medium", "long", "full", "json", "jsonl", "uri", "path", "link"},
		Extra: []spec.Flag{
			{Name: "link", Group: spec.GroupFormat, Target: &cmd.Link,
				Help: "Print the zk:// URI of each note (shorthand for --format uri)."},
		},
	})

	return &spec.Command{
		Name:  "list",
		Short: "List notes matching the given criteria.",
		Group: GroupNotes,
		Args:  []spec.Arg{arg},
		Flags: append(flags, filteringFlags(container, &cmd.Filtering)...),
		Run:   func() error { return cmd.Run(container) },
	}
}

func (cmd *List) Run(container *cli.Container) error {
	if cmd.Link {
		if cmd.Format != "" {
			return errors.New("--link and --format can't be used together")
		}
		cmd.Format = "uri"
	}

	if err := cmd.FormatFlags.Prepare("[", "]\n", true); err != nil {
		return err
	}

	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	format, err := notebook.NewNoteFormatter(cmd.noteTemplate())
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

	count := len(notes)
	err = cmd.FormatFlags.Paginate(container, count, func(out io.Writer, i int) error {
		ft, err := format(notes[i])
		if err != nil {
			return err
		}
		fmt.Fprint(out, ft)
		return nil
	})
	if err != nil {
		return err
	}

	cmd.FormatFlags.PrintFound(container, count, strings.Pluralize("note", count))

	return nil
}

func (cmd *List) noteTemplate() string {
	return cmd.Template("short", defaultNoteFormats)
}

var defaultNoteFormats = map[string]string{
	"json":  `{{json .}}`,
	"jsonl": `{{json .}}`,
	"path":  `{{path}}`,
	"link":  `{{link}}`,
	"uri":   `{{uri}}`,

	"oneline": `{{style "title" title}} {{style "path" path}} ({{format-date created "elapsed"}})`,

	"short": `{{style "title" title}} {{style "path" path}} ({{format-date created "elapsed"}})

{{list snippets}}`,

	"medium": `{{style "title" title}} {{style "path" path}}
Created: {{format-date created "short"}}

{{list snippets}}`,

	"long": `{{style "title" title}} {{style "path" path}}
Created: {{format-date created "short"}}
Modified: {{format-date modified "short"}}

{{list snippets}}`,

	"full": `{{style "title" title}} {{style "path" path}}
Created: {{format-date created "short"}}
Modified: {{format-date modified "short"}}
Tags: {{join tags ", "}}

{{prepend "  " body}}
`,
}
