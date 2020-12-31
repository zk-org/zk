package zk

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/opt"
)

func TestParseDefaultConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(""), "")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
		Editor: opt.NullString,
		DirConfig: DirConfig{
			FilenameTemplate: "{{id}}",
			BodyTemplatePath: opt.NullString,
			IDOptions: IDOptions{
				Length:  5,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			Extra: make(map[string]string),
		},
		Dirs: make(map[string]DirConfig),
	})
}

func TestParseInvalidConfig(t *testing.T) {
	conf, err := ParseConfig([]byte("unknown = 'value'"), "")

	assert.NotNil(t, err)
	assert.Nil(t, conf)
}

func TestParseComplete(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		// Comment
		editor = "vim"
		filename = "{{id}}.note"
		template = "default.note"
		id {
			charset = "alphanum"
			length = 4
			case = "lower"
		}
		extra = {
			hello = "world"
			salut = "le monde"
		}
		dir "log" {
			filename = "{{date}}.md"
			template = "log.md"
			id {
				charset = "letters"
				length = 8
				case = "mixed"
			}
			extra = {
				log-ext = "value"
			}
		}
		dir "ref" {
			filename = "{{slug title}}.md"
		}
	`), "")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
		DirConfig: DirConfig{
			FilenameTemplate: "{{id}}.note",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: IDOptions{
				Length:  4,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			},
			Extra: map[string]string{
				"hello": "world",
				"salut": "le monde",
			},
		},
		Dirs: map[string]DirConfig{
			"log": {
				FilenameTemplate: "{{date}}.md",
				BodyTemplatePath: opt.NewString("log.md"),
				IDOptions: IDOptions{
					Length:  8,
					Charset: CharsetLetters,
					Case:    CaseMixed,
				},
				Extra: map[string]string{
					"hello":   "world",
					"salut":   "le monde",
					"log-ext": "value",
				},
			},
			"ref": {
				FilenameTemplate: "{{slug title}}.md",
				BodyTemplatePath: opt.NewString("default.note"),
				IDOptions: IDOptions{
					Length:  4,
					Charset: CharsetAlphanum,
					Case:    CaseLower,
				},
				Extra: map[string]string{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
		Editor: opt.NewString("vim"),
	})
}

func TestParseMergesDirConfig(t *testing.T) {
	conf, err := ParseConfig([]byte(`
		filename = "root-filename"
		template = "root-template"
		id {
			charset = "letters"
			length = 42
			case = "upper"
		}
		extra = {
			hello = "world"
			salut = "le monde"
		}
		dir "log" {
			filename = "log-filename"
			template = "log-template"
			id {
				charset = "numbers"
				length = 8
				case = "mixed"
			}
			extra = {
				hello = "override"
				log-ext = "value"
			}
		}
		dir "inherited" {}
	`), "")

	assert.Nil(t, err)
	assert.Equal(t, conf, &Config{
		DirConfig: DirConfig{
			FilenameTemplate: "root-filename",
			BodyTemplatePath: opt.NewString("root-template"),
			IDOptions: IDOptions{
				Length:  42,
				Charset: CharsetLetters,
				Case:    CaseUpper,
			},
			Extra: map[string]string{
				"hello": "world",
				"salut": "le monde",
			},
		},
		Dirs: map[string]DirConfig{
			"log": {
				FilenameTemplate: "log-filename",
				BodyTemplatePath: opt.NewString("log-template"),
				IDOptions: IDOptions{
					Length:  8,
					Charset: CharsetNumbers,
					Case:    CaseMixed,
				},
				Extra: map[string]string{
					"hello":   "override",
					"salut":   "le monde",
					"log-ext": "value",
				},
			},
			"inherited": {
				FilenameTemplate: "root-filename",
				BodyTemplatePath: opt.NewString("root-template"),
				IDOptions: IDOptions{
					Length:  42,
					Charset: CharsetLetters,
					Case:    CaseUpper,
				},
				Extra: map[string]string{
					"hello": "world",
					"salut": "le monde",
				},
			},
		},
	})
}

func TestParseIDCharset(t *testing.T) {
	test := func(charset string, expected Charset) {
		hcl := fmt.Sprintf(`id { charset = "%v" }`, charset)
		conf, err := ParseConfig([]byte(hcl), "")
		assert.Nil(t, err)
		if !cmp.Equal(conf.IDOptions.Charset, expected) {
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
		hcl := fmt.Sprintf(`id { case = "%v" }`, letterCase)
		conf, err := ParseConfig([]byte(hcl), "")
		assert.Nil(t, err)
		if !cmp.Equal(conf.IDOptions.Case, expected) {
			t.Errorf("Didn't parse ID case `%v` as expected", letterCase)
		}
	}

	test("lower", CaseLower)
	test("upper", CaseUpper)
	test("mixed", CaseMixed)
	test("unknown", CaseLower)
}

func TestParseResolvesTemplatePaths(t *testing.T) {
	test := func(template string, expected string) {
		hcl := fmt.Sprintf(`template = "%v"`, template)
		conf, err := ParseConfig([]byte(hcl), "/test/.zk/templates")
		assert.Nil(t, err)
		if !cmp.Equal(conf.BodyTemplatePath, opt.NewString(expected)) {
			t.Errorf("Didn't resolve template `%v` as expected: %v", template, conf.BodyTemplatePath)
		}
	}

	test("template.tpl", "/test/.zk/templates/template.tpl")
	test("/abs/template.tpl", "/abs/template.tpl")
}

func TestDirConfigClone(t *testing.T) {
	original := DirConfig{
		FilenameTemplate: "{{id}}.note",
		BodyTemplatePath: opt.NewString("default.note"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetAlphanum,
			Case:    CaseLower,
		},
		Extra: map[string]string{
			"hello": "world",
		},
	}

	clone := original.Clone()
	// Check that the clone is equivalent
	assert.Equal(t, clone, original)

	clone.FilenameTemplate = "modified"
	clone.BodyTemplatePath = opt.NewString("modified")
	clone.IDOptions.Length = 41
	clone.IDOptions.Charset = CharsetNumbers
	clone.IDOptions.Case = CaseUpper
	clone.Extra["test"] = "modified"

	// Check that we didn't modify the original
	assert.Equal(t, original, DirConfig{
		FilenameTemplate: "{{id}}.note",
		BodyTemplatePath: opt.NewString("default.note"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetAlphanum,
			Case:    CaseLower,
		},
		Extra: map[string]string{
			"hello": "world",
		},
	})
}

func TestDirConfigOverride(t *testing.T) {
	sut := DirConfig{
		FilenameTemplate: "filename",
		BodyTemplatePath: opt.NewString("body.tpl"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	}

	// Empty overrides
	sut.Override(ConfigOverrides{})
	assert.Equal(t, sut, DirConfig{
		FilenameTemplate: "filename",
		BodyTemplatePath: opt.NewString("body.tpl"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	})

	// Some overrides
	sut.Override(ConfigOverrides{
		BodyTemplatePath: opt.NewString("overriden-template"),
		Extra: map[string]string{
			"hello":      "overriden",
			"additional": "value",
		},
	})
	assert.Equal(t, sut, DirConfig{
		FilenameTemplate: "filename",
		BodyTemplatePath: opt.NewString("overriden-template"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello":      "overriden",
			"salut":      "le monde",
			"additional": "value",
		},
	})
}
