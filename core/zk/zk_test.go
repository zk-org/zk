package zk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickael-menu/zk/util/test/assert"
	"github.com/mickael-menu/zk/util/opt"
)

func TestDBPath(t *testing.T) {
	wd, _ := os.Getwd()
	zk := &Zk{Path: wd}

	assert.Equal(t, zk.DBPath(), filepath.Join(wd, ".zk/data.db"))
}

func TestDirAtGivenPath(t *testing.T) {
	// The tests are relative to the working directory, for convenience.
	wd, _ := os.Getwd()

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
		assert.Equal(t, actual.Name, name)
		assert.Equal(t, actual.Path, filepath.Join(wd, name))
	}
}

// When requesting the root directory `.`, the config is the default one.
func TestDirAtRoot(t *testing.T) {
	wd, _ := os.Getwd()

	zk := Zk{
		Path: wd,
		Config: Config{
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
				},
			},
			Dirs: map[string]DirConfig{
				"log": {
					FilenameTemplate: "{{date}}.md",
				},
			},
		},
	}

	dir, err := zk.DirAt(".")
	assert.Nil(t, err)
	assert.Equal(t, dir.Name, "")
	assert.Equal(t, dir.Path, wd)
	assert.Equal(t, dir.Config, DirConfig{
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

// Modifying the DirConfig of the returned Dir should not modify the global config.
func TestDirAtReturnsClonedConfig(t *testing.T) {
	zk := Zk{
		Path: "/test",
		Config: Config{
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
				},
			},
		},
	}

	dir, err := zk.DirAt(".")
	assert.Nil(t, err)

	dir.Config.FilenameTemplate = "modified"
	dir.Config.BodyTemplatePath = opt.NewString("modified")
	dir.Config.IDOptions.Length = 41
	dir.Config.IDOptions.Charset = CharsetNumbers
	dir.Config.IDOptions.Case = CaseUpper
	dir.Config.Extra["test"] = "modified"

	assert.Equal(t, zk.Config.DirConfig, DirConfig{
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

func TestDirAtWithOverrides(t *testing.T) {
	zk := Zk{
		Path: "/test",
		Config: Config{
			DirConfig: DirConfig{
				FilenameTemplate: "{{id}}.note",
				BodyTemplatePath: opt.NewString("default.note"),
				IDOptions: IDOptions{
					Length:  4,
					Charset: CharsetLetters,
					Case:    CaseUpper,
				},
				Extra: map[string]string{
					"hello": "world",
				},
			},
		},
	}

	dir, err := zk.DirAt(".",
		ConfigOverrides{
			BodyTemplatePath: opt.NewString("overriden-template"),
			Extra: map[string]string{
				"hello":      "overriden",
				"additional": "value",
			},
		},
		ConfigOverrides{
			Extra: map[string]string{
				"additional":  "value2",
				"additional2": "value3",
			},
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, dir.Config, DirConfig{
		FilenameTemplate: "{{id}}.note",
		BodyTemplatePath: opt.NewString("overriden-template"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello":       "overriden",
			"additional":  "value2",
			"additional2": "value3",
		},
	})

	// Check that the original config was not modified.
	assert.Equal(t, zk.Config.DirConfig, DirConfig{
		FilenameTemplate: "{{id}}.note",
		BodyTemplatePath: opt.NewString("default.note"),
		IDOptions: IDOptions{
			Length:  4,
			Charset: CharsetLetters,
			Case:    CaseUpper,
		},
		Extra: map[string]string{
			"hello": "world",
		},
	})
}
