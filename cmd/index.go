package cmd

import (
	"fmt"
)

// Index indexes the content of all the notes in the notebook.
type Index struct {
	Force bool `short:"f" help:"Force indexing all the notes."`
	Quiet bool `short:"q" help:"Do not print statistics nor progress."`
}

func (cmd *Index) Help() string {
	return "You usually do not need to run `zk index` manually, as notes are indexed automatically when needed."
}

func (cmd *Index) Run(container *Container) error {
	zk, err := container.Zk()
	if err != nil {
		return err
	}

	_, stats, err := container.Database(zk, cmd.Force)
	if err != nil {
		return err
	}

	if err == nil && !cmd.Quiet {
		fmt.Println(stats)
	}

	return err
}
