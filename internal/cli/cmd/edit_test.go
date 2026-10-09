package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func setupTestNotebook(t *testing.T) (*cli.Container, *core.Notebook) {
	t.Helper()

	dir := t.TempDir()

	container, err := cli.NewContainer("dev")
	assert.Nil(t, err)

	notebook, err := container.Notebooks.Init(dir, core.NewDefaultInitOpts())
	assert.Nil(t, err)

	err = container.SetCurrentNotebook([]cli.Dirs{
		{
			NotebookDir: dir,
			WorkingDir:  dir,
		},
	})
	assert.Nil(t, err)

	journalDir := filepath.Join(dir, "journal")
	err = os.MkdirAll(journalDir, 0755)
	assert.Nil(t, err)

	rootNote, err := notebook.NewNote(core.NewNoteOpts{
		Title:   opt.NewString("Root Note"),
		Content: "Root content",
		Date:    time.Now(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, rootNote)

	journalNote, err := notebook.NewNote(core.NewNoteOpts{
		Directory: opt.NewString(journalDir),
		Title:     opt.NewString("Journal Note"),
		Content:   "Journal content",
		Date:      time.Now(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, journalNote)

	return container, notebook
}

func TestEditNewNoteDir_EmptyPath(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, notebook.Path)
	assert.Equal(t, dir.Name, "")
}

func TestEditNewNoteDir_ExistingDirectory(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"journal"},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, filepath.Join(notebook.Path, "journal"))
	assert.Equal(t, dir.Name, "journal")
}

func TestEditNewNoteDir_ExistingDirectoryWithSlash(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"journal/"},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, filepath.Join(notebook.Path, "journal"))
	assert.Equal(t, dir.Name, "journal")
}

func TestEditNewNoteDir_NonExistentFileInRoot(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"nonexistent.md"},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, notebook.Path)
	assert.Equal(t, dir.Name, "")
}

func TestEditNewNoteDir_NonExistentFileInSubdir(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"journal/nonexistent.md"},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, filepath.Join(notebook.Path, "journal"))
	assert.Equal(t, dir.Name, "journal")
}

func TestEditNewNoteDir_ExistingFileInRoot(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	notes, err := notebook.FindNotes(core.NoteFindOpts{})
	assert.Nil(t, err)

	var rootNotePath string
	for _, n := range notes {
		if filepath.Dir(n.Path) == "." {
			rootNotePath = n.Path
			break
		}
	}
	assert.True(t, rootNotePath != "")

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{rootNotePath},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, notebook.Path)
	assert.Equal(t, dir.Name, "")
}

func TestEditNewNoteDir_ExistingFileInSubdir(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	notes, err := notebook.FindNotes(core.NoteFindOpts{})
	assert.Nil(t, err)

	var subNotePath string
	for _, n := range notes {
		if filepath.Dir(n.Path) == "journal" {
			subNotePath = n.Path
			break
		}
	}
	assert.True(t, subNotePath != "")

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{subNotePath},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, filepath.Join(notebook.Path, "journal"))
	assert.Equal(t, dir.Name, "journal")
}

func TestEditNewNoteDir_MultiplePathsAmbiguous(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"journal", "other"},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.Nil(t, dir)
}

func TestEditNewNoteDir_NonExistentDeepPathFallsBackToRoot(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"nonexistent/deep/nested"},
		},
	}
	dir := edit.newNoteDir(notebook)

	assert.NotNil(t, dir)
	assert.Equal(t, dir.Path, notebook.Path)
	assert.Equal(t, dir.Name, "")
}

func TestEditInteractiveNoteCreationLifecycle(t *testing.T) {
	_, notebook := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"journal/missing-note.md"},
		},
	}
	targetDir := edit.newNoteDir(notebook)
	assert.NotNil(t, targetDir)
	assert.Equal(t, targetDir.Name, "journal")

	note, err := notebook.NewNote(core.NewNoteOpts{
		Directory: opt.NewString(targetDir.Path),
		Title:     opt.NewString("Created From Interactive Edit"),
		Content:   "Interactive body content",
		Date:      time.Now(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, note)

	assert.Equal(t, filepath.Dir(note.Path), "journal")
	fullPath := filepath.Join(notebook.Path, note.Path)
	content, err := os.ReadFile(fullPath)
	assert.Nil(t, err)
	assert.True(t, len(content) > 0)
}

func TestEditRun_InvalidCriteria(t *testing.T) {
	container, _ := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			MatchStrategy: "invalid-strategy",
		},
	}
	err := edit.Run(container)
	assert.NotNil(t, err)
}

func TestEditRun_NonInteractiveEmptyResult(t *testing.T) {
	container, _ := setupTestNotebook(t)

	edit := &Edit{
		Filtering: cli.Filtering{
			Path: []string{"definitely-does-not-exist.md"},
		},
	}
	err := edit.Run(container)
	assert.Nil(t, err)
}
