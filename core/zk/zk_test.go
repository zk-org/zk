package zk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/opt"
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
	assert.Equal(t, zk.FilenameTemplate(dir("")), "{{id}}")
	assert.Equal(t, zk.FilenameTemplate(dir(".")), "{{id}}")
	assert.Equal(t, zk.FilenameTemplate(dir("unknown")), "{{id}}")
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

func TestDefaultIDOptions(t *testing.T) {
	zk := &Zk{}
	defaultOpts := IDOptions{
		Charset: CharsetAlphanum,
		Length:  5,
		Case:    CaseLower,
	}

	assert.Equal(t, zk.IDOptions(dir("")), defaultOpts)
	assert.Equal(t, zk.IDOptions(dir(".")), defaultOpts)
	assert.Equal(t, zk.IDOptions(dir("unknown")), defaultOpts)
}

func TestOverrideIDOptions(t *testing.T) {
	zk := &Zk{config: config{
		ID: &idConfig{
			Charset: "alphanum",
			Length:  42,
		},
		Dirs: []dirConfig{
			{
				Dir: "log",
				ID: &idConfig{
					Length: 28,
				},
			},
		},
	}}

	expectedRootOpts := IDOptions{
		Charset: CharsetAlphanum,
		Length:  42,
		Case:    CaseLower,
	}
	assert.Equal(t, zk.IDOptions(dir("")), expectedRootOpts)
	assert.Equal(t, zk.IDOptions(dir(".")), expectedRootOpts)
	assert.Equal(t, zk.IDOptions(dir("unknown")), expectedRootOpts)

	assert.Equal(t, zk.IDOptions(dir("log")), IDOptions{
		Charset: CharsetAlphanum,
		Length:  28,
		Case:    CaseLower,
	})
}

func TestParseIDCharset(t *testing.T) {
	test := func(charset string, expected Charset) {
		zk := &Zk{config: config{
			ID: &idConfig{
				Charset: charset,
			},
		}}

		if !cmp.Equal(zk.IDOptions(dir("")).Charset, expected) {
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
		zk := &Zk{config: config{
			ID: &idConfig{
				Case: letterCase,
			},
		}}

		if !cmp.Equal(zk.IDOptions(dir("")).Case, expected) {
			t.Errorf("Didn't parse ID case `%v` as expected", letterCase)
		}
	}

	test("lower", CaseLower)
	test("upper", CaseUpper)
	test("mixed", CaseMixed)
	test("unknown", CaseLower)
}

func dir(name string) Dir {
	return Dir{Name: name, Path: name}
}
