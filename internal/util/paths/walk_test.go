package paths

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/fixtures"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestWalk(t *testing.T) {
	var path = fixtures.Path("walk")

	shouldIgnore := func(path string, isDir bool) (bool, error) {
		if isDir {
			return false, nil
		}
		return filepath.Ext(path) != ".md", nil
	}

	notebookRoot := filepath.Base(path)
	actual := make([]string, 0)
	for m := range Walk(path, &util.NullLogger, notebookRoot, shouldIgnore) {
		assert.NotNil(t, m.Modified)
		actual = append(actual, m.Path)
	}

	assert.Equal(t, actual, []string{
		"Dir3/a.md",
		"a.md",
		"b.md",
		"dir1/a.md",
		"dir1/b.md",
		"dir1/dir1/a.md",
		"dir1 a space/a.md",
		"dir2/a.md",
	})
}

// Walk should ignore all hidden files and dirs (prefixed with "."), with
// exception of the notebook's root dir; i.e the root dir is allowed to be
// hidden.
func TestWalkHidden(t *testing.T) {
	var path = fixtures.Path(".walk-hidden")

	shouldIgnore := func(path string, isDir bool) (bool, error) {
		if isDir {
			return false, nil
		}
		return filepath.Ext(path) != ".md", nil
	}

	notebookRoot := filepath.Base(path)
	actual := make([]string, 0)
	for m := range Walk(path, &util.NullLogger, notebookRoot, shouldIgnore) {
		assert.NotNil(t, m.Modified)
		actual = append(actual, m.Path)
	}

	assert.Equal(t, actual, []string{
		"Dir3/a.md",
		"a.md",
		"b.md",
		"dir1/a.md",
		"dir1/b.md",
		"dir1/dir1/a.md",
		"dir1 a space/a.md",
		"dir2/a.md",
	})
}

// Walk should prune directories rejected by shouldIgnorePath, so an excluded
// subtree is never traversed (rather than being filtered file by file).
func TestWalkExcludedDirsArePruned(t *testing.T) {
	var path = fixtures.Path("walk")

	// Record every path the walker asks about, to prove it never descends into
	// the excluded directory.
	queried := make([]string, 0)
	shouldIgnore := func(path string, isDir bool) (bool, error) {
		queried = append(queried, path)
		if isDir {
			return path == "dir1", nil
		}
		return filepath.Ext(path) != ".md", nil
	}

	notebookRoot := filepath.Base(path)
	actual := make([]string, 0)
	for m := range Walk(path, &util.NullLogger, notebookRoot, shouldIgnore) {
		actual = append(actual, m.Path)
	}

	// No note is emitted from the pruned dir1 subtree.
	assert.Equal(t, actual, []string{
		"Dir3/a.md",
		"a.md",
		"b.md",
		"dir1 a space/a.md",
		"dir2/a.md",
	})

	// The walker never even queried anything inside dir1 — proof it pruned the
	// directory rather than walking and filtering each file.
	for _, p := range queried {
		if strings.HasPrefix(p, "dir1"+string(filepath.Separator)) {
			t.Errorf("walker descended into pruned directory: %q", p)
		}
	}
}
