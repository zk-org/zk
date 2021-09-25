package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/internal/cli"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/os"
)

// New adds a new note to the notebook.
type New struct {
	Directory string            `arg optional default:"." help:"Directory in which to create the note."`
	Title     string            `short:t   placeholder:TITLE help:"Title of the new note."`
	Group     string            `short:g   placeholder:NAME  help:"Name of the config group this note belongs to. Takes precedence over the config of the directory."`
	Extra     map[string]string `                            help:"Extra variables passed to the templates." mapsep:","`
	Template  string            `          placeholder:PATH  help:"Custom template used to render the note."`
	PrintPath bool              `short:p                     help:"Print the path of the created note instead of editing it."`
}

func (cmd *New) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	content, err := os.ReadStdinPipe()
	if err != nil {
		return err
	}

	note, err := notebook.NewNote(core.NewNoteOpts{
		Title:     opt.NewNotEmptyString(cmd.Title),
		Content:   content.Unwrap(),
		Directory: opt.NewNotEmptyString(cmd.Directory),
		Group:     opt.NewNotEmptyString(cmd.Group),
		Template:  opt.NewNotEmptyString(cmd.Template),
		Extra:     cmd.Extra,
		Date:      time.Now(),
	})
	path := filepath.Join(notebook.Path, note.Path)
	if err != nil {
		var noteExists core.ErrNoteExists
		if !errors.As(err, &noteExists) {
			return err
		}

		if confirmed, _ := container.Terminal.Confirm(
			fmt.Sprintf("%s already exists, do you want to edit this note instead?", note.Path),
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
