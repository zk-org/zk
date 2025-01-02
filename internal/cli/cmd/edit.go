package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/errors"
)

// Edit opens notes matching a set of criteria with the user editor.
type Edit struct {
	Force bool `short:f help:"Do not confirm before editing many notes at the same time."`
	cli.Filtering
}

func (cmd *Edit) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
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
		AlwaysFilter: true,
		NewNoteDir:   cmd.newNoteDir(notebook),
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
		if !cmd.Force && count > 5 {
			confirmed, skipped := container.Terminal.Confirm(fmt.Sprintf("Are you sure you want to open %v notes in the editor?", count), false)
			if skipped {
				return fmt.Errorf("too many notes to be opened in the editor, aborting…")
			} else if !confirmed {
				return nil
			}
		}
		paths := make([]string, 0)
		for _, note := range notes {
			absPath := filepath.Join(notebook.Path, note.Path)
			paths = append(paths, absPath)
		}

		editor, err := container.NewNoteEditor(notebook)
		if err != nil {
			return err
		}
		return editor.Open(paths...)

	} else {
		fmt.Fprintln(os.Stderr, "Found 0 notes.")
		return nil
	}
}

// newNoteDir returns the directory in which to create a new note when the fzf
// binding is triggered.
func (cmd *Edit) newNoteDir(notebook *core.Notebook) *core.Dir {
	switch len(cmd.Path) {
	case 0:
		dir := notebook.RootDir()
		return &dir
	case 1:
		dir, err := notebook.DirAt(cmd.Path[0])
		if err != nil {
			return nil
		}
		return &dir
	default:
		// More than one directory, it's ambiguous for the "new note" fzf binding.
		return nil
	}
}
