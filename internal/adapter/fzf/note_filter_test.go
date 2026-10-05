package fzf

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNoteFilterFzfOptions(t *testing.T) {
	test := func(name string, opts NoteFilterOpts, expected string) {
		filter := NoteFilter{opts: opts}
		assert.Equal(t, filter.fzfOptions(), expected)
	}

	test("keep the default options when the notes are not sorted",
		NoteFilterOpts{}, defaultOptions)

	test("disable fzf sorting when the notes are sorted",
		NoteFilterOpts{Sorted: true},
		defaultOptions+" --no-sort")

	test("keep the configured fzf options when the notes are sorted",
		NoteFilterOpts{
			FzfOptions: opt.NewNotEmptyString("--height 50%"),
			Sorted:     true,
		},
		"--height 50% --no-sort")
}

func TestCmdSuffixPlatform(t *testing.T) {
	if runtime.GOOS == "windows" {
		assert.Equal(t, CMD_SUFFIX, "")
	} else {
		assert.Equal(t, CMD_SUFFIX, " < /dev/tty > /dev/tty")
	}
}

func TestNoteFilterBindings_NilNewNoteDir(t *testing.T) {
	filter := &NoteFilter{
		opts: NoteFilterOpts{
			NewNoteDir: nil,
		},
	}
	bindings := filter.Bindings()
	assert.Equal(t, len(bindings), 0)
}

func TestNoteFilterBindings_DefaultBinding(t *testing.T) {
	zkBin, err := os.Executable()
	assert.Nil(t, err)

	filter := &NoteFilter{
		opts: NoteFilterOpts{
			NewNoteDir: &core.Dir{
				Name: "",
				Path: "/notebook",
			},
		},
	}

	bindings := filter.Bindings()
	assert.Equal(t, len(bindings), 1)
	assert.Equal(t, bindings[0].Keys, "Ctrl-E")
	assert.Equal(t, bindings[0].Description, "create a note with the query as title")

	expectedAction := fmt.Sprintf(`become("%s" new "/notebook" --title {q}%s)`, zkBin, CMD_SUFFIX)
	assert.Equal(t, bindings[0].Action, expectedAction)
}

func TestNoteFilterBindings_SubDir(t *testing.T) {
	zkBin, err := os.Executable()
	assert.Nil(t, err)

	filter := &NoteFilter{
		opts: NoteFilterOpts{
			NewNoteDir: &core.Dir{
				Name: "journal",
				Path: "/notebook/journal",
			},
		},
	}

	bindings := filter.Bindings()
	assert.Equal(t, len(bindings), 1)
	assert.Equal(t, bindings[0].Keys, "Ctrl-E")
	assert.Equal(t, bindings[0].Description, "create a note with the query as title in journal/")

	expectedAction := fmt.Sprintf(`become("%s" new "/notebook/journal" --title {q}%s)`, zkBin, CMD_SUFFIX)
	assert.Equal(t, bindings[0].Action, expectedAction)
}

func TestNoteFilterBindings_CustomBinding(t *testing.T) {
	filter := &NoteFilter{
		opts: NoteFilterOpts{
			NewBinding: opt.NewString("Ctrl-N"),
			NewNoteDir: &core.Dir{
				Name: "notes",
				Path: "/notebook/notes",
			},
		},
	}

	bindings := filter.Bindings()
	assert.Equal(t, len(bindings), 1)
	assert.Equal(t, bindings[0].Keys, "Ctrl-N")
	assert.Equal(t, bindings[0].Description, "create a note with the query as title in notes/")
}

func TestNoteFilterBindings_DisabledWhenEmpty(t *testing.T) {
	filter := &NoteFilter{
		opts: NoteFilterOpts{
			NewBinding: opt.NewString(""),
			NewNoteDir: &core.Dir{
				Name: "",
				Path: "/notebook",
			},
		},
	}

	bindings := filter.Bindings()
	assert.Equal(t, len(bindings), 0)
}

func TestNoteFilterBindings_DoesNotContainDevTtyOnWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		filter := &NoteFilter{
			opts: NoteFilterOpts{
				NewNoteDir: &core.Dir{
					Name: "",
					Path: "C:\\notes",
				},
			},
		}
		bindings := filter.Bindings()
		assert.Equal(t, len(bindings), 1)
		assert.True(t, !strings.Contains(bindings[0].Action, "/dev/tty"))
	}
}
