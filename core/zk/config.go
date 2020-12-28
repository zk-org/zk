package zk

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/rand"
)

// Config holds the user configuration of a slip box.
type Config struct {
	rootConfig rootConfig
}

type rootConfig struct {
	Filename string            `hcl:"filename,optional"`
	Template string            `hcl:"template,optional"`
	RandomID *randomIDConfig   `hcl:"random_id,block"`
	Editor   string            `hcl:"editor,optional"`
	Dirs     []dirConfig       `hcl:"dir,block"`
	Extra    map[string]string `hcl:"extra,optional"`
}

type dirConfig struct {
	Dir      string            `hcl:"dir,label"`
	Filename string            `hcl:"filename,optional"`
	Template string            `hcl:"template,optional"`
	RandomID *randomIDConfig   `hcl:"random_id,block"`
	Extra    map[string]string `hcl:"extra,optional"`
}

type randomIDConfig struct {
	Charset string `hcl:"charset,optional"`
	Length  int    `hcl:"length,optional"`
	Case    string `hcl:"case,optional"`
}

// ParseConfig creates a new Config instance from its HCL representation.
func ParseConfig(content []byte) (*Config, error) {
	var root rootConfig
	err := hclsimple.Decode(".zk/config.hcl", content, nil, &root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}
	return &Config{root}, nil
}

// Filename returns the filename template for the notes in the given directory.
func (c *Config) Filename(dir Dir) string {
	dirConfig := c.dirConfig(dir)

	switch {
	case dirConfig != nil && dirConfig.Filename != "":
		return dirConfig.Filename
	case c.rootConfig.Filename != "":
		return c.rootConfig.Filename
	default:
		return "{{random-id}}"
	}
}

// Template returns the file template to use for the notes in the given directory.
func (c *Config) Template(dir Dir) opt.String {
	dirConfig := c.dirConfig(dir)

	switch {
	case dirConfig != nil && dirConfig.Template != "":
		return opt.NewString(dirConfig.Template)
	case c.rootConfig.Template != "":
		return opt.NewString(c.rootConfig.Template)
	default:
		return opt.NullString
	}
}

// RandIDOpts returns the options to use to generate a random ID for the given directory.
func (c *Config) RandIDOpts(dir Dir) rand.IDOpts {
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

	if root := c.rootConfig.RandomID; root != nil {
		merge(root)
	}

	if dir := c.dirConfig(dir); dir != nil && dir.RandomID != nil {
		merge(dir.RandomID)
	}

	return opts
}

// Extra returns the extra variables for the given directory.
func (c *Config) Extra(dir Dir) map[string]string {
	extra := make(map[string]string)

	for k, v := range c.rootConfig.Extra {
		extra[k] = v
	}

	if dirConfig := c.dirConfig(dir); dirConfig != nil {
		for k, v := range dirConfig.Extra {
			extra[k] = v
		}
	}

	return extra
}

// dirConfig returns the dirConfig instance for the given directory.
func (c *Config) dirConfig(dir Dir) *dirConfig {
	for _, dirConfig := range c.rootConfig.Dirs {
		if dirConfig.Dir == dir.Name {
			return &dirConfig
		}
	}
	return nil
}
