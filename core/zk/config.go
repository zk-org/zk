package zk

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
)

// Config holds the global user configuration.
type Config struct {
	DirConfig
	Dirs   map[string]DirConfig
	Editor opt.String
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

// ParseConfig creates a new Config instance from its HCL representation.
// templatesDir is the base path for the relative templates.
func ParseConfig(content []byte, templatesDir string) (*Config, error) {
	var hcl hclConfig
	err := hclsimple.Decode(".zk/config.hcl", content, nil, &hcl)
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

	if hcl.Filename != "" {
		root.FilenameTemplate = hcl.Filename
	}
	if hcl.Extension != "" {
		root.Extension = hcl.Extension
	}
	if hcl.Template != "" {
		root.BodyTemplatePath = templatePathFromString(hcl.Template, templatesDir)
	}
	if hcl.ID != nil {
		if hcl.ID.Length != 0 {
			root.IDOptions.Length = hcl.ID.Length
		}
		if hcl.ID.Charset != "" {
			root.IDOptions.Charset = charsetFromString(hcl.ID.Charset)
		}
		if hcl.ID.Case != "" {
			root.IDOptions.Case = caseFromString(hcl.ID.Case)
		}
	}
	if hcl.Lang != "" {
		root.Lang = hcl.Lang
	}
	if hcl.DefaultTitle != "" {
		root.DefaultTitle = hcl.DefaultTitle
	}
	if hcl.Extra != nil {
		for k, v := range hcl.Extra {
			root.Extra[k] = v
		}
	}

	config := Config{
		DirConfig: root,
		Dirs:      make(map[string]DirConfig),
		Editor:    opt.NewNotEmptyString(hcl.Editor),
	}

	for _, dirHCL := range hcl.Dirs {
		config.Dirs[dirHCL.Dir] = root.merge(dirHCL, templatesDir)
	}

	return &config, nil
}

func (c DirConfig) merge(hcl hclDirConfig, templatesDir string) DirConfig {
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

	if hcl.Filename != "" {
		res.FilenameTemplate = hcl.Filename
	}
	if hcl.Extension != "" {
		res.Extension = hcl.Extension
	}
	if hcl.Template != "" {
		res.BodyTemplatePath = templatePathFromString(hcl.Template, templatesDir)
	}
	if hcl.ID != nil {
		if hcl.ID.Length != 0 {
			res.IDOptions.Length = hcl.ID.Length
		}
		if hcl.ID.Charset != "" {
			res.IDOptions.Charset = charsetFromString(hcl.ID.Charset)
		}
		if hcl.ID.Case != "" {
			res.IDOptions.Case = caseFromString(hcl.ID.Case)
		}
	}
	if hcl.Lang != "" {
		res.Lang = hcl.Lang
	}
	if hcl.DefaultTitle != "" {
		res.DefaultTitle = hcl.DefaultTitle
	}
	if hcl.Extra != nil {
		for k, v := range hcl.Extra {
			res.Extra[k] = v
		}
	}
	return res
}

// hclConfig holds the HCL representation of Config
type hclConfig struct {
	Filename     string            `hcl:"filename,optional"`
	Extension    string            `hcl:"extension,optional"`
	Template     string            `hcl:"template,optional"`
	ID           *hclIDConfig      `hcl:"id,block"`
	Lang         string            `hcl:"language,optional"`
	DefaultTitle string            `hcl:"default-title,optional"`
	Extra        map[string]string `hcl:"extra,optional"`
	Dirs         []hclDirConfig    `hcl:"dir,block"`
	Editor       string            `hcl:"editor,optional"`
}

type hclDirConfig struct {
	Dir          string            `hcl:"dir,label"`
	Filename     string            `hcl:"filename,optional"`
	Extension    string            `hcl:"extension,optional"`
	Template     string            `hcl:"template,optional"`
	ID           *hclIDConfig      `hcl:"id,block"`
	Lang         string            `hcl:"language,optional"`
	DefaultTitle string            `hcl:"default-title,optional"`
	Extra        map[string]string `hcl:"extra,optional"`
}

type hclIDConfig struct {
	Charset string `hcl:"charset,optional"`
	Length  int    `hcl:"length,optional"`
	Case    string `hcl:"case,optional"`
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
