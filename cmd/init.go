package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/core/zk"
)

// Init creates a slip box in the given directory
type Init struct {
	Directory string `arg optional type:"path" default:"." help:"Directory containing the slip box."`
}

func (cmd *Init) Run() error {
	err := zk.Create(cmd.Directory)
	if err == nil {
		path, err := filepath.Abs(cmd.Directory)
		if err != nil {
			path = cmd.Directory
		}

		fmt.Printf("Initialized a slip box in %v\n", path)
	}
	return err
}
