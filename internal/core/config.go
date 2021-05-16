package core

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
	toml "github.com/pelletier/go-toml"
)

// Config holds the user configuration.
type Config struct {
	Note    NoteConfig
	Groups  map[string]GroupConfig
	Format  FormatConfig
	Tool    ToolConfig
	LSP     LSPConfig
	Filters map[string]string
	Aliases map[string]string
	Extra   map[string]string
}

// NewDefaultConfig creates a new Config with the default settings.
func NewDefaultConfig() Config {
	return Config{
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
		Groups: map[string]GroupConfig{},
		Format: FormatConfig{
			Markdown: MarkdownConfig{
				Hashtags:          true,
				ColonTags:         false,
				MultiwordTags:     false,
				LinkFormat:        "markdown",
				LinkEncodePath:    true,
				LinkDropExtension: true,
			},
		},
		LSP: LSPConfig{
			Diagnostics: LSPDiagnosticConfig{
				WikiTitle: LSPDiagnosticNone,
				DeadLink:  LSPDiagnosticError,
			},
		},
		Filters: map[string]string{},
		Aliases: map[string]string{},
		Extra:   map[string]string{},
	}
}

// RootGroupConfig returns the default GroupConfig for the root directory and its descendants.
func (c Config) RootGroupConfig() GroupConfig {
	return GroupConfig{
		Paths: []string{},
		Note:  c.Note,
		Extra: c.Extra,
	}
}

// GroupConfigNamed returns the GroupConfig for the group with the given name.
// An empty name matches the root GroupConfig.
func (c Config) GroupConfigNamed(name string) (GroupConfig, error) {
	if name == "" {
		return c.RootGroupConfig(), nil
	} else {
		group, ok := c.Groups[name]
		if !ok {
			return GroupConfig{}, fmt.Errorf("no group named `%s` found in the config", name)
		}
		return group, nil
	}
}

// GroupNameForPath returns the name of the GroupConfig matching the given
// path, relative to the notebook.
func (c Config) GroupNameForPath(path string) (string, error) {
	for name, config := range c.Groups {
		for _, groupPath := range config.Paths {
			matches, err := filepath.Match(groupPath, path)
			if err != nil {
				return "", errors.Wrapf(err, "failed to match group %s to %s", name, path)
			} else if matches {
				return name, nil
			}
		}
	}

	return "", nil
}

// FormatConfig holds the configuration for document formats, such as Markdown.
type FormatConfig struct {
	Markdown MarkdownConfig
}

// MarkdownConfig holds the configuration for Markdown documents.
type MarkdownConfig struct {
	// Hashtags indicates whether #hashtags are supported.
	Hashtags bool
	// ColonTags indicates whether :colon:tags: are supported.
	ColonTags bool
	// MultiwordTags indicates whether #multi-word tags# are supported.
	MultiwordTags bool

	// Format used to generate links between notes.
	// Either "wiki", "markdown" or a custom template. Default is "markdown".
	LinkFormat string
	// Indicates whether a link's path will be percent-encoded.
	// Defaults to true for "markdown" format only, false otherwise.
	LinkEncodePath bool
	// Indicates whether a link's path file extension will be removed.
	LinkDropExtension bool
}

// ToolConfig holds the external tooling configuration.
type ToolConfig struct {
	Editor     opt.String
	Pager      opt.String
	FzfPreview opt.String
	FzfLine    opt.String
}

// LSPConfig holds the Language Server Protocol configuration.
type LSPConfig struct {
	Diagnostics LSPDiagnosticConfig
}

// LSPDiagnosticConfig holds the LSP diagnostics configuration.
type LSPDiagnosticConfig struct {
	WikiTitle LSPDiagnosticSeverity
	DeadLink  LSPDiagnosticSeverity
}

type LSPDiagnosticSeverity int

const (
	LSPDiagnosticNone    LSPDiagnosticSeverity = 0
	LSPDiagnosticError   LSPDiagnosticSeverity = 1
	LSPDiagnosticWarning LSPDiagnosticSeverity = 2
	LSPDiagnosticInfo    LSPDiagnosticSeverity = 3
	LSPDiagnosticHint    LSPDiagnosticSeverity = 4
)

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

// OpenConfig creates a new Config instance from its TOML representation stored
// in the given file.
func OpenConfig(path string, parentConfig Config, fs FileStorage) (Config, error) {
	// The local config is optional.
	exists, err := fs.FileExists(path)
	if err == nil && !exists {
		return parentConfig, nil
	}

	content, err := fs.Read(path)
	if err != nil {
		return parentConfig, errors.Wrapf(err, "failed to open config file at %s", path)
	}

	return ParseConfig(content, path, parentConfig)
}

// ParseConfig creates a new Config instance from its TOML representation.
// path is the config absolute path, from which will be derived the base path
// for templates.
//
// The parentConfig will be used to inherit default config settings.
func ParseConfig(content []byte, path string, parentConfig Config) (Config, error) {
	wrap := errors.Wrapperf("failed to read config")

	config := parentConfig

	var tomlConf tomlConfig
	err := toml.Unmarshal(content, &tomlConf)
	if err != nil {
		return config, wrap(err)
	}

	// Note
	note := tomlConf.Note
	if note.Filename != "" {
		config.Note.FilenameTemplate = note.Filename
	}
	if note.Extension != "" {
		config.Note.Extension = note.Extension
	}
	if note.Template != "" {
		config.Note.BodyTemplatePath = opt.NewNotEmptyString(note.Template)
	}
	if note.IDLength != 0 {
		config.Note.IDOptions.Length = note.IDLength
	}
	if note.IDCharset != "" {
		config.Note.IDOptions.Charset = charsetFromString(note.IDCharset)
	}
	if note.IDCase != "" {
		config.Note.IDOptions.Case = caseFromString(note.IDCase)
	}
	if note.Lang != "" {
		config.Note.Lang = note.Lang
	}
	if note.DefaultTitle != "" {
		config.Note.DefaultTitle = note.DefaultTitle
	}
	if tomlConf.Extra != nil {
		for k, v := range tomlConf.Extra {
			config.Extra[k] = v
		}
	}

	// Groups
	for name, dirTOML := range tomlConf.Groups {
		parent, ok := config.Groups[name]
		if !ok {
			parent = config.RootGroupConfig()
		}

		config.Groups[name] = parent.merge(dirTOML, name)
	}

	// Format
	markdown := tomlConf.Format.Markdown
	if markdown.Hashtags != nil {
		config.Format.Markdown.Hashtags = *markdown.Hashtags
	}
	if markdown.ColonTags != nil {
		config.Format.Markdown.ColonTags = *markdown.ColonTags
	}
	if markdown.MultiwordTags != nil {
		config.Format.Markdown.MultiwordTags = *markdown.MultiwordTags
	}
	if markdown.LinkFormat != nil && *markdown.LinkFormat == "" {
		*markdown.LinkFormat = "markdown"
	}
	if markdown.LinkFormat != nil {
		config.Format.Markdown.LinkFormat = *markdown.LinkFormat
	}
	if markdown.LinkEncodePath != nil {
		config.Format.Markdown.LinkEncodePath = *markdown.LinkEncodePath
	} else if markdown.LinkFormat != nil {
		config.Format.Markdown.LinkEncodePath = (*markdown.LinkFormat == "markdown")
	}
	if markdown.LinkDropExtension != nil {
		config.Format.Markdown.LinkDropExtension = *markdown.LinkDropExtension
	}

	// Tool
	tool := tomlConf.Tool
	if tool.Editor != nil {
		config.Tool.Editor = opt.NewNotEmptyString(*tool.Editor)
	}
	if tool.Pager != nil {
		config.Tool.Pager = opt.NewStringWithPtr(tool.Pager)
	}
	if tool.FzfPreview != nil {
		config.Tool.FzfPreview = opt.NewStringWithPtr(tool.FzfPreview)
	}
	if tool.FzfLine != nil {
		config.Tool.FzfLine = opt.NewNotEmptyString(*tool.FzfLine)
	}

	// LSP
	lspDiags := tomlConf.LSP.Diagnostics
	if lspDiags.WikiTitle != nil {
		config.LSP.Diagnostics.WikiTitle, err = lspDiagnosticSeverityFromString(*lspDiags.WikiTitle)
		if err != nil {
			return config, wrap(err)
		}
	}
	if lspDiags.DeadLink != nil {
		config.LSP.Diagnostics.DeadLink, err = lspDiagnosticSeverityFromString(*lspDiags.DeadLink)
		if err != nil {
			return config, wrap(err)
		}
	}

	// Filters
	if tomlConf.Filters != nil {
		for k, v := range tomlConf.Filters {
			config.Filters[k] = v
		}
	}

	// Aliases
	if tomlConf.Aliases != nil {
		for k, v := range tomlConf.Aliases {
			config.Aliases[k] = v
		}
	}

	return config, nil
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
	Format  tomlFormatConfig
	Tool    tomlToolConfig
	LSP     tomlLSPConfig
	Extra   map[string]string
	Filters map[string]string `toml:"filter"`
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

type tomlFormatConfig struct {
	Markdown tomlMarkdownConfig
}

type tomlMarkdownConfig struct {
	Hashtags          *bool   `toml:"hashtags"`
	ColonTags         *bool   `toml:"colon-tags"`
	MultiwordTags     *bool   `toml:"multiword-tags"`
	LinkFormat        *string `toml:"link-format"`
	LinkEncodePath    *bool   `toml:"link-encode-path"`
	LinkDropExtension *bool   `toml:"link-drop-extension"`
}

type tomlToolConfig struct {
	Editor     *string
	Pager      *string
	FzfPreview *string `toml:"fzf-preview"`
	FzfLine    *string `toml:"fzf-line"`
}

type tomlLSPConfig struct {
	Diagnostics struct {
		WikiTitle *string `toml:"wiki-title"`
		DeadLink  *string `toml:"dead-link"`
	}
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

func lspDiagnosticSeverityFromString(s string) (LSPDiagnosticSeverity, error) {
	switch s {
	case "", "none":
		return LSPDiagnosticNone, nil
	case "error":
		return LSPDiagnosticError, nil
	case "warning":
		return LSPDiagnosticWarning, nil
	case "info":
		return LSPDiagnosticInfo, nil
	case "hint":
		return LSPDiagnosticHint, nil
	default:
		return LSPDiagnosticNone, fmt.Errorf("%s: unknown LSP diagnostic severity - may be none, hint, info, warning or error", s)
	}
}
