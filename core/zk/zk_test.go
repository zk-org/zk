package zk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mickael-menu/zk/util/assert"
)

func TestDirAt(t *testing.T) {
	// The tests are relative to the working directory, for convenience.
	wd, err := os.Getwd()
	assert.Nil(t, err)

	zk := &Zk{
		Path:   wd,
		Config: Config{},
	}

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
