package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Init struct {
	Directory string `arg optional name:"directory" default:"."`
}

func (cmd *Init) Run() error {
	path, err := filepath.Abs(cmd.Directory)
	if err != nil {
		return err
	}

	if existingPath, err := locateZk(path); err == nil {
		return fmt.Errorf("a slip box already exists in %v", existingPath)
	}

	// Create .zk and .zk/templates directories.
	err = os.MkdirAll(filepath.Join(path, ".zk/templates"), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

var ErrNotFound = errors.New("slip box not found")

// locate finds the root of the slip box containing the given path.
func locateZk(path string) (string, error) {
	if path == "/" || path == "." {
		return "", ErrNotFound
	}
	if !filepath.IsAbs(path) {
		panic("locateZk expects an absolute path")
	}
	if dotPath := filepath.Join(path, ".zk"); dirExists(dotPath) {
		return path, nil
	}

	return locateZk(filepath.Dir(path))
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return true
	default:
		return false
	}
}
