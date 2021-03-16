package zk

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestParseDefaultConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(""), ".zk/config.toml")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
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
		},
		Groups: make(map[string]GroupConfig),
		Format: FormatConfig{
			Markdown: MarkdownConfig{
				Hashtags:      true,
				ColonTags:     false,
				MultiwordTags: false,
			},
		},
		Tool: ToolConfig{
			Editor:     opt.NullString,
			Pager:      opt.NullString,
			FzfPreview: opt.NullString,
		},
		Aliases:       make(map[string]string),
		Extra:         make(map[string]string),
		TemplatesDirs: []string{".zk/templates"},
	})
}

func TestParseInvalidConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(`;`), ".zk/config.toml")

	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestParseComplete(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		# Comment

		[note]
		filename = "{{id}}.note"
		extension = "txt"
		template = "default.note"
		language = "fr"
		default-title = "Sans titre"
		id-charset = "alphanum"
		id-length = 4
		id-case = "lower"

		[format.markdown]
		hashtags = false
		colon-tags = true
		multiword-tags = true

		[tool]
		editor = "vim"
		pager = "less"
		fzf-preview = "bat {1}"

		[extra]
		hello = "world"
		salut = "le monde"

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
		
		[group.log.extra]
		log-ext = "value"

		[group.ref.note]
		filename = "{{slug title}}.md"

		[group."without path"]
		paths = []
	`), ".zk/config.toml")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
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
				},
				Extra: map[string]string{
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
				},
				Extra: map[string]string{
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
				},
				Extra: map[string]string{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
		Format: FormatConfig{
			Markdown: MarkdownConfig{
				Hashtags:      false,
				ColonTags:     true,
				MultiwordTags: true,
			},
		},
		Tool: ToolConfig{
			Editor:     opt.NewString("vim"),
			Pager:      opt.NewString("less"),
			FzfPreview: opt.NewString("bat {1}"),
		},
		Aliases: map[string]string{
			"ls": "zk list $@",
			"ed": "zk edit $@",
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
		TemplatesDirs: []string{".zk/templates"},
	})
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
	`), ".zk/config.toml")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
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
				},
				Extra: map[string]string{
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
				},
				Extra: map[string]string{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
		Format: FormatConfig{
			Markdown: MarkdownConfig{
				Hashtags:      true,
				ColonTags:     false,
				MultiwordTags: false,
			},
		},
		Aliases: make(map[string]string),
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
		TemplatesDirs: []string{".zk/templates"},
	})
}

// Some properties like `pager` and `fzf.preview` differentiate between not
// being set and an empty string.
func TestParsePreservePropertiesAllowingEmptyValues(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		[tool]
		pager = ""
		fzf-preview = ""
	`), ".zk/config.toml")

	assert.Nil(t, err)
	assert.Equal(t, conf.Tool.Pager.IsNull(), false)
	assert.Equal(t, conf.Tool.Pager, opt.NewString(""))
	assert.Equal(t, conf.Tool.FzfPreview.IsNull(), false)
	assert.Equal(t, conf.Tool.FzfPreview, opt.NewString(""))
}

func TestParseIDCharset(t *testing.T) {
	test := func(charset string, expected Charset) {
		toml := fmt.Sprintf(`
			[note]
			id-charset = "%v"
		`, charset)
		conf, err := ParseConfig([]byte(toml), ".zk/config.toml")
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
		conf, err := ParseConfig([]byte(toml), ".zk/config.toml")
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

func TestLocateTemplate(t *testing.T) {
	root := fmt.Sprintf("/tmp/zk-test-%d", time.Now().Unix())
	os.Remove(root)
	os.MkdirAll(filepath.Join(root, "templates"), os.ModePerm)

	test := func(template string, expected string, exists bool) {
		conf, err := ParseConfig([]byte(""), filepath.Join(root, "config.toml"))
		assert.Nil(t, err)

		path, ok := conf.LocateTemplate(template)
		if exists {
			assert.True(t, ok)
			if path != expected {
				t.Errorf("Didn't resolve template `%v` as expected: %v", template, expected)
			}
		} else {
			assert.False(t, ok)
		}
	}

	tpl1 := filepath.Join(root, "templates/template.tpl")
	test("template.tpl", tpl1, false)
	os.Create(tpl1)
	test("template.tpl", tpl1, true)

	tpl2 := filepath.Join(root, "abs.tpl")
	test(tpl2, tpl2, false)
	os.Create(tpl2)
	test(tpl2, tpl2, true)
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
		},
		Extra: map[string]string{
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
		},
		Extra: map[string]string{
			"hello": "world",
		},
	})
}

func TestGroupConfigOverride(t *testing.T) {
	sut := GroupConfig{
		Paths: []string{"path"},
		Note: NoteConfig{
			FilenameTemplate: "filename",
			BodyTemplatePath: opt.NewString("body.tpl"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetLetters,
				Case:    CaseUpper,
			},
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	}

	// Empty overrides
	sut.Override(ConfigOverrides{})
	assert.Equal(t, sut, GroupConfig{
		Paths: []string{"path"},
		Note: NoteConfig{
			FilenameTemplate: "filename",
			BodyTemplatePath: opt.NewString("body.tpl"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetLetters,
				Case:    CaseUpper,
			},
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	})

	// Some overrides
	sut.Override(ConfigOverrides{
		BodyTemplatePath: opt.NewString("overridden-template"),
		Extra: map[string]string{
			"hello":      "overridden",
			"additional": "value",
		},
	})
	assert.Equal(t, sut, GroupConfig{
		Paths: []string{"path"},
		Note: NoteConfig{
			FilenameTemplate: "filename",
			BodyTemplatePath: opt.NewString("overridden-template"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetLetters,
				Case:    CaseUpper,
			},
		},
		Extra: map[string]string{
			"hello":      "overridden",
			"salut":      "le monde",
			"additional": "value",
		},
	})
}
