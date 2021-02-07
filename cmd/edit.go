package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/adapter/fzf"
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
)

// Edit opens notes matching a set of criteria with the user editor.
type Edit struct {
	Filtering
	Sorting
	Force bool `help:"Don't confirm before editing many notes at the same time." short:"f"`
}

func (cmd *Edit) Run(container *Container) error {
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

	var notes []note.Match
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		finder := container.NoteFinder(tx, fzf.NoteFinderOpts{
			AlwaysFilter: true,
			NewNoteDir:   cmd.newNoteDir(zk),
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
		if !cmd.Force && count > 5 {
			if !container.Terminal.Confirm(
				fmt.Sprintf("Are you sure you want to open %v notes in the editor?", count),
				"Open all the notes",
				"Don't open any note",
			) {
				return nil
			}
		}
		paths := make([]string, 0)
		for _, note := range notes {
			absPath := filepath.Join(zk.Path, note.Path)
			paths = append(paths, absPath)
		}

		note.Edit(zk, paths...)

	} else {
		fmt.Println("Found 0 note")
	}

	return err
}

// newNoteDir returns the directory in which to create a new note when the fzf
// binding is triggered.
func (cmd *Edit) newNoteDir(zk *zk.Zk) *zk.Dir {
	switch len(cmd.Path) {
	case 0:
		dir := zk.RootDir()
		return &dir
	case 1:
		dir, err := zk.DirAt(cmd.Path[0])
		if err != nil {
			return nil
		}
		return dir
	default:
		// More than one directory, it's ambiguous for the "new note" fzf binding.
		return nil
	}
}
