package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/strings"
)

// List displays notes matching a set of criteria.
type List struct {
	Format    string `help:"Pretty prints the list using the given format" short:"f" placeholder:"<template>"`
	NoPager   bool   `help:"Do not pipe zk output into a pager" short:"P"`
	Quiet     bool   `help:"Don't show anything besides the notes themselves" short:"q"`
	Delimiter string `default:"
" help:"Delimiter separating each result" short:"d"`
	Filtering
	Sorting
}

func (cmd *List) Run(container *Container) error {
	zk, err := container.OpenZk()
	if err != nil {
		return err
	}

	opts, err := NewFinderOpts(zk, cmd.Filtering, cmd.Sorting)
	if err != nil {
		return errors.Wrapf(err, "incorrect criteria")
	}

	db, err := container.Database(zk.DBPath())
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	templates := container.TemplateLoader(zk.Config.Lang)
	styler := container.Terminal
	format := opt.NewNotEmptyString(cmd.Format)
	formatter, err := note.NewFormatter(zk.Path, wd, format, templates, styler)
	if err != nil {
		return err
	}

	var notes []note.Match
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		notes, err = container.NoteFinder(tx).Find(*opts)
		return err
	})
	if err != nil {
		return err
	}

	count := len(notes)
	if count > 0 {
		err = container.Paginate(cmd.NoPager, zk.Config, func(out io.Writer) error {
			for i, note := range notes {
				if i > 0 {
					fmt.Fprintf(out, cmd.Delimiter)
				}

				ft, err := formatter.Format(note)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, ft)
			}

			return nil
		})
	}

	if err == nil && !cmd.Quiet {
		fmt.Printf("\n\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return err
}
