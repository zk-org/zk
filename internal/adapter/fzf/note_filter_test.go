package fzf

import (
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
		NoteFilterOpts{Sorters: []core.NoteSorter{{Field: core.NoteSortPath, Ascending: true}}},
		defaultOptions+" --no-sort")

	test("keep the configured fzf options when the notes are sorted",
		NoteFilterOpts{
			FzfOptions: opt.NewNotEmptyString("--height 50%"),
			Sorters:    []core.NoteSorter{{Field: core.NoteSortPath, Ascending: true}},
		},
		"--height 50% --no-sort")
}
