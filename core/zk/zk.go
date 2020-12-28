package zk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/rand"
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
	config config
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

	config, err := parseConfig(configContent)
	if err != nil {
		return nil, wrap(err)
	}

	return &Zk{
		Path:   path,
		config: *config,
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

// FilenameTemplate returns the filename template for the notes in the given directory.
func (zk *Zk) FilenameTemplate(dir Dir) string {
	dirConfig := zk.dirConfig(dir)

	switch {
	case dirConfig != nil && dirConfig.Filename != "":
		return dirConfig.Filename
	case zk.config.Filename != "":
		return zk.config.Filename
	default:
		return "{{random-id}}"
	}
}

// Template returns the file template to use for the notes in the given directory.
func (zk *Zk) Template(dir Dir) opt.String {
	dirConfig := zk.dirConfig(dir)

	var template string
	switch {
	case dirConfig != nil && dirConfig.Template != "":
		template = dirConfig.Template
	case zk.config.Template != "":
		template = zk.config.Template
	}

	if template == "" {
		return opt.NullString
	}

	if !filepath.IsAbs(template) {
		template = filepath.Join(zk.Path, ".zk/templates", template)
	}

	return opt.NewString(template)
}

// RandIDOpts returns the options to use to generate a random ID for the given directory.
func (zk *Zk) RandIDOpts(dir Dir) rand.IDOpts {
	toCharset := func(charset string) []rune {
		switch charset {
		case "alphanum":
			return rand.AlphanumCharset
		case "hex":
			return rand.HexCharset
		case "letters":
			return rand.LettersCharset
		case "numbers":
			return rand.NumbersCharset
		default:
			return []rune(charset)
		}
	}

	toCase := func(c string) rand.Case {
		switch c {
		case "lower":
			return rand.LowerCase
		case "upper":
			return rand.UpperCase
		case "mixed":
			return rand.MixedCase
		default:
			return rand.LowerCase
		}
	}

	// Default options
	opts := rand.IDOpts{
		Charset: rand.AlphanumCharset,
		Length:  5,
		Case:    rand.LowerCase,
	}

	merge := func(more *randomIDConfig) {
		if more.Charset != "" {
			opts.Charset = toCharset(more.Charset)
		}
		if more.Length > 0 {
			opts.Length = more.Length
		}
		if more.Case != "" {
			opts.Case = toCase(more.Case)
		}
	}

	if root := zk.config.RandomID; root != nil {
		merge(root)
	}

	if dir := zk.dirConfig(dir); dir != nil && dir.RandomID != nil {
		merge(dir.RandomID)
	}

	return opts
}

// Extra returns the extra variables for the given directory.
func (zk *Zk) Extra(dir Dir) map[string]string {
	extra := make(map[string]string)

	for k, v := range zk.config.Extra {
		extra[k] = v
	}

	if dirConfig := zk.dirConfig(dir); dirConfig != nil {
		for k, v := range dirConfig.Extra {
			extra[k] = v
		}
	}

	return extra
}

// dirConfig returns the dirConfig instance for the given directory.
func (zk *Zk) dirConfig(dir Dir) *dirConfig {
	for _, dirConfig := range zk.config.Dirs {
		if dirConfig.Dir == dir.Name {
			return &dirConfig
		}
	}
	return nil
}
