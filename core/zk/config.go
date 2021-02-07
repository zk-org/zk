package zk

import (
	"path/filepath"

	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	toml "github.com/pelletier/go-toml"
)

// Config holds the global user configuration.
type Config struct {
	DirConfig
	Dirs    map[string]DirConfig
	Editor  opt.String
	Pager   opt.String
	Aliases map[string]string
}

// DirConfig holds the user configuration for a given directory.
type DirConfig struct {
	FilenameTemplate string
	Extension        string
	BodyTemplatePath opt.String
	IDOptions        IDOptions
	Lang             string
	DefaultTitle     string
	Extra            map[string]string
}

// ConfigOverrides holds user configuration overriden values, for example fed
// from CLI flags.
type ConfigOverrides struct {
	BodyTemplatePath opt.String
	Extra            map[string]string
}

// Clone creates a copy of the DirConfig receiver.
func (c DirConfig) Clone() DirConfig {
	clone := c
	clone.Extra = make(map[string]string)
	for k, v := range c.Extra {
		clone.Extra[k] = v
	}
	return clone
}

// Override modifies the DirConfig receiver by updating the properties
// overriden in ConfigOverrides.
func (c *DirConfig) Override(overrides ConfigOverrides) {
	if !overrides.BodyTemplatePath.IsNull() {
		c.BodyTemplatePath = overrides.BodyTemplatePath
	}
	if overrides.Extra != nil {
		for k, v := range overrides.Extra {
			c.Extra[k] = v
		}
	}
}

// ParseConfig creates a new Config instance from its TOML representation.
// templatesDir is the base path for the relative templates.
func ParseConfig(content []byte, templatesDir string) (*Config, error) {
	var tomlConf tomlConfig
	err := toml.Unmarshal(content, &tomlConf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	root := DirConfig{
		FilenameTemplate: "{{id}}",
		Extension:        "md",
		BodyTemplatePath: opt.NullString,
		IDOptions: IDOptions{
			Charset: CharsetAlphanum,
			Length:  5,
			Case:    CaseLower,
		},
		Lang:         "en",
		DefaultTitle: "Untitled",
		Extra:        make(map[string]string),
	}

	if tomlConf.Filename != "" {
		root.FilenameTemplate = tomlConf.Filename
	}
	if tomlConf.Extension != "" {
		root.Extension = tomlConf.Extension
	}
	if tomlConf.Template != "" {
		root.BodyTemplatePath = templatePathFromString(tomlConf.Template, templatesDir)
	}
	if tomlConf.ID.Length != 0 {
		root.IDOptions.Length = tomlConf.ID.Length
	}
	if tomlConf.ID.Charset != "" {
		root.IDOptions.Charset = charsetFromString(tomlConf.ID.Charset)
	}
	if tomlConf.ID.Case != "" {
		root.IDOptions.Case = caseFromString(tomlConf.ID.Case)
	}
	if tomlConf.Lang != "" {
		root.Lang = tomlConf.Lang
	}
	if tomlConf.DefaultTitle != "" {
		root.DefaultTitle = tomlConf.DefaultTitle
	}
	if tomlConf.Extra != nil {
		for k, v := range tomlConf.Extra {
			root.Extra[k] = v
		}
	}

	dirs := make(map[string]DirConfig)
	for name, dirTOML := range tomlConf.Dirs {
		dirs[name] = root.merge(dirTOML, templatesDir)
	}

	aliases := make(map[string]string)
	if tomlConf.Aliases != nil {
		for k, v := range tomlConf.Aliases {
			aliases[k] = v
		}
	}

	return &Config{
		DirConfig: root,
		Dirs:      dirs,
		Editor:    opt.NewNotEmptyString(tomlConf.Editor),
		Pager:     opt.NewNotEmptyString(tomlConf.Pager),
		Aliases:   aliases,
	}, nil
}

func (c DirConfig) merge(tomlConf tomlDirConfig, templatesDir string) DirConfig {
	res := DirConfig{
		FilenameTemplate: c.FilenameTemplate,
		Extension:        c.Extension,
		BodyTemplatePath: c.BodyTemplatePath,
		IDOptions:        c.IDOptions,
		Lang:             c.Lang,
		DefaultTitle:     c.DefaultTitle,
		Extra:            make(map[string]string),
	}
	for k, v := range c.Extra {
		res.Extra[k] = v
	}

	if tomlConf.Filename != "" {
		res.FilenameTemplate = tomlConf.Filename
	}
	if tomlConf.Extension != "" {
		res.Extension = tomlConf.Extension
	}
	if tomlConf.Template != "" {
		res.BodyTemplatePath = templatePathFromString(tomlConf.Template, templatesDir)
	}
	if tomlConf.ID.Length != 0 {
		res.IDOptions.Length = tomlConf.ID.Length
	}
	if tomlConf.ID.Charset != "" {
		res.IDOptions.Charset = charsetFromString(tomlConf.ID.Charset)
	}
	if tomlConf.ID.Case != "" {
		res.IDOptions.Case = caseFromString(tomlConf.ID.Case)
	}
	if tomlConf.Lang != "" {
		res.Lang = tomlConf.Lang
	}
	if tomlConf.DefaultTitle != "" {
		res.DefaultTitle = tomlConf.DefaultTitle
	}
	if tomlConf.Extra != nil {
		for k, v := range tomlConf.Extra {
			res.Extra[k] = v
		}
	}
	return res
}

// tomlConfig holds the TOML representation of Config
type tomlConfig struct {
	Filename     string
	Extension    string
	Template     string
	ID           tomlIDConfig
	Lang         string `toml:"language"`
	DefaultTitle string `toml:"default-title"`
	Extra        map[string]string
	Dirs         map[string]tomlDirConfig `toml:"dir"`
	Editor       string
	Pager        string
	Aliases      map[string]string `toml:"alias"`
}

type tomlDirConfig struct {
	Filename     string
	Extension    string
	Template     string
	ID           tomlIDConfig
	Lang         string `toml:"language"`
	DefaultTitle string `toml:"default-title"`
	Extra        map[string]string
}

type tomlIDConfig struct {
	Charset string
	Length  int
	Case    string
}

func charsetFromString(charset string) Charset {
	switch charset {
	case "alphanum":
		return CharsetAlphanum
	case "hex":
		return CharsetHex
	case "letters":
		return CharsetLetters
	case "numbers":
		return CharsetNumbers
	default:
		return Charset(charset)
	}
}

func caseFromString(c string) Case {
	switch c {
	case "lower":
		return CaseLower
	case "upper":
		return CaseUpper
	case "mixed":
		return CaseMixed
	default:
		return CaseLower
	}
}

func templatePathFromString(template string, templatesDir string) opt.String {
	if template == "" {
		return opt.NullString
	}
	if !filepath.IsAbs(template) {
		template = filepath.Join(templatesDir, template)
	}
	return opt.NewString(template)
}
