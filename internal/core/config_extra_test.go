package core

import (
	"testing"

	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestParseNotebooksGlobal(t *testing.T) {
	toml := `
		notebooks = [
			"/home/user/notes",
			"/home/user/work"
		]
	`
	// Should parse notebooks if isGlobal == true
	conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), true)
	assert.Nil(t, err)
	assert.Equal(t, conf.Notebooks, []string{"/home/user/notes", "/home/user/work"})
}

func TestParseNotebookContexts(t *testing.T) {
	toml := `
		[notebook]
		dir = "/home/user/notes"
		contexts = [
			"/home/user/project1",
			"/home/user/project2"
		]
	`
	// Should parse contexts if isGlobal == true
	conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), true)
	assert.Nil(t, err)
	assert.Equal(t, conf.Notebook.Dir, opt.NewString("/home/user/notes"))
	assert.Equal(t, conf.Notebook.Contexts, []string{"/home/user/project1", "/home/user/project2"})
}

func TestParseNotebookContextsEmpty(t *testing.T) {
	toml := `
		[notebook]
		dir = "/home/user/notes"
	`
	conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), true)
	assert.Nil(t, err)
	assert.Equal(t, len(conf.Notebook.Contexts), 0)
}

func TestParseNotebooksEmpty(t *testing.T) {
	toml := `
		[notebook]
		dir = "/home/user/notes"
	`
	conf, err := ParseConfig([]byte(toml), ".zk/config.toml", NewDefaultConfig(), true)
	assert.Nil(t, err)
	assert.Equal(t, len(conf.Notebooks), 0)
}
