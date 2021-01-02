package zk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/paths"
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
	// Global user configuration.
	Config Config
}

// Dir represents a directory inside a slip box.
type Dir struct {
	// Name of the directory, which is the path relative to the slip box's root.
	Name string
	// Absolute path to the directory.
	Path string
	// User configuration for this directory.
	Config DirConfig
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

	templatesDir := filepath.Join(path, ".zk/templates")
	config, err := ParseConfig(configContent, templatesDir)
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
		exists, err := paths.DirExists(filepath.Join(currentPath, ".zk"))
		switch {
		case err != nil:
			return "", err
		case exists:
			return currentPath, nil
		default:
			return locate(filepath.Dir(currentPath))
		}
	}

	return locate(path)
}

// DBPath returns the path to the slip box database.
func (zk *Zk) DBPath() string {
	return filepath.Join(zk.Path, ".zk/data.db")
}

// DirAt returns a Dir representation of the slip box directory at the given path.
func (zk *Zk) DirAt(path string, overrides ...ConfigOverrides) (*Dir, error) {
	wrap := errors.Wrapperf("%v: not a valid slip box directory", path)

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}

	name, err := filepath.Rel(zk.Path, path)
	if err != nil {
		return nil, wrap(err)
	}

	config, ok := zk.Config.Dirs[name]
	if !ok {
		// Fallback on root config.
		config = zk.Config.DirConfig
	}
	config = config.Clone()

	for _, v := range overrides {
		config.Override(v)
	}

	return &Dir{
		Name:   name,
		Path:   path,
		Config: config,
	}, nil
}

// RequiredDirAt is the same as DirAt, but checks that the directory exists
// before returning the Dir.
func (zk *Zk) RequireDirAt(path string, overrides ...ConfigOverrides) (*Dir, error) {
	dir, err := zk.DirAt(path, overrides...)
	if err != nil {
		return nil, err
	}
	exists, err := paths.Exists(dir.Path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%v: directory not found", path)
	}
	return dir, nil
}
