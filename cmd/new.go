package cmd

import (
	"fmt"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/os"
)

// New adds a new note to the slip box.
type New struct {
	Directory string            `arg optional type:"path" default:"." help:"Directory in which to create the note"`
	PrintPath bool              `help:"Prints the path of the created note to stdin instead of editing it"`
	Title     string            `short:"t" help:"Title of the new note" placeholder:"TITLE"`
	Template  string            `type:"path" help:"Custom template to use to render the note" placeholder:"PATH"`
	Extra     map[string]string `help:"Extra variables passed to the templates"`
}

func (cmd *New) ConfigOverrides() zk.ConfigOverrides {
	return zk.ConfigOverrides{
		BodyTemplatePath: opt.NewNotEmptyString(cmd.Template),
		Extra:            cmd.Extra,
	}
}

func (cmd *New) Run(container *Container) error {
	zk, err := zk.Open(".")
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
		Dir:     *dir,
		Title:   opt.NewNotEmptyString(cmd.Title),
		Content: content,
	}
	file, err := note.Create(opts, container.TemplateLoader())
	if err != nil {
		return err
	}

	if cmd.PrintPath {
		fmt.Printf("%+v\n", file)
		return nil
	} else {
		return note.Edit(zk, file)
	}
}
