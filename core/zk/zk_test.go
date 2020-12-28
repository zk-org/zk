package zk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/rand"
)

func TestDirAt(t *testing.T) {
	// The tests are relative to the working directory, for convenience.
	wd, err := os.Getwd()
	assert.Nil(t, err)

	zk := &Zk{Path: wd}

	for path, name := range map[string]string{
		"log":                        "log",
		"log/sub":                    "log/sub",
		"log/sub/..":                 "log",
		"log/sub/../sub":             "log/sub",
		filepath.Join(wd, "log"):     "log",
		filepath.Join(wd, "log/sub"): "log/sub",
	} {
		actual, err := zk.DirAt(path)
		assert.Nil(t, err)
		assert.Equal(t, actual, &Dir{Name: name, Path: filepath.Join(wd, name)})
	}
}

func TestDefaultFilenameTemplate(t *testing.T) {
	zk := &Zk{}
	assert.Equal(t, zk.FilenameTemplate(dir("")), "{{random-id}}")
	assert.Equal(t, zk.FilenameTemplate(dir(".")), "{{random-id}}")
	assert.Equal(t, zk.FilenameTemplate(dir("unknown")), "{{random-id}}")
}

func TestCustomFilenameTemplate(t *testing.T) {
	zk := &Zk{config: config{
		Filename: "root-filename",
		Dirs: []dirConfig{
			{
				Dir:      "log",
				Filename: "log-filename",
			},
		},
	}}
	assert.Equal(t, zk.FilenameTemplate(dir("")), "root-filename")
	assert.Equal(t, zk.FilenameTemplate(dir(".")), "root-filename")
	assert.Equal(t, zk.FilenameTemplate(dir("unknown")), "root-filename")
	assert.Equal(t, zk.FilenameTemplate(dir("log")), "log-filename")
}

func TestDefaultTemplate(t *testing.T) {
	zk := &Zk{}
	assert.Equal(t, zk.Template(dir("")), opt.NullString)
	assert.Equal(t, zk.Template(dir(".")), opt.NullString)
	assert.Equal(t, zk.Template(dir("unknown")), opt.NullString)
}

func TestCustomTemplate(t *testing.T) {
	zk := &Zk{
		Path: "/test",
		config: config{
			Template: "root.tpl",
			Dirs: []dirConfig{
				{
					Dir:      "log",
					Template: "log.tpl",
				},
				{
					Dir:      "abs",
					Template: "/abs/template.tpl",
				},
			},
		},
	}
	assert.Equal(t, zk.Template(dir("")), opt.NewString("/test/.zk/templates/root.tpl"))
	assert.Equal(t, zk.Template(dir(".")), opt.NewString("/test/.zk/templates/root.tpl"))
	assert.Equal(t, zk.Template(dir("unknown")), opt.NewString("/test/.zk/templates/root.tpl"))
	assert.Equal(t, zk.Template(dir("log")), opt.NewString("/test/.zk/templates/log.tpl"))
	assert.Equal(t, zk.Template(dir("abs")), opt.NewString("/abs/template.tpl"))
}

func TestNoExtra(t *testing.T) {
	zk := &Zk{}
	assert.Equal(t, zk.Extra(dir("")), map[string]string{})
}

func TestMergeExtra(t *testing.T) {
	zk := &Zk{config: config{
		Extra: map[string]string{
			"hello": "world",
			"salut": "le monde",
		},
		Dirs: []dirConfig{
			{
				Dir: "log",
				Extra: map[string]string{
					"hello":      "override",
					"additional": "value",
				},
			},
		},
	}}
	assert.Equal(t, zk.Extra(dir("")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
	assert.Equal(t, zk.Extra(dir(".")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
	assert.Equal(t, zk.Extra(dir("unknown")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
	assert.Equal(t, zk.Extra(dir("log")), map[string]string{
		"hello":      "override",
		"salut":      "le monde",
		"additional": "value",
	})
	// Makes sure we didn't modify the extra in place by getting the `log` ones.
	assert.Equal(t, zk.Extra(dir("")), map[string]string{
		"hello": "world",
		"salut": "le monde",
	})
}

func TestDefaultRandIDOpts(t *testing.T) {
	zk := &Zk{}
	defaultOpts := rand.IDOpts{
		Charset: rand.AlphanumCharset,
		Length:  5,
		Case:    rand.LowerCase,
	}

	assert.Equal(t, zk.RandIDOpts(dir("")), defaultOpts)
	assert.Equal(t, zk.RandIDOpts(dir(".")), defaultOpts)
	assert.Equal(t, zk.RandIDOpts(dir("unknown")), defaultOpts)
}

func TestOverrideRandIDOpts(t *testing.T) {
	zk := &Zk{config: config{
		RandomID: &randomIDConfig{
			Charset: "alphanum",
			Length:  42,
		},
		Dirs: []dirConfig{
			{
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
	assert.Equal(t, zk.RandIDOpts(dir("")), expectedRootOpts)
	assert.Equal(t, zk.RandIDOpts(dir(".")), expectedRootOpts)
	assert.Equal(t, zk.RandIDOpts(dir("unknown")), expectedRootOpts)

	assert.Equal(t, zk.RandIDOpts(dir("log")), rand.IDOpts{
		Charset: rand.AlphanumCharset,
		Length:  28,
		Case:    rand.LowerCase,
	})
}

func TestParseRandIDCharset(t *testing.T) {
	test := func(charset string, expected []rune) {
		zk := &Zk{config: config{
			RandomID: &randomIDConfig{
				Charset: charset,
			},
		}}

		if !cmp.Equal(zk.RandIDOpts(dir("")).Charset, expected) {
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
		zk := &Zk{config: config{
			RandomID: &randomIDConfig{
				Case: letterCase,
			},
		}}

		if !cmp.Equal(zk.RandIDOpts(dir("")).Case, expected) {
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
