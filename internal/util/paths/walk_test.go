package paths

import (
	"path/filepath"
	"testing"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/fixtures"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestWalk(t *testing.T) {
	var path = fixtures.Path("walk")

	shouldIgnore := func(path string) (bool, error) {
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

	shouldIgnore := func(path string) (bool, error) {
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
