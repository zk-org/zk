package zk

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/paths"
	toml "github.com/pelletier/go-toml"
)

// Config holds the global user configuration.
type Config struct {
	Note    NoteConfig
	Groups  map[string]GroupConfig
	Format  FormatConfig
	Tool    ToolConfig
	Aliases map[string]string
	Extra   map[string]string
	// Base directories for the relative template paths used in NoteConfig.
	TemplatesDirs []string
}

// RootGroupConfig returns the default GroupConfig for the root directory and its descendants.
func (c Config) RootGroupConfig() GroupConfig {
	return GroupConfig{
		Paths: []string{},
		Note:  c.Note,
		Extra: c.Extra,
	}
}

// LocateTemplate returns the absolute path for the given template path, by
// looking for it in the templates directories registered in this Config.
func (c Config) LocateTemplate(path string) (string, bool) {
	if path == "" {
		return "", false
	}

	exists := func(path string) bool {
		fmt.Println("Check exists", path)
		exists, err := paths.Exists(path)
		return exists && err == nil
	}

	if filepath.IsAbs(path) {
		return path, exists(path)
	}

	for _, dir := range c.TemplatesDirs {
		if candidate := filepath.Join(dir, path); exists(candidate) {
			return candidate, true
		}
	}

	return path, false
}

// FormatConfig holds the configuration for document formats, such as Markdown.
type FormatConfig struct {
	Markdown MarkdownConfig
}

// MarkdownConfig holds the configuration for Markdown documents.
type MarkdownConfig struct {
	// Hashtags indicates whether #hashtags are supported.
	Hashtags bool `toml:"hashtags" default:"true"`
	// ColonTags indicates whether :colon:tags: are supported.
	ColonTags bool `toml:"colon-tags" default:"false"`
	// MultiwordTags indicates whether #multi-word tags# are supported.
	MultiwordTags bool `toml:"multiword-tags" default:"false"`
}

// ToolConfig holds the external tooling configuration.
type ToolConfig struct {
	Editor     opt.String
	Pager      opt.String
	FzfPreview opt.String
}

// NoteConfig holds the user configuration used when generating new notes.
type NoteConfig struct {
	// Handlebars template used when generating a new filename.
	FilenameTemplate string
	// Extension appended to the filename.
	Extension string
	// Path to the handlebars template used when generating the note content.
	BodyTemplatePath opt.String
	// Language of the note content.
	Lang string
	// Default title to use when none is provided.
	DefaultTitle string
	// Settings used when generating a random ID.
	IDOptions IDOptions
}

// GroupConfig holds the user configuration for a given group of notes.
type GroupConfig struct {
	Paths []string
	Note  NoteConfig
	Extra map[string]string
}

// ConfigOverrides holds user configuration overridden values, for example fed
// from CLI flags.
type ConfigOverrides struct {
	Group            opt.String
	BodyTemplatePath opt.String
	Extra            map[string]string
}

// Clone creates a copy of the GroupConfig receiver.
func (c GroupConfig) Clone() GroupConfig {
	clone := c

	clone.Paths = make([]string, len(c.Paths))
	copy(clone.Paths, c.Paths)

	clone.Extra = make(map[string]string)
	for k, v := range c.Extra {
		clone.Extra[k] = v
	}
	return clone
}

// Override modifies the GroupConfig receiver by updating the properties
// overridden in ConfigOverrides.
func (c *GroupConfig) Override(overrides ConfigOverrides) {
	if !overrides.BodyTemplatePath.IsNull() {
		c.Note.BodyTemplatePath = overrides.BodyTemplatePath
	}
	if overrides.Extra != nil {
		for k, v := range overrides.Extra {
			c.Extra[k] = v
		}
	}
}

// OpenConfig creates a new Config instance from its TOML representation stored
// in the given file.
func OpenConfig(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open config file at %s", path)
	}

	return ParseConfig(content, path)
}

// ParseConfig creates a new Config instance from its TOML representation.
// path is the config absolute path, from which will be derived the base path
// for templates.
func ParseConfig(content []byte, path string) (*Config, error) {
	var tomlConf tomlConfig
	err := toml.Unmarshal(content, &tomlConf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	root := GroupConfig{
		Paths: []string{},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}",
			Extension:        "md",
			BodyTemplatePath: opt.NullString,
			Lang:             "en",
			DefaultTitle:     "Untitled",
			IDOptions: IDOptions{
				Charset: CharsetAlphanum,
				Length:  4,
				Case:    CaseLower,
			},
		},
		Extra: make(map[string]string),
	}

	note := tomlConf.Note
	if note.Filename != "" {
		root.Note.FilenameTemplate = note.Filename
	}
	if note.Extension != "" {
		root.Note.Extension = note.Extension
	}
	if note.Template != "" {
		root.Note.BodyTemplatePath = opt.NewNotEmptyString(note.Template)
	}
	if note.IDLength != 0 {
		root.Note.IDOptions.Length = note.IDLength
	}
	if note.IDCharset != "" {
		root.Note.IDOptions.Charset = charsetFromString(note.IDCharset)
	}
	if note.IDCase != "" {
		root.Note.IDOptions.Case = caseFromString(note.IDCase)
	}
	if note.Lang != "" {
		root.Note.Lang = note.Lang
	}
	if note.DefaultTitle != "" {
		root.Note.DefaultTitle = note.DefaultTitle
	}
	if tomlConf.Extra != nil {
		for k, v := range tomlConf.Extra {
			root.Extra[k] = v
		}
	}

	groups := make(map[string]GroupConfig)
	for name, dirTOML := range tomlConf.Groups {
		groups[name] = root.merge(dirTOML, name)
	}

	aliases := make(map[string]string)
	if tomlConf.Aliases != nil {
		for k, v := range tomlConf.Aliases {
			aliases[k] = v
		}
	}

	return &Config{
		Note:   root.Note,
		Groups: groups,
		Format: tomlConf.Format,
		Tool: ToolConfig{
			Editor:     opt.NewNotEmptyString(tomlConf.Tool.Editor),
			Pager:      opt.NewStringWithPtr(tomlConf.Tool.Pager),
			FzfPreview: opt.NewStringWithPtr(tomlConf.Tool.FzfPreview),
		},
		Aliases:       aliases,
		Extra:         root.Extra,
		TemplatesDirs: []string{filepath.Join(filepath.Dir(path), "templates")},
	}, nil
}

func (c GroupConfig) merge(tomlConf tomlGroupConfig, name string) GroupConfig {
	res := c.Clone()

	if tomlConf.Paths != nil {
		for _, p := range tomlConf.Paths {
			res.Paths = append(res.Paths, p)
		}
	} else {
		// If no `paths` config property was given for this group, we assume
		// that its name will be used as the path.
		res.Paths = append(res.Paths, name)
	}

	note := tomlConf.Note
	if note.Filename != "" {
		res.Note.FilenameTemplate = note.Filename
	}
	if note.Extension != "" {
		res.Note.Extension = note.Extension
	}
	if note.Template != "" {
		res.Note.BodyTemplatePath = opt.NewNotEmptyString(note.Template)
	}
	if note.IDLength != 0 {
		res.Note.IDOptions.Length = note.IDLength
	}
	if note.IDCharset != "" {
		res.Note.IDOptions.Charset = charsetFromString(note.IDCharset)
	}
	if note.IDCase != "" {
		res.Note.IDOptions.Case = caseFromString(note.IDCase)
	}
	if note.Lang != "" {
		res.Note.Lang = note.Lang
	}
	if note.DefaultTitle != "" {
		res.Note.DefaultTitle = note.DefaultTitle
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
	Note    tomlNoteConfig
	Groups  map[string]tomlGroupConfig `toml:"group"`
	Format  FormatConfig
	Tool    tomlToolConfig
	Extra   map[string]string
	Aliases map[string]string `toml:"alias"`
}

type tomlNoteConfig struct {
	Filename     string
	Extension    string
	Template     string
	Lang         string `toml:"language"`
	DefaultTitle string `toml:"default-title"`
	IDCharset    string `toml:"id-charset"`
	IDLength     int    `toml:"id-length"`
	IDCase       string `toml:"id-case"`
}

type tomlGroupConfig struct {
	Paths []string
	Note  tomlNoteConfig
	Extra map[string]string
}

type tomlToolConfig struct {
	Editor     string
	Pager      *string
	FzfPreview *string `toml:"fzf-preview"`
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
