package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/strings"
)

// List displays notes matching a set of criteria.
type List struct {
	FormatFlags
	cli.Filtering
}

func (cmd *List) Run(container *cli.Container) error {
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
		err = container.Paginate(cmd.NoPager, func(out io.Writer) error {
			if cmd.Header != "" {
				fmt.Fprint(out, cmd.Header)
			}
			for i, note := range notes {
				if i > 0 {
					fmt.Fprint(out, cmd.Delimiter)
				}

				ft, err := format(note)
				if err != nil {
					return err
				}
				fmt.Fprint(out, ft)
			}
			if cmd.Footer != "" {
				fmt.Fprint(out, cmd.Footer)
			}

			return nil
		})
	}

	if err == nil && !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return err
}

func (cmd *List) noteTemplate() string {
	format := cmd.Format
	if format == "" {
		format = "short"
	}

	templ, ok := defaultNoteFormats[format]
	if !ok {
		templ = strings.ExpandWhitespaceLiterals(format)
	}

	return templ
}

var defaultNoteFormats = map[string]string{
	"json":  `{{json .}}`,
	"jsonl": `{{json .}}`,
	"path":  `{{path}}`,
	"link":  `{{link}}`,

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
