package core

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestParseDefaultConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(""), ".zk/config.toml", NewDefaultConfig(), true)

	assert.Nil(t, err)
	assert.Equal(t, conf, Config{
		Notebook: NotebookConfig{
			Dir: opt.NullString,
		},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}",
			Extension:        "md",
			BodyTemplatePath: opt.NullString,
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			DefaultTitle: "Untitled",
			Lang:         "en",
			Exclude:      []string{},
		},
		Groups: make(map[string]GroupConfig),
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
		Tool: ToolConfig{
			Editor:     opt.NullString,
			Shell:      opt.NullString,
			Pager:      opt.NullString,
			FzfPreview: opt.NullString,
			FzfLine:    opt.NullString,
		},
		LSP: LSPConfig{
			Diagnostics: LSPDiagnosticConfig{
				WikiTitle: LSPDiagnosticNone,
				DeadLink:  LSPDiagnosticError,
			},
		},
		Filters: make(map[string]string),
		Aliases: make(map[string]string),
		Extra:   make(map[string]any),
	})
}

func TestParseInvalidConfig(t *testing.T) {
	_, err := ParseConfig([]byte(`;`), ".zk/config.toml", NewDefaultConfig(), false)
	assert.NotNil(t, err)
}

func TestParseComplete(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		# Comment

		[notebook]
		dir = "~/notebook"

		[note]
		filename = "{{id}}.note"
		extension = "txt"
		template = "default.note"
		language = "fr"
		default-title = "Sans titre"
		id-charset = "alphanum"
		id-length = 4
		id-case = "lower"
		exclude = ["ignored", ".git"]

		[format.markdown]
		hashtags = false
		colon-tags = true
		multiword-tags = true
		link-format = "custom"
		link-encode-path = true
		link-drop-extension = false

		[tool]
		editor = "vim"
		shell = "/bin/bash"
		pager = "less"
		fzf-preview = "bat {1}"
		fzf-line = "{{title}}"
		fzf-options = "--border --height 40%"
		fzf-bind-new = "Ctrl-C"

		[extra]
		hello = "world"
		salut = "le monde"

		[filter]
		recents = "--created-after '2 weeks ago'"
		journal = "journal --sort created"

		[alias]
		ls = "zk list $@"
		ed = "zk edit $@"

		[group.log]
		paths = ["journal/daily", "journal/weekly"]

		[group.log.note]
		filename = "{{date}}.md"
		extension = "note"
		template = "log.md"
		language = "de"
		default-title = "Ohne Titel"
		id-charset = "letters"
		id-length = 8
		id-case = "mixed"
		exclude = ["new-ignored"]
		
		[group.log.extra]
		log-ext = "value"

		[group.ref.note]
		filename = "{{slug title}}.md"

		[group."without path"]
		paths = []

		[lsp.completion]
		use-additional-text-edits = true
		note-label = "notelabel"
		note-filter-text = "notefiltertext"
		note-detail = "notedetail"
		
		[lsp.diagnostics]
		wiki-title = "hint"
		dead-link = "none"
	`), ".zk/config.toml", NewDefaultConfig(), true)

	assert.Nil(t, err)
	assert.Equal(t, conf, Config{
		Notebook: NotebookConfig{
			Dir: opt.NewString("~/notebook"),
		},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			Extension:        "txt",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			Lang:         "fr",
			DefaultTitle: "Sans titre",
			Exclude:      []string{"ignored", ".git"},
		},
		Groups: map[string]GroupConfig{
			"log": {
				Paths: []string{"journal/daily", "journal/weekly"},
				Note: NoteConfig{
					FilenameTemplate: "{{date}}.md",
					Extension:        "note",
					BodyTemplatePath: opt.NewString("log.md"),
					IDOptions: IDOptions{
						Length:  8,
						Charset: CharsetLetters,
						Case:    CaseMixed,
					},
					Lang:         "de",
					DefaultTitle: "Ohne Titel",
					Exclude:      []string{"ignored", ".git", "new-ignored"},
				},
				Extra: map[string]any{
					"hello":   "world",
					"salut":   "le monde",
					"log-ext": "value",
				},
			},
			"ref": {
				Paths: []string{"ref"},
				Note: NoteConfig{
					FilenameTemplate: "{{slug title}}.md",
					Extension:        "txt",
					BodyTemplatePath: opt.NewString("default.note"),
					IDOptions: IDOptions{
						Length:  4,
						Charset: CharsetAlphanum,
						Case:    CaseLower,
					},
					Lang:         "fr",
					DefaultTitle: "Sans titre",
					Exclude:      []string{"ignored", ".git"},
				},
				Extra: map[string]any{
					"hello": "world",
					"salut": "le monde",
				},
			},
			"without path": {
				Paths: []string{},
				Note: NoteConfig{
					FilenameTemplate: "{{id}}.note",
					Extension:        "txt",
					BodyTemplatePath: opt.NewString("default.note"),
					IDOptions: IDOptions{
						Length:  4,
						Charset: CharsetAlphanum,
						Case:    CaseLower,
					},
					Lang:         "fr",
					DefaultTitle: "Sans titre",
					Exclude:      []string{"ignored", ".git"},
				},
				Extra: map[string]any{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
		Format: FormatConfig{
			Markdown: MarkdownConfig{
				Hashtags:          false,
				ColonTags:         true,
				MultiwordTags:     true,
				LinkFormat:        "custom",
				LinkEncodePath:    true,
				LinkDropExtension: false,
			},
		},
		Tool: ToolConfig{
			Editor:     opt.NewString("vim"),
			Shell:      opt.NewString("/bin/bash"),
			Pager:      opt.NewString("less"),
			FzfPreview: opt.NewString("bat {1}"),
			FzfLine:    opt.NewString("{{title}}"),
			FzfOptions: opt.NewString("--border --height 40%"),
			FzfBindNew: opt.NewString("Ctrl-C"),
		},
		LSP: LSPConfig{
			Completion: LSPCompletionConfig{
				Note: LSPCompletionTemplates{
					Label:      opt.NewString("notelabel"),
					FilterText: opt.NewString("notefiltertext"),
					Detail:     opt.NewString("notedetail"),
				},
				UseAdditionalTextEdits: opt.True,
			},
			Diagnostics: LSPDiagnosticConfig{
				WikiTitle: LSPDiagnosticHint,
				DeadLink:  LSPDiagnosticNone,
			},
		},
		Filters: map[string]string{
			"recents": "--created-after '2 weeks ago'",
			"journal": "journal --sort created",
		},
		Aliases: map[string]string{
			"ls": "zk list $@",
			"ed": "zk edit $@",
		},
		Extra: map[string]any{
			"hello": "world",
			"salut": "le monde",
		},
	})
}

func TestGroupNameForPathApplyDeepestMatch(t *testing.T) {
	config := Config{
		Groups: map[string]GroupConfig{
			"parent": {
				Paths: []string{"dir1"},
			},
			"child": {
				Paths: []string{"dir1/dir2"},
			},
			"other": {
				Paths: []string{"other"},
			},
			"star": {
				Paths: []string{"star/star*.md"},
			},
			"false positive for doublestar": {
				Paths: []string{"*/doublestar.md"},
			},
			"doublestar": {
				Paths: []string{"**/doublestar.md"},
			},
		},
	}

	name, err := config.GroupNameForPath("dir1/dir2/note.md")
	assert.Nil(t, err)
	assert.Equal(t, name, "child")

	name, err = config.GroupNameForPath("dir1/note.md")
	assert.Nil(t, err)
	assert.Equal(t, name, "parent")

	name, err = config.GroupNameForPath("other/note.md")
	assert.Nil(t, err)
	assert.Equal(t, name, "other")

	name, err = config.GroupNameForPath("star/start.md")
	assert.Nil(t, err)
	assert.Equal(t, name, "star")

	name, err = config.GroupNameForPath("star/star.md")
	assert.Nil(t, err)
	assert.Equal(t, name, "star")

	name, err = config.GroupNameForPath("double/star/doublestar.md")
	assert.Nil(t, err)
	assert.Equal(t, name, "doublestar")
}

func TestParseMergesGroupConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		[note]
		filename = "root-filename"
		extension = "txt"
		template = "root-template"
		language = "fr"
		default-title = "Sans titre"
		id-charset = "letters"
		id-length = 42
		id-case = "upper"
		exclude = ["ignored", ".git"]

		[extra]
		hello = "world"
		salut = "le monde"

		[group.log.note]
		filename = "log-filename"
		template = "log-template"
		id-charset = "numbers"
		id-length = 8
		id-case = "mixed"

		[group.log.extra]
		hello = "override"
		log-ext = "value"

		[group.inherited]
	`), ".zk/config.toml", NewDefaultConfig(), false)

	assert.Nil(t, err)
	assert.Equal(t, conf, Config{
		Note: NoteConfig{
			FilenameTemplate: "root-filename",
			Extension:        "txt",
			BodyTemplatePath: opt.NewString("root-template"),
			IDOptions: IDOptions{
				Length:  42,
				Charset: CharsetLetters,
				Case:    CaseUpper,
			},
			Lang:         "fr",
			DefaultTitle: "Sans titre",
			Exclude:      []string{"ignored", ".git"},
		},
		Groups: map[string]GroupConfig{
			"log": {
				Paths: []string{"log"},
				Note: NoteConfig{
					FilenameTemplate: "log-filename",
					Extension:        "txt",
					BodyTemplatePath: opt.NewString("log-template"),
					IDOptions: IDOptions{
						Length:  8,
						Charset: CharsetNumbers,
						Case:    CaseMixed,
					},
					Lang:         "fr",
					DefaultTitle: "Sans titre",
					Exclude:      []string{"ignored", ".git"},
				},
				Extra: map[string]any{
					"hello":   "override",
					"salut":   "le monde",
					"log-ext": "value",
				},
			},
			"inherited": {
				Paths: []string{"inherited"},
				Note: NoteConfig{
					FilenameTemplate: "root-filename",
					Extension:        "txt",
					BodyTemplatePath: opt.NewString("root-template"),
					IDOptions: IDOptions{
						Length:  42,
						Charset: CharsetLetters,
						Case:    CaseUpper,
					},
					Lang:         "fr",
					DefaultTitle: "Sans titre",
					Exclude:      []string{"ignored", ".git"},
				},
				Extra: map[string]any{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
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
			Completion: LSPCompletionConfig{
				Note: LSPCompletionTemplates{
					Label:      opt.NullString,
					FilterText: opt.NullString,
					Detail:     opt.NullString,
				},
			},
			Diagnostics: LSPDiagnosticConfig{
				WikiTitle: LSPDiagnosticNone,
				DeadLink:  LSPDiagnosticError,
			},
		},
		Filters: make(map[string]string),
		Aliases: make(map[string]string),
		Extra: map[string]any{
			"hello": "world",
			"salut": "le monde",
		},
	})
}

// Some properties like `pager` and `fzf.preview` differentiate between not
// being set and an empty string.
func TestParsePreservePropertiesAllowingEmptyValues(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		[tool]
		pager = ""
		fzf-preview = ""
	`), ".zk/config.toml", NewDefaultConfig(), false)

	assert.Nil(t, err)
	assert.Equal(t, conf.Tool.Pager.IsNull(), false)
	assert.Equal(t, conf.Tool.Pager, opt.NewString(""))
	assert.Equal(t, conf.Tool.FzfPreview.IsNull(), false)
	assert.Equal(t, conf.Tool.FzfPreview, opt.NewString(""))
}

func TestParseNotebook(t *testing.T) {
	toml := `
			[notebook]
			dir = "/home/user/folder"
		`
	// Should parse notebook if isGlobal == true
	conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), true)
	assert.Nil(t, err)
	assert.Equal(t, conf.Notebook.Dir, opt.NewString("/home/user/folder"))

	// Should not parse notebook if isGlobal == false
	conf, err = ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), false)
	assert.NotNil(t, err)
	assert.Err(t, err, "notebook.dir should not be set on local configuration")
}

func TestParseIDCharset(t *testing.T) {
	test := func(charset string, expected Charset) {
		toml := fmt.Sprintf(`
			[note]
			id-charset = "%v"
		`, charset)
		conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), false)
		assert.Nil(t, err)
		if !cmp.Equal(conf.Note.IDOptions.Charset, expected) {
			t.Errorf("Didn't parse ID charset `%v` as expected", charset)
		}
	}

	test("alphanum", CharsetAlphanum)
	test("hex", CharsetHex)
	test("letters", CharsetLetters)
	test("numbers", CharsetNumbers)
	test("HEX", []rune("HEX")) // case sensitive
	test("custom", []rune("custom"))
}

func TestParseIDCase(t *testing.T) {
	test := func(letterCase string, expected Case) {
		toml := fmt.Sprintf(`
			[note]
			id-case = "%v"
		`, letterCase)
		conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), false)
		assert.Nil(t, err)
		if !cmp.Equal(conf.Note.IDOptions.Case, expected) {
			t.Errorf("Didn't parse ID case `%v` as expected", letterCase)
		}
	}

	test("lower", CaseLower)
	test("upper", CaseUpper)
	test("mixed", CaseMixed)
	test("unknown", CaseLower)
}

// If link-encode-path is not set explicitly, it defaults to true for
// "markdown" format and false for anything else.
func TestParseMarkdownLinkEncodePath(t *testing.T) {
	test := func(format string, expected bool) {
		toml := fmt.Sprintf(`
			[format.markdown]
			link-format = "%s"
		`, format)
		conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), false)
		assert.Nil(t, err)
		assert.Equal(t, conf.Format.Markdown.LinkEncodePath, expected)
	}

	test("", true)
	test("markdown", true)
	test("wiki", false)
	test("custom", false)
}

func TestParseLSPDiagnosticsSeverity(t *testing.T) {
	test := func(value string, expected LSPDiagnosticSeverity) {
		toml := fmt.Sprintf(`
			[lsp.diagnostics]
			wiki-title = "%s"
			dead-link = "%s"
		`, value, value)
		conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), false)
		assert.Nil(t, err)
		assert.Equal(t, conf.LSP.Diagnostics.WikiTitle, expected)
		assert.Equal(t, conf.LSP.Diagnostics.DeadLink, expected)
	}

	test("", LSPDiagnosticNone)
	test("none", LSPDiagnosticNone)
	test("error", LSPDiagnosticError)
	test("warning", LSPDiagnosticWarning)
	test("info", LSPDiagnosticInfo)
	test("hint", LSPDiagnosticHint)

	toml := `
		[lsp.diagnostics]
		wiki-title = "foobar"
	`
	_, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), false)
	assert.Err(t, err, "foobar: unknown LSP diagnostic severity - may be none, hint, info, warning or error")
}

func TestGroupConfigExcludeGlobs(t *testing.T) {
	// empty globs
	config := GroupConfig{
		Paths: []string{"path"},
		Note:  NoteConfig{Exclude: []string{}},
	}
	assert.Equal(t, config.ExcludeGlobs(), []string{})

	// empty paths
	config = GroupConfig{
		Paths: []string{},
		Note: NoteConfig{
			Exclude: []string{"ignored", ".git"},
		},
	}
	assert.Equal(t, config.ExcludeGlobs(), []string{"ignored", ".git"})

	// several paths
	config = GroupConfig{
		Paths: []string{"log", "drafts"},
		Note: NoteConfig{
			Exclude: []string{"ignored", "*.git"},
		},
	}
	assert.Equal(t, config.ExcludeGlobs(), []string{"log/ignored", "log/*.git", "drafts/ignored", "drafts/*.git"})
}

func TestGroupConfigClone(t *testing.T) {
	original := GroupConfig{
		Paths: []string{"original"},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			Extension:        "md",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			Lang:         "fr",
			DefaultTitle: "Sans titre",
			Exclude:      []string{"ignored", ".git"},
		},
		Extra: map[string]any{
			"hello": "world",
		},
	}

	clone := original.Clone()
	// Check that the clone is equivalent
	assert.Equal(t, clone, original)

	clone.Paths = []string{"cloned"}
	clone.Note.FilenameTemplate = "modified"
	clone.Note.Extension = "txt"
	clone.Note.BodyTemplatePath = opt.NewString("modified")
	clone.Note.IDOptions.Length = 41
	clone.Note.IDOptions.Charset = CharsetNumbers
	clone.Note.IDOptions.Case = CaseUpper
	clone.Note.Lang = "de"
	clone.Note.DefaultTitle = "Ohne Titel"
	clone.Note.Exclude = []string{"other-ignored"}
	clone.Extra["test"] = "modified"

	// Check that we didn't modify the original
	assert.Equal(t, original, GroupConfig{
		Paths: []string{"original"},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			Extension:        "md",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			Lang:         "fr",
			DefaultTitle: "Sans titre",
			Exclude:      []string{"ignored", ".git"},
		},
		Extra: map[string]any{
			"hello": "world",
		},
	})
}
