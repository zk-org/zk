package sqlite

import (
	"testing"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

// FIXME: Missing tests

func TestNoteIndexAddWithTags(t *testing.T) {
	db, index := testNoteIndex(t)

	assertSQL := func(after bool) {
		assertTagExistsOrNot(t, db, true, "fiction")
		assertTagExistsOrNot(t, db, after, "new-tag")
	}

	assertSQL(false)
	id, err := index.Add(core.Note{
		Path: "log/added.md",
		Tags: []string{"new-tag", "fiction"},
	})
	assert.Nil(t, err)
	assertSQL(true)
	assertTaggedOrNot(t, db, true, id, "new-tag")
	assertTaggedOrNot(t, db, true, id, "fiction")
}

func TestNoteIndexUpdateWithTags(t *testing.T) {
	db, index := testNoteIndex(t)
	id := core.NoteID(1)

	assertSQL := func(after bool) {
		assertTaggedOrNot(t, db, true, id, "fiction")
		assertTaggedOrNot(t, db, after, id, "new-tag")
		assertTaggedOrNot(t, db, after, id, "fantasy")
	}

	assertSQL(false)
	err := index.Update(core.Note{
		Path: "log/2021-01-03.md",
		Tags: []string{"new-tag", "fiction", "fantasy"},
	})
	assert.Nil(t, err)
	assertSQL(true)
}

func testNoteIndex(t *testing.T) (*DB, *NoteIndex) {
	db := testDB(t)
	return db, NewNoteIndex(db, &util.NullLogger)
}

func assertTagExistsOrNot(t *testing.T, db *DB, shouldExist bool, tag string) {
	assertExistOrNot(t, db, shouldExist, "SELECT id FROM collections WHERE kind = 'tag' AND name = ?", tag)
}

func assertTaggedOrNot(t *testing.T, db *DB, shouldBeTagged bool, noteId core.NoteID, tag string) {
	assertExistOrNot(t, db, shouldBeTagged, "SELECT id FROM notes_collections WHERE note_id = ? AND collection_id IS (SELECT id FROM collections WHERE kind = 'tag' AND name = ?)", noteId, tag)
}
