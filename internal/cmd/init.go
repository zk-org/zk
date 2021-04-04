package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/core/zk"
)

// Init creates a notebook in the given directory
type Init struct {
	Directory string `arg optional type:"path" default:"." help:"Directory containing the notebook."`
}

func (cmd *Init) Run() error {
	err := zk.Create(cmd.Directory)
	if err == nil {
		path, err := filepath.Abs(cmd.Directory)
		if err != nil {
			path = cmd.Directory
		}

		fmt.Printf("Initialized a notebook in %v\n", path)
	}
	return err
}
