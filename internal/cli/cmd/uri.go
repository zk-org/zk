package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
)

// URI prints the zk:// URI of a note, to link to it from outside systems.
type URI struct {
	Path   string
	Line   int
	Column int
}

// URISpec describes the `zk uri` command.
func URISpec(container *cli.Container) *spec.Command {
	cmd := URI{}

	return &spec.Command{
		Name:  "uri",
		Short: "Print the zk:// URI of a note.",
		Group: GroupNotes,
		Args: []spec.Arg{
			{
				Name:     "path",
				Help:     "Path of the note in the notebook.",
				Target:   &cmd.Path,
				Complete: completeNotePaths(container),
			},
		},
		Flags: []spec.Flag{
			{Name: "line", Short: "l", Placeholder: "LINE", Target: &cmd.Line,
				Help: "Line to point at, starting at 1."},
			{Name: "column", Short: "c", Placeholder: "COLUMN", Target: &cmd.Column,
				Help: "Column to point at, starting at 1."},
		},
		Run: func() error { return cmd.Run(container) },
	}
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
	fmt.Fprintln(container.Stdout, uri.String())
	return nil
}
