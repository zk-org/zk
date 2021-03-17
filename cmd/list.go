package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mickael-menu/zk/adapter/fzf"
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/strings"
)

// List displays notes matching a set of criteria.
type List struct {
	Format     string `group:format short:f placeholder:TEMPLATE help:"Pretty print the list using the given format."`
	Delimiter  string "group:format short:d default:\n             help:\"Print notes delimited by the given separator.\""
	Delimiter0 bool   "group:format short:0 name:delimiter0        help:\"Print notes delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.\""
	NoPager    bool   `group:format short:P help:"Do not pipe output into a pager."`
	Quiet      bool   `group:format short:q help:"Do not print the total number of notes found."`

	Filtering
	Sorting
}

func (cmd *List) Run(container *Container) error {
	if cmd.Delimiter0 {
		cmd.Delimiter = "\x00"
	}

	zk, err := container.Zk()
	if err != nil {
		return err
	}

	opts, err := NewFinderOpts(zk, cmd.Filtering, cmd.Sorting)
	if err != nil {
		return err
	}

	db, _, err := container.Database(false)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	templates := container.TemplateLoader(container.Config.Note.Lang)
	styler := container.Terminal
	format := opt.NewNotEmptyString(cmd.Format)
	formatter, err := note.NewFormatter(zk.Path, wd, format, templates, styler)
	if err != nil {
		return err
	}

	var notes []note.Match
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		finder := container.NoteFinder(tx, fzf.NoteFinderOpts{
			AlwaysFilter: false,
			PreviewCmd:   container.Config.Tool.FzfPreview,
			BasePath:     zk.Path,
			CurrentPath:  wd,
		})
		notes, err = finder.Find(*opts)
		return err
	})
	if err != nil {
		if err == note.ErrCanceled {
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

				ft, err := formatter.Format(note)
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
		fmt.Fprintf(os.Stderr, "\n\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return err
}
