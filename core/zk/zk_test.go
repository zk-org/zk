package zk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickael-menu/zk/util/assert"
	"github.com/mickael-menu/zk/util/opt"
)

func TestDirAtGivenPath(t *testing.T) {
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

// When requesting the root directory `.`, the config is the default one.
func TestDirAtRoot(t *testing.T) {
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
			Dirs: map[string]DirConfig{
				"log": {
					FilenameTemplate: "{{date}}.md",
				},
			},
		},
	}

	dir, err := zk.DirAt(".")
	assert.Nil(t, err)
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
