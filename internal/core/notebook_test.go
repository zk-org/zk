package core

import (
	"testing"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/test/assert"
)

// TestRelPathResolvesInsideNotebook tests that Notebook.RelPath() resolves
// relative paths relative to the notebook directory.
func TestRelPathResolvesInsideNotebook(t *testing.T) {
	notebookPath := "/notebook"
	workingDir := "/notebook"

	fs := newFileStorageMock(workingDir, []string{notebookPath})

	notebook := NewNotebook(notebookPath, NewDefaultConfig(), NotebookPorts{
		FS:     fs,
		Logger: &util.NullLogger,
	})

	relPath, err := notebook.RelPath("inbox/note.md")

	assert.Nil(t, err)
	assert.Equal(t, relPath, "inbox/note.md")
}

// TestRelPathResolvesOutsideNotebook tests that Notebook.RelPath() resolves
// relative paths relative to the notebook directory, not the current working
// directory.
func TestRelPathResolvesOutsideNotebook(t *testing.T) {
	notebookPath := "/notebook"
	workingDir := "/other/dir"

	fs := newFileStorageMock(workingDir, []string{notebookPath})

	notebook := NewNotebook(notebookPath, NewDefaultConfig(), NotebookPorts{
		FS:     fs,
		Logger: &util.NullLogger,
	})

	relPath, err := notebook.RelPath("inbox/note.md")

	assert.Nil(t, err)
	assert.Equal(t, relPath, "inbox/note.md")
}

// TestRelPathResolvesInsideNotebookSubdir tests that Notebook.RelPath() resolves
// relative paths relative to the current working directory when it is a child
// of the notebook root.
func TestRelPathResolvesInsideNotebookSubdir(t *testing.T) {
	notebookPath := "/notebook"
	workingDir := "/notebook/subdir"

	fs := newFileStorageMock(workingDir, []string{notebookPath, workingDir})

	notebook := NewNotebook(notebookPath, NewDefaultConfig(), NotebookPorts{
		FS:     fs,
		Logger: &util.NullLogger,
	})

	// Relative path from /notebook/subdir should resolve to /notebook/subdir/note.md
	relPath, err := notebook.RelPath("note.md")

	assert.Nil(t, err)
	assert.Equal(t, relPath, "subdir/note.md")
}
