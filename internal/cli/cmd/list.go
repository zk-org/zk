package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mickael-menu/zk/internal/adapter/fzf"
	"github.com/mickael-menu/zk/internal/cli"
	"github.com/mickael-menu/zk/internal/util/errors"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
)

// List displays notes matching a set of criteria.
type List struct {
	Format     string `group:format short:f placeholder:TEMPLATE   help:"Pretty print the list using the given format."`
	Delimiter  string "group:format short:d default:\n             help:\"Print notes delimited by the given separator.\""
	Delimiter0 bool   "group:format short:0 name:delimiter0        help:\"Print notes delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.\""
	NoPager    bool   `group:format short:P help:"Do not pipe output into a pager."`
	Quiet      bool   `group:format short:q help:"Do not print the total number of notes found."`
	cli.Filtering
}

func (cmd *List) Run(container *cli.Container) error {
	if cmd.Delimiter0 {
		cmd.Delimiter = "\x00"
	}

	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	format, err := notebook.NewNoteFormatter(cmd.noteTemplate())
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
		PreviewCmd:   container.Config.Tool.FzfPreview,
		NotebookDir:  notebook.Path,
		WorkingDir:   container.WorkingDir,
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
			if cmd.Delimiter0 {
				fmt.Fprint(out, "\x00")
			}

			return nil
		})
	}

	if err == nil && !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\n\nFound %d %s\n", count, strutil.Pluralize("note", count))
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
		templ = format
		// Replace raw \n and \t by actual newlines and tabs in user format.
		templ = strings.ReplaceAll(templ, "\\n", "\n")
		templ = strings.ReplaceAll(templ, "\\t", "\t")
	}

	return templ
}

var defaultNoteFormats = map[string]string{
	"path": `{{path}}`,

	"oneline": `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})`,

	"short": `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{list snippets}}`,

	"medium": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}

{{list snippets}}`,

	"long": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{list snippets}}`,

	"full": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}
Tags: {{join tags ", "}}

{{prepend "  " body}}
`,
}
