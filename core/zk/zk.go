package zk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mickael-menu/zk/util/errors"
)

const defaultConfig = `editor = "nvim"
dir "log" {
	template = "log.md"
}
`

// Zk (Zettelkasten) represents an opened slip box.
type Zk struct {
	// Slip box root path.
	Path string
	// User configuration parsed from .zk/config.hsl.
	Config Config
}

// Dir represents a directory inside a slip box.
type Dir struct {
	// Name of the directory, which is the path relative to the slip box's root.
	Name string
	// Absolute path to the directory.
	Path string
}

// Open locates a slip box at the given path and parses its configuration.
func Open(path string) (*Zk, error) {
	wrap := errors.Wrapper("open failed")

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}
	path, err = locateRoot(path)
	if err != nil {
		return nil, wrap(err)
	}

	configContent, err := ioutil.ReadFile(filepath.Join(path, ".zk/config.hcl"))
	if err != nil {
		return nil, wrap(err)
	}

	config, err := ParseConfig(configContent)
	if err != nil {
		return nil, wrap(err)
	}

	return &Zk{
		Path:   path,
		Config: *config,
	}, nil
}

// Create initializes a new slip box at the given path.
func Create(path string) error {
	wrap := errors.Wrapper("init failed")

	path, err := filepath.Abs(path)
	if err != nil {
		return wrap(err)
	}

	if existingPath, err := locateRoot(path); err == nil {
		return wrap(fmt.Errorf("a slip box already exists in %v", existingPath))
	}

	// Create .zk and .zk/templates directories.
	err = os.MkdirAll(filepath.Join(path, ".zk/templates"), os.ModePerm)
	if err != nil {
		return wrap(err)
	}

	// Write default config.toml.
	f, err := os.Create(filepath.Join(path, ".zk/config.hcl"))
	if err != nil {
		return wrap(err)
	}
	_, err = f.WriteString(defaultConfig)
	if err != nil {
		return wrap(err)
	}

	return nil
}

// locate finds the root of the slip box containing the given path.
func locateRoot(path string) (string, error) {
	if !filepath.IsAbs(path) {
		panic("absolute path expected")
	}

	var locate func(string) (string, error)
	locate = func(currentPath string) (string, error) {
		if currentPath == "/" || currentPath == "." {
			return "", fmt.Errorf("no slip box found in %v or a parent directory", path)
		}
		if dotPath := filepath.Join(currentPath, ".zk"); dirExists(dotPath) {
			return currentPath, nil
		}

		return locate(filepath.Dir(currentPath))
	}

	return locate(path)
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

// DirAt creates a Dir representation of the slip box directory at the given path.
func (zk *Zk) DirAt(path string) (*Dir, error) {
	wrap := errors.Wrapperf("%v: not a valid slip box directory", path)

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}

	name, err := filepath.Rel(zk.Path, path)
	if err != nil {
		return nil, wrap(err)
	}

	return &Dir{
		Name: name,
		Path: path,
	}, nil
}
