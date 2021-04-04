package sqlite

import (
	"testing"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/core/note"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestNoteIndexerAddWithTags(t *testing.T) {
	testNoteIndexer(t, func(tx Transaction, indexer *NoteIndexer) {
		assertSQL := func(after bool) {
			assertTagExistsOrNot(t, tx, true, "fiction")
			assertTagExistsOrNot(t, tx, after, "new-tag")
		}

		assertSQL(false)
		id, err := indexer.Add(note.Metadata{
			Path: "log/added.md",
			Tags: []string{"new-tag", "fiction"},
		})
		assert.Nil(t, err)
		assertSQL(true)
		assertTaggedOrNot(t, tx, true, id, "new-tag")
		assertTaggedOrNot(t, tx, true, id, "fiction")
	})
}

func TestNoteIndexerUpdateWithTags(t *testing.T) {
	testNoteIndexer(t, func(tx Transaction, indexer *NoteIndexer) {
		id := core.NoteId(1)

		assertSQL := func(after bool) {
			assertTaggedOrNot(t, tx, true, id, "fiction")
			assertTaggedOrNot(t, tx, after, id, "new-tag")
			assertTaggedOrNot(t, tx, after, id, "fantasy")
		}

		assertSQL(false)
		err := indexer.Update(note.Metadata{
			Path: "log/2021-01-03.md",
			Tags: []string{"new-tag", "fiction", "fantasy"},
		})
		assert.Nil(t, err)
		assertSQL(true)
	})
}

func testNoteIndexer(t *testing.T, callback func(tx Transaction, dao *NoteIndexer)) {
	testTransaction(t, func(tx Transaction) {
		logger := &util.NullLogger
		callback(tx, NewNoteIndexer(NewNoteDAO(tx, logger), NewCollectionDAO(tx, logger), logger))
	})
}

func assertTagExistsOrNot(t *testing.T, tx Transaction, shouldExist bool, tag string) {
	assertExistOrNot(t, tx, shouldExist, "SELECT id FROM collections WHERE kind = 'tag' AND name = ?", tag)
}

func assertTaggedOrNot(t *testing.T, tx Transaction, shouldBeTagged bool, noteId core.NoteId, tag string) {
	assertExistOrNot(t, tx, shouldBeTagged, "SELECT id FROM notes_collections WHERE note_id = ? AND collection_id IS (SELECT id FROM collections WHERE kind = 'tag' AND name = ?)", noteId, tag)
}
