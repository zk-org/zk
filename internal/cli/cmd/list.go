package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/strings"
)

// List displays notes matching a set of criteria.
type List struct {
	Format     string `group:format short:f placeholder:TEMPLATE   help:"Pretty print the list using a custom template or one of the predefined formats: oneline, short, medium, long, full, json, jsonl."`
	Header     string `group:format                                help:"Arbitrary text printed at the start of the list."`
	Footer     string `group:format default:\n                     help:"Arbitrary text printed at the end of the list."`
	Delimiter  string "group:format short:d default:\n             help:\"Print notes delimited by the given separator.\""
	Delimiter0 bool   "group:format short:0 name:delimiter0        help:\"Print notes delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.\""
	NoPager    bool   `group:format short:P help:"Do not pipe output into a pager."`
	Quiet      bool   `group:format short:q help:"Do not print the total number of notes found."`
	cli.Filtering
}

func (cmd *List) Run(container *cli.Container) error {
	cmd.Header = strings.ExpandWhitespaceLiterals(cmd.Header)
	cmd.Footer = strings.ExpandWhitespaceLiterals(cmd.Footer)
	cmd.Delimiter = strings.ExpandWhitespaceLiterals(cmd.Delimiter)

	if cmd.Delimiter0 {
		if cmd.Delimiter != "\n" {
			return errors.New("--delimiter and --delimiter0 can't be used together")
		}
		if cmd.Header != "" {
			return errors.New("--footer and --delimiter0 can't be used together")
		}
		if cmd.Footer != "\n" {
			return errors.New("--footer and --delimiter0 can't be used together")
		}

		cmd.Delimiter = "\x00"
		cmd.Footer = "\x00"
	}

	if cmd.Format == "json" || cmd.Format == "jsonl" {
		if cmd.Header != "" {
			return errors.New("--header can't be used with JSON format")
		}
		if cmd.Footer != "\n" {
			return errors.New("--footer can't be used with JSON format")
		}
		if cmd.Delimiter != "\n" {
			return errors.New("--delimiter can't be used with JSON format")
		}

		switch cmd.Format {
		case "json":
			cmd.Delimiter = ","
			cmd.Header = "["
			cmd.Footer = "]\n"

		case "jsonl":
			// > The last character in the file may be a line separator, and it
			// > will be treated the same as if there was no line separator
			// > present.
			// > https://jsonlines.org/
			cmd.Footer = "\n"
		}
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
