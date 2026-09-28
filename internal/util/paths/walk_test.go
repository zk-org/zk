package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// Walk should follow symbolic links, emitting the notes of a linked directory
// under the link's path, without looping on link cycles or emitting a note
// twice.
func TestWalkFollowsSymlinks(t *testing.T) {
	// Symlinks are created at runtime because patch files can't carry them.
	tmp := t.TempDir()
	mkfile := func(path string) {
		path = filepath.Join(tmp, path)
		assert.Nil(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.Nil(t, os.WriteFile(path, []byte("# Note\n"), 0o644))
	}
	symlink := func(target string, link string) {
		assert.Nil(t, os.Symlink(target, filepath.Join(tmp, link)))
	}

	mkfile("shared/x.md")
	mkfile("shared/sub/y.md")
	mkfile("excluded/e.md")
	mkfile("hidden/h.md")
	mkfile("nb/a.md")
	mkfile("nb/g-after.md")
	mkfile("nb/inner/i.md")
	symlink("../shared", "nb/b-shared")
	symlink("../nb", "shared/back")
	symlink(".", "shared/self")
	symlink("inner", "nb/c-inner")
	symlink("../shared/x.md", "nb/d-file.md")
	symlink("../missing", "nb/e-broken")
	symlink("../excluded", "nb/f-excluded")
	symlink("../hidden", "nb/.hidden")

	// The link's own mtime would make zk re-parse the note on every index.
	targetTime := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	assert.Nil(t, os.Chtimes(filepath.Join(tmp, "shared/x.md"), targetTime, targetTime))

	shouldIgnore := func(path string, isDir bool) (bool, error) {
		if isDir {
			return path == "f-excluded", nil
		}
		return filepath.Ext(path) != ".md", nil
	}

	path := filepath.Join(tmp, "nb")
	actual := make([]string, 0)
	modified := map[string]time.Time{}
	for m := range Walk(path, &util.NullLogger, filepath.Base(path), shouldIgnore) {
		actual = append(actual, m.Path)
		modified[m.Path] = m.Modified
	}

	assert.Equal(t, actual, []string{
		"a.md",
		"b-shared/sub/y.md",
		"b-shared/x.md",
		"d-file.md",
		"g-after.md",
		"inner/i.md",
	})
	assert.Equal(t, modified["d-file.md"], targetTime)
}
