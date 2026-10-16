package paths

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestStatKeepsOnlyExistingNotes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "note.md", "content")

	indexed := []Metadata{
		{Path: "gone.md", Modified: date1},
		{Path: "note.md", Modified: date1},
	}

	notes := Collect(Stat(dir, indexed, &util.NullLogger))

	assert.Equal(t, len(notes), 1)
	assert.Equal(t, notes[0].Path, "note.md")
	// The current modification time replaces the indexed one, so the differ
	// reports the note as modified only when it really changed.
	assert.True(t, notes[0].Modified.After(date1))
	assert.Equal(t, notes[0].Modified, fileModTime(t, dir, "note.md"))
}

func TestStatKeepsAnIndexedNoteOnAReadError(t *testing.T) {
	dir := t.TempDir()
	// A note path whose parent is a file cannot be read.
	writeFile(t, dir, "parent", "not a directory")

	indexed := []Metadata{{Path: "parent/note.md", Modified: date1}}
	notes := Collect(Stat(dir, indexed, &util.NullLogger))

	assert.Equal(t, len(notes), 1)
	assert.Equal(t, notes[0].Modified, date1)
}

func TestStatSkipsAPathWhichIsNowADirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "note.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	indexed := []Metadata{{Path: "note.md", Modified: date1}}
	notes := Collect(Stat(dir, indexed, &util.NullLogger))

	assert.Equal(t, len(notes), 0)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fileModTime(t *testing.T, dir, name string) time.Time {
	t.Helper()
	info, err := os.Stat(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime().UTC()
}
