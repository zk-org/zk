package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/internal/cli"
	"github.com/mickael-menu/zk/internal/core"
	dateutil "github.com/mickael-menu/zk/internal/util/date"
	"github.com/mickael-menu/zk/internal/util/opt"
)

// New adds a new note to the notebook.
type New struct {
	Directory   string            `arg optional default:"." help:"Directory in which to create the note."`
	Interactive bool              `short:i                  help:"Read contents from standard input."`
	Title       string            `short:t   placeholder:TITLE help:"Title of the new note."`
	Date        string            `          placeholder:DATE  help:"Set the current date."`
	Group       string            `short:g   placeholder:NAME  help:"Name of the config group this note belongs to. Takes precedence over the config of the directory."`
	Extra       map[string]string `                            help:"Extra variables passed to the templates." mapsep:","`
	Template    string            `          placeholder:PATH  help:"Custom template used to render the note."`
	PrintPath   bool              `short:p                     help:"Print the path of the created note instead of editing it."`
	DryRun      bool              `short:n                     help:"Don't actually create the note. Instead, prints its content on stdout and the generated path on stderr."`
	ID          string            `          placeholder:ID    help:"Skip id generation and use provided value."`
}

func (cmd *New) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	var content []byte
	if cmd.Interactive {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}

	date := time.Now()
	if cmd.Date != "" {
		date, err = dateutil.TimeFromNatural(cmd.Date)
		if err != nil {
			return err
		}
	}

	note, err := notebook.NewNote(core.NewNoteOpts{
		Title:     opt.NewNotEmptyString(cmd.Title),
		Content:   string(content),
		Directory: opt.NewNotEmptyString(cmd.Directory),
		Group:     opt.NewNotEmptyString(cmd.Group),
		Template:  opt.NewNotEmptyString(cmd.Template),
		Extra:     cmd.Extra,
		Date:      date,
		DryRun:    cmd.DryRun,
		ID:        cmd.ID,
	})

	if cmd.DryRun {
		if err != nil {
			return err
		}
		path := filepath.Join(notebook.Path, note.Path)
		fmt.Fprintln(os.Stderr, path)
		fmt.Print(note.RawContent)
		return nil
	}

	var path string
	if err == nil {
		path = filepath.Join(notebook.Path, note.Path)
	} else {
		var noteExists core.ErrNoteExists
		if !errors.As(err, &noteExists) {
			return err
		}

		if confirmed, _ := container.Terminal.Confirm(
			fmt.Sprintf("%s already exists, do you want to edit this note instead?", noteExists.Name),
			true,
		); !confirmed {
			// abort...
			return nil
		}

		path = noteExists.Path
	}

	if cmd.PrintPath {
		fmt.Printf("%+v\n", path)
		return nil
	} else {
		editor, err := container.NewNoteEditor(notebook)
		if err != nil {
			return err
		}
		return editor.Open(path)
	}
}
