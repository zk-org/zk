package sqlite

import (
	"testing"

	"github.com/mickael-menu/zk/internal/core/note"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestCollectionDAOFindOrCreate(t *testing.T) {
	testCollectionDAO(t, func(tx Transaction, dao *CollectionDAO) {
		// Finds existing ones
		id, err := dao.FindOrCreate("tag", "adventure")
		assert.Nil(t, err)
		assert.Equal(t, id, SQLCollectionID(2))
		id, err = dao.FindOrCreate("genre", "fiction")
		assert.Nil(t, err)
		assert.Equal(t, id, SQLCollectionID(3))

		// The name is case sensitive
		id, err = dao.FindOrCreate("tag", "Adventure")
		assert.Nil(t, err)
		assert.NotEqual(t, id, SQLCollectionID(2))

		// Creates when not found
		sql := "SELECT id FROM collections WHERE kind = ? AND name = ?"
		assertNotExist(t, tx, sql, "unknown", "created")
		_, err = dao.FindOrCreate("unknown", "created")
		assert.Nil(t, err)
		assertExist(t, tx, sql, "unknown", "created")
	})
}

func TestCollectionDaoFindAll(t *testing.T) {
	testCollectionDAO(t, func(tx Transaction, dao *CollectionDAO) {
		// Finds none
		cs, err := dao.FindAll("missing")
		assert.Nil(t, err)
		assert.Equal(t, len(cs), 0)

		// Finds existing
		cs, err = dao.FindAll("tag")
		assert.Nil(t, err)
		assert.Equal(t, cs, []note.Collection{
			{Kind: "tag", Name: "adventure", NoteCount: 2},
			{Kind: "tag", Name: "fantasy", NoteCount: 1},
			{Kind: "tag", Name: "fiction", NoteCount: 1},
			{Kind: "tag", Name: "history", NoteCount: 1},
		})
	})
}

func TestCollectionDAOAssociate(t *testing.T) {
	testCollectionDAO(t, func(tx Transaction, dao *CollectionDAO) {
		// Returns existing association
		id, err := dao.Associate(SQLNoteID(1), SQLCollectionID(2))
		assert.Nil(t, err)
		assert.Equal(t, id, SQLNoteCollectionID(2))

		// Creates a new association if missing
		noteId := SQLNoteID(5)
		collectionId := SQLCollectionID(3)
		sql := "SELECT id FROM notes_collections WHERE note_id = ? AND collection_id = ?"
		assertNotExist(t, tx, sql, noteId, collectionId)
		_, err = dao.Associate(noteId, collectionId)
		assert.Nil(t, err)
		assertExist(t, tx, sql, noteId, collectionId)
	})
}

func TestCollectionDAORemoveAssociations(t *testing.T) {
	testCollectionDAO(t, func(tx Transaction, dao *CollectionDAO) {
		noteId := SQLNoteID(1)
		sql := "SELECT id FROM notes_collections WHERE note_id = ?"
		assertExist(t, tx, sql, noteId)
		err := dao.RemoveAssociations(noteId)
		assert.Nil(t, err)
		assertNotExist(t, tx, sql, noteId)

		// Removes associations of note without any.
		err = dao.RemoveAssociations(SQLNoteID(999))
		assert.Nil(t, err)
	})
}

func testCollectionDAO(t *testing.T, callback func(tx Transaction, dao *CollectionDAO)) {
	testTransaction(t, func(tx Transaction) {
		callback(tx, NewCollectionDAO(tx, &util.NullLogger))
	})
}
