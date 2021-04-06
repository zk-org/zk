package zk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestDBPath(t *testing.T) {
	wd, _ := os.Getwd()
	zk := &Zk{Path: wd}

	assert.Equal(t, zk.DBPath(), filepath.Join(wd, ".zk/notebook.db"))
}

func TestRootDir(t *testing.T) {
	wd, _ := os.Getwd()
	zk := &Zk{Path: wd}

	assert.Equal(t, zk.RootDir(), Dir{
		Name:   "",
		Path:   wd,
		Config: zk.Config.RootGroupConfig(),
	})
}

func TestRelativePathFromGivenPath(t *testing.T) {
	// The tests are relative to the working directory, for convenience.
	wd, _ := os.Getwd()

	zk := &Zk{Path: wd}

	for path, expected := range map[string]string{
		"log":                        "log",
		"log/sub":                    "log/sub",
		"log/sub/..":                 "log",
		"log/sub/../sub":             "log/sub",
		filepath.Join(wd, "log"):     "log",
		filepath.Join(wd, "log/sub"): "log/sub",
	} {
		actual, err := zk.RelPath(path)
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	}
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

func TestDirAtOutsideNotebook(t *testing.T) {
	wd, _ := os.Getwd()
	zk := &Zk{Path: wd}

	for _, path := range []string{
		"..",
		"../..",
		"/tmp",
	} {
		_, err := zk.DirAt(path)
		assert.Err(t, err, "path is outside the notebook")
	}
}

// When requesting the root directory `.`, the config is the default one.
func TestDirAtRoot(t *testing.T) {
	wd, _ := os.Getwd()

	zk := Zk{
		Path: wd,
		Config: Config{
			Note: NoteConfig{
				FilenameTemplate: "{{id}}.note",
				BodyTemplatePath: opt.NewString("default.note"),
				IDOptions: core.IDOptions{
					Length:  4,
					Charset: core.CharsetAlphanum,
					Case:    core.CaseLower,
				},
			},
			Groups: map[string]GroupConfig{
				"log": {
					Note: NoteConfig{
						FilenameTemplate: "{{date}}.md",
					},
				},
			},
			Extra: map[string]string{
				"hello": "world",
			},
		},
	}

	dir, err := zk.DirAt(".")
	assert.Nil(t, err)
	assert.Equal(t, dir.Name, "")
	assert.Equal(t, dir.Path, wd)
	assert.Equal(t, dir.Config, GroupConfig{
		Paths: []string{},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: core.IDOptions{
				Length:  4,
				Charset: core.CharsetAlphanum,
				Case:    core.CaseLower,
			},
		},
		Extra: map[string]string{
			"hello": "world",
		},
	})
}

// When requesting a directory, the matching GroupConfig will be returned.
func TestDirAtFindsGroup(t *testing.T) {
	wd, _ := os.Getwd()

	zk := Zk{
		Path: wd,
		Config: Config{
			Groups: map[string]GroupConfig{
				"ref": {
					Paths: []string{"ref"},
				},
				"log": {
					Paths: []string{"journal/daily", "journal/weekly"},
				},
				"glob": {
					Paths: []string{"glob/*"},
				},
			},
		},
	}

	dir, err := zk.DirAt("ref")
	assert.Nil(t, err)
	assert.Equal(t, dir.Config.Paths, []string{"ref"})

	dir, err = zk.DirAt("journal/weekly")
	assert.Nil(t, err)
	assert.Equal(t, dir.Config.Paths, []string{"journal/daily", "journal/weekly"})

	dir, err = zk.DirAt("glob/qwfpgj")
	assert.Nil(t, err)
	assert.Equal(t, dir.Config.Paths, []string{"glob/*"})

	dir, err = zk.DirAt("glob/qwfpgj/no")
	assert.Nil(t, err)
	assert.Equal(t, dir.Config.Paths, []string{})

	dir, err = zk.DirAt("glob")
	assert.Nil(t, err)
	assert.Equal(t, dir.Config.Paths, []string{})
}

// Modifying the GroupConfig of the returned Dir should not modify the global config.
func TestDirAtReturnsClonedConfig(t *testing.T) {
	wd, _ := os.Getwd()
	zk := Zk{
		Path: wd,
		Config: Config{
			Note: NoteConfig{
				FilenameTemplate: "{{id}}.note",
				BodyTemplatePath: opt.NewString("default.note"),
				IDOptions: core.IDOptions{
					Length:  4,
					Charset: core.CharsetAlphanum,
					Case:    core.CaseLower,
				},
			},
			Extra: map[string]string{
				"hello": "world",
			},
		},
	}

	dir, err := zk.DirAt(".")
	assert.Nil(t, err)

	dir.Config.Note.FilenameTemplate = "modified"
	dir.Config.Note.BodyTemplatePath = opt.NewString("modified")
	dir.Config.Note.IDOptions.Length = 41
	dir.Config.Note.IDOptions.Charset = core.CharsetNumbers
	dir.Config.Note.IDOptions.Case = core.CaseUpper
	dir.Config.Extra["test"] = "modified"

	assert.Equal(t, zk.Config.RootGroupConfig(), GroupConfig{
		Paths: []string{},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: core.IDOptions{
				Length:  4,
				Charset: core.CharsetAlphanum,
				Case:    core.CaseLower,
			},
		},
		Extra: map[string]string{
			"hello": "world",
		},
	})
}

func TestDirAtWithOverrides(t *testing.T) {
	wd, _ := os.Getwd()
	zk := Zk{
		Path: wd,
		Config: Config{
			Note: NoteConfig{
				FilenameTemplate: "{{id}}.note",
				BodyTemplatePath: opt.NewString("default.note"),
				IDOptions: core.IDOptions{
					Length:  4,
					Charset: core.CharsetLetters,
					Case:    core.CaseUpper,
				},
			},
			Extra: map[string]string{
				"hello": "world",
			},
			Groups: map[string]GroupConfig{
				"group": {
					Paths: []string{"group-path"},
					Note: NoteConfig{
						BodyTemplatePath: opt.NewString("group.note"),
					},
					Extra: map[string]string{},
				},
			},
		},
	}

	dir, err := zk.DirAt(".",
		ConfigOverrides{
			BodyTemplatePath: opt.NewString("overridden-template"),
			Extra: map[string]string{
				"hello":      "overridden",
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
	assert.Equal(t, dir.Config, GroupConfig{
		Paths: []string{},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			BodyTemplatePath: opt.NewString("overridden-template"),
			IDOptions: core.IDOptions{
				Length:  4,
				Charset: core.CharsetLetters,
				Case:    core.CaseUpper,
			},
		},
		Extra: map[string]string{
			"hello":       "overridden",
			"additional":  "value2",
			"additional2": "value3",
		},
	})

	// Overriding the group will select a different group config.
	dir, err = zk.DirAt(".", ConfigOverrides{Group: opt.NewString("group")})
	assert.Nil(t, err)
	assert.Equal(t, dir.Config, GroupConfig{
		Paths: []string{"group-path"},
		Note: NoteConfig{
			BodyTemplatePath: opt.NewString("group.note"),
		},
		Extra: map[string]string{},
	})

	// An unknown group override returns an error.
	_, err = zk.DirAt(".", ConfigOverrides{Group: opt.NewString("foobar")})
	assert.Err(t, err, "foobar: group not find in the config file")

	// Check that the original config was not modified.
	assert.Equal(t, zk.Config.RootGroupConfig(), GroupConfig{
		Paths: []string{},
		Note: NoteConfig{
			FilenameTemplate: "{{id}}.note",
			BodyTemplatePath: opt.NewString("default.note"),
			IDOptions: core.IDOptions{
				Length:  4,
				Charset: core.CharsetLetters,
				Case:    core.CaseUpper,
			},
		},
		Extra: map[string]string{
			"hello": "world",
		},
	})
}
