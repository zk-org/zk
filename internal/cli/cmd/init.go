package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/cli"
)

// Init creates a notebook in the given directory
type Init struct {
	Directory string `arg optional type:"path" default:"." help:"Directory containing the notebook."`
}

func (cmd *Init) Run(container *cli.Container) error {
	notebook, err := container.Notebooks.Init(cmd.Directory)
	if err != nil {
		return err
	}

	force := false
	_, err = notebook.Index(force)
	if err != nil {
		return err
	}

	path, err := filepath.Abs(cmd.Directory)
	if err != nil {
		path = cmd.Directory
	}

	fmt.Printf("Initialized a notebook in %v\n", path)
	return nil
}
