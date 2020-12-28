package zk

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/rand"
)

func TestParseMinimal(t *testing.T) {
	config, err := ParseConfig([]byte(""))

	assert.Nil(t, err)
	assert.Equal(t, config, &Config{rootConfig{}})
}

func TestParseComplete(t *testing.T) {
	config, err := ParseConfig([]byte(`
		// Comment
		editor = "vim"
		filename = "{{random-id}}.note"
		template = "default.note"
		random_id {
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
			random_id {
				charset = "letters"
				length = 8
				case = "mixed"
			}
			extra = {
				log-ext = "value"
			}
		}
	`))

	assert.Nil(t, err)
	assert.Equal(t, config, &Config{rootConfig{
		Filename: "{{random-id}}.note",
		Template: "default.note",
		RandomID: &randomIDConfig{
			Charset: "alphanum",
			Length:  4,
			Case:    "lower",
		},
		Editor: "vim",
		Dirs: []dirConfig{
			dirConfig{
				Dir:      "log",
				Filename: "{{date}}.md",
				Template: "log.md",
				RandomID: &randomIDConfig{
					Charset: "letters",
					Length:  8,
					Case:    "mixed",
				},
				Extra: map[string]string{"log-ext": "value"},
			},
		},
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
	}})
}

func TestParseInvalidConfig(t *testing.T) {
	config, err := ParseConfig([]byte("unknown = 'value'"))

	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestDefaultFilename(t *testing.T) {
	config := &Config{}
	assert.Equal(t, config.Filename(dir("")), "{{random-id}}")
	assert.Equal(t, config.Filename(dir(".")), "{{random-id}}")
	assert.Equal(t, config.Filename(dir("unknown")), "{{random-id}}")
}

func TestCustomFilename(t *testing.T) {
	config := &Config{rootConfig{
		Filename: "root-filename",
		Dirs: []dirConfig{
			dirConfig{
				Dir:      "log",
				Filename: "log-filename",
			},
		},
	}}
	assert.Equal(t, config.Filename(dir("")), "root-filename")
	assert.Equal(t, config.Filename(dir(".")), "root-filename")
	assert.Equal(t, config.Filename(dir("unknown")), "root-filename")
	assert.Equal(t, config.Filename(dir("log")), "log-filename")
}

func TestDefaultTemplate(t *testing.T) {
	config := &Config{}
	assert.Equal(t, config.Template(dir("")), opt.NullString)
	assert.Equal(t, config.Template(dir(".")), opt.NullString)
	assert.Equal(t, config.Template(dir("unknown")), opt.NullString)
}

func TestCustomTemplate(t *testing.T) {
	config := &Config{rootConfig{
		Template: "root.tpl",
		Dirs: []dirConfig{
			dirConfig{
				Dir:      "log",
				Template: "log.tpl",
			},
		},
	}}
	assert.Equal(t, config.Template(dir("")), opt.NewString("root.tpl"))
	assert.Equal(t, config.Template(dir(".")), opt.NewString("root.tpl"))
	assert.Equal(t, config.Template(dir("unknown")), opt.NewString("root.tpl"))
	assert.Equal(t, config.Template(dir("log")), opt.NewString("log.tpl"))
}

func TestNoExtra(t *testing.T) {
	config := &Config{}
	assert.Equal(t, config.Extra(dir("")), map[string]string{})
}

func TestMergeExtra(t *testing.T) {
	config := &Config{rootConfig{
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
		Dirs: []dirConfig{
			dirConfig{
				Dir: "log",
				Extra: map[string]string{
					"hello":      "override",
					"additional": "value",
				},
			},
		},
	}}
	assert.Equal(t, config.Extra(dir("")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
	assert.Equal(t, config.Extra(dir(".")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
	assert.Equal(t, config.Extra(dir("unknown")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
	assert.Equal(t, config.Extra(dir("log")), map[string]string{
		"hello":      "override",
		"salut":      "le monde",
		"additional": "value",
	})
	// Makes sure we didn't modify the extra in place by getting the `log` ones.
	assert.Equal(t, config.Extra(dir("")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
}

func TestDefaultRandIDOpts(t *testing.T) {
	config := &Config{}
	defaultOpts := rand.IDOpts{
		Charset: rand.AlphanumCharset,
		Length:  5,
		Case:    rand.LowerCase,
	}

	assert.Equal(t, config.RandIDOpts(dir("")), defaultOpts)
	assert.Equal(t, config.RandIDOpts(dir(".")), defaultOpts)
	assert.Equal(t, config.RandIDOpts(dir("unknown")), defaultOpts)
}

func TestOverrideRandIDOpts(t *testing.T) {
	config := &Config{rootConfig{
		RandomID: &randomIDConfig{
			Charset: "alphanum",
			Length:  42,
		},
		Dirs: []dirConfig{
			dirConfig{
				Dir: "log",
				RandomID: &randomIDConfig{
					Length: 28,
				},
			},
		},
	}}

	expectedRootOpts := rand.IDOpts{
		Charset: rand.AlphanumCharset,
		Length:  42,
		Case:    rand.LowerCase,
	}
	assert.Equal(t, config.RandIDOpts(dir("")), expectedRootOpts)
	assert.Equal(t, config.RandIDOpts(dir(".")), expectedRootOpts)
	assert.Equal(t, config.RandIDOpts(dir("unknown")), expectedRootOpts)

	assert.Equal(t, config.RandIDOpts(dir("log")), rand.IDOpts{
		Charset: rand.AlphanumCharset,
		Length:  28,
		Case:    rand.LowerCase,
	})
}

func TestParseRandIDCharset(t *testing.T) {
	test := func(charset string, expected []rune) {
		config := &Config{rootConfig{
			RandomID: &randomIDConfig{
				Charset: charset,
			},
		}}

		if !cmp.Equal(config.RandIDOpts(dir("")).Charset, expected) {
			t.Errorf("Didn't parse random ID charset `%v` as expected", charset)
		}
	}

	test("alphanum", rand.AlphanumCharset)
	test("hex", rand.HexCharset)
	test("letters", rand.LettersCharset)
	test("numbers", rand.NumbersCharset)
	test("HEX", []rune("HEX")) // case sensitive
	test("custom", []rune("custom"))
}

func TestParseRandIDCase(t *testing.T) {
	test := func(letterCase string, expected rand.Case) {
		config := &Config{rootConfig{
			RandomID: &randomIDConfig{
				Case: letterCase,
			},
		}}

		if !cmp.Equal(config.RandIDOpts(dir("")).Case, expected) {
			t.Errorf("Didn't parse random ID case `%v` as expected", letterCase)
		}
	}

	test("lower", rand.LowerCase)
	test("upper", rand.UpperCase)
	test("mixed", rand.MixedCase)
	test("unknown", rand.LowerCase)
}

func dir(name string) Dir {
	return Dir{Name: name, Path: name}
}
