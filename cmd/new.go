package cmd

import (
	"errors"
	"fmt"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/os"
)

// New adds a new note to the notebook.
type New struct {
	Directory string `arg optional type:"path" default:"." help:"Directory in which to create the note."`

	Title     string            `short:t   placeholder:TITLE help:"Title of the new note."`
	Group     string            `short:g   placeholder:NAME  help:"Name of the config group this note belongs to. Takes precedence over the config of the directory."`
	Extra     map[string]string `                            help:"Extra variables passed to the templates." mapsep:","`
	Template  string            `type:path placeholder:PATH  help:"Custom template used to render the note."`
	PrintPath bool              `short:p                     help:"Print the path of the created note instead of editing it."`
}

func (cmd *New) ConfigOverrides() zk.ConfigOverrides {
	return zk.ConfigOverrides{
		Group:            opt.NewNotEmptyString(cmd.Group),
		BodyTemplatePath: opt.NewNotEmptyString(cmd.Template),
		Extra:            cmd.Extra,
	}
}

func (cmd *New) Run(container *Container) error {
	zk, err := container.OpenZk()
	if err != nil {
		return err
	}

	dir, err := zk.RequireDirAt(cmd.Directory, cmd.ConfigOverrides())
	if err != nil {
		return err
	}

	content, err := os.ReadStdinPipe()
	if err != nil {
		return err
	}

	opts := note.CreateOpts{
		Config:  zk.Config,
		Dir:     *dir,
		Title:   opt.NewNotEmptyString(cmd.Title),
		Content: content,
	}

	file, err := note.Create(opts, container.TemplateLoader(dir.Config.Note.Lang), container.Date)
	if err != nil {
		var noteExists note.ErrNoteExists
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

		file = noteExists.Path
	}

	if cmd.PrintPath {
		fmt.Printf("%+v\n", file)
		return nil
	} else {
		return note.Edit(zk, file)
	}
}
