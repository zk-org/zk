package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
)

// URI prints the zk:// URI of a note, to link to it from outside systems.
type URI struct {
	Path   string `arg placeholder:PATH help:"Path of the note in the notebook."`
	Line   int    `short:l placeholder:LINE help:"Line to point at, starting at 1."`
	Column int    `short:c placeholder:COLUMN help:"Column to point at, starting at 1."`
}

func (cmd *URI) Run(container *cli.Container) error {
	if cmd.Line < 0 || cmd.Column < 0 {
		return errors.New("--line and --column must be positive")
	}
	if cmd.Column > 0 && cmd.Line == 0 {
		return errors.New("--column requires --line")
	}

	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	// RelPath resolves the path against the working directory and rejects
	// anything outside the notebook.
	path, err := notebook.RelPath(cmd.Path)
	if err != nil {
		return err
	}

	exists, err := container.FS.FileExists(filepath.Join(notebook.Path, path))
	if err != nil {
		return err
	}
	if !exists {
		// Not an existing path, so treat the argument as an href and look it up
		// in the index, as the other commands taking a note do.
		note, err := notebook.FindByHref(cmd.Path, true)
		if err != nil {
			return err
		}
		if note == nil {
			return fmt.Errorf("%s: note not found", cmd.Path)
		}
		path = note.Path
	}

	uri := core.NoteURI{
		NotebookPath: notebook.Path,
		Path:         path,
		Line:         cmd.Line,
		Column:       cmd.Column,
	}
	fmt.Println(uri.String())
	return nil
}
