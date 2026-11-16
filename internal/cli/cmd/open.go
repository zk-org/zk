package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zk-org/zk/internal/adapter/editor"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
)

// Open opens the file referenced by a zk:// URI with the user editor.
type Open struct {
	URI string `arg help:"zk:// URI to open, e.g. 'zk://?v=/path/to/notebook&f=note.md&l=12&c=4'."`
}

func (cmd *Open) Run(container *cli.Container) error {
	uri, err := core.ParseNoteURI(cmd.URI)
	if err != nil {
		return err
	}

	// The notebook is resolved from the URI alone, so that a link can be
	// opened from anywhere, in particular by a desktop URI handler which
	// knows nothing about notebooks.
	//
	// It is resolved by hand instead of with container.Notebooks.Open(), which
	// would parse the notebook's configuration and create its index database
	// as a side effect. Neither is needed to open a file, and doing either on
	// untrusted input is asking for trouble.
	notebookPath := container.FS.Canonical(uri.NotebookPath)
	isNotebook, err := container.FS.DirExists(filepath.Join(notebookPath, ".zk"))
	if err != nil {
		return err
	}
	if !isNotebook {
		return fmt.Errorf("%s: not a notebook, no .zk directory found", uri.NotebookPath)
	}

	// ParseNoteURI already rejected a path which could escape, so this only
	// guards against the joining itself going wrong, e.g. after a change to
	// how the URI is parsed.
	absPath := filepath.Join(notebookPath, filepath.FromSlash(uri.Path))
	if !isInsideDir(notebookPath, absPath) {
		return fmt.Errorf("%s: path is outside the notebook at %s", uri.Path, notebookPath)
	}

	exists, err := container.FS.FileExists(absPath)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s: file not found in notebook %s", uri.Path, notebookPath)
	}

	// So that `zk` commands run from within the editor target the notebook the
	// URI pointed at, and not the one we happened to be started in.
	os.Setenv("ZK_NOTEBOOK_DIR", notebookPath)

	ed, err := container.NewEditor()
	if err != nil {
		return err
	}
	return ed.OpenAt(absPath, editor.Position{Line: uri.Line, Column: uri.Column})
}

// isInsideDir returns whether path is located inside dir. Both are expected to
// be absolute and clean.
func isInsideDir(dir string, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil || filepath.IsAbs(rel) {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
