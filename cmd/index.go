package cmd

import (
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
)

// Index indexes the content of all the notes in the slip box.
type Index struct {
	Directory string `arg optional type:"path" default:"." help:"Directory containing the notes to index"`
}

func (cmd *Index) Run(container *Container) error {
	zk, err := zk.Open(".")
	if err != nil {
		return err
	}

	dir, err := zk.RequireDirAt(cmd.Directory)
	if err != nil {
		return err
	}

	db, err := container.Database(zk)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	indexer, err := sqlite.NewNoteIndexer(tx, zk.Path, container.Logger)
	if err != nil {
		return err
	}
	err = note.Index(*dir, indexer, container.Logger)
	if err != nil {
		return err
	}

	return tx.Commit()
}
