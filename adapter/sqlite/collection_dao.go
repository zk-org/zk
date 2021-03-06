package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/mickael-menu/zk/core"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
)

// CollectionDAO persists collections (e.g. tags) in the SQLite database.
type CollectionDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	createCollectionStmt   *LazyStmt
	findCollectionStmt     *LazyStmt
	findAssociationStmt    *LazyStmt
	createAssociationStmt  *LazyStmt
	removeAssociationsStmt *LazyStmt
}

// NewCollectionDAO creates a new instance of a DAO working on the given
// database transaction.
func NewCollectionDAO(tx Transaction, logger util.Logger) *CollectionDAO {
	return &CollectionDAO{
		tx:     tx,
		logger: logger,

		// Create a new collection.
		createCollectionStmt: tx.PrepareLazy(`
			INSERT INTO collections (kind, name)
			VALUES (?, ?)
		`),

		// Finds a collection's ID from its kind and name.
		findCollectionStmt: tx.PrepareLazy(`
			SELECT id FROM collections
			 WHERE kind = ? AND name = ?
		`),

		// Returns whether a note and a collection are associated.
		findAssociationStmt: tx.PrepareLazy(`
			SELECT id FROM notes_collections
			 WHERE note_id = ? AND collection_id = ?
		`),

		// Creates a new association between a note and a collection.
		createAssociationStmt: tx.PrepareLazy(`
			INSERT INTO notes_collections (note_id, collection_id)
			VALUES (?, ?)
		`),

		// Removes all associations for the given note.
		removeAssociationsStmt: tx.PrepareLazy(`
			DELETE FROM notes_collections
			 WHERE note_id = ?
		`),
	}
}

// FindOrCreate returns the ID of the collection with given kind and name.
// Creates the collection if it does not already exist.
func (d *CollectionDAO) FindOrCreate(kind string, name string) (core.CollectionId, error) {
	id, err := d.findCollection(kind, name)

	switch {
	case err != nil:
		return id, err
	case id.IsValid():
		return id, nil
	default:
		return d.create(kind, name)
	}
}

func (d *CollectionDAO) findCollection(kind string, name string) (core.CollectionId, error) {
	wrap := errors.Wrapperf("failed to get %s named %s", kind, name)

	row, err := d.findCollectionStmt.QueryRow(kind, name)
	if err != nil {
		return core.CollectionId(0), wrap(err)
	}

	var id sql.NullInt64
	err = row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return core.CollectionId(0), nil
	case err != nil:
		return core.CollectionId(0), wrap(err)
	default:
		return core.CollectionId(id.Int64), nil
	}
}

func (d *CollectionDAO) create(kind string, name string) (core.CollectionId, error) {
	wrap := errors.Wrapperf("failed to create new %s named %s", kind, name)

	res, err := d.createCollectionStmt.Exec(kind, name)
	if err != nil {
		return 0, wrap(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, wrap(err)
	}

	return core.CollectionId(id), nil
}

// Associate creates a new association between a note and a collection, if it
// does not already exist.
func (d *CollectionDAO) Associate(noteId core.NoteId, collectionId core.CollectionId) (core.NoteCollectionId, error) {
	wrap := errors.Wrapperf("failed to associate note %d to collection %d", noteId, collectionId)

	id, err := d.findAssociation(noteId, collectionId)

	switch {
	case err != nil:
		return id, wrap(err)
	case id.IsValid():
		return id, nil
	default:
		id, err = d.createAssociation(noteId, collectionId)
		return id, wrap(err)
	}
}

func (d *CollectionDAO) findAssociation(noteId core.NoteId, collectionId core.CollectionId) (core.NoteCollectionId, error) {
	if !noteId.IsValid() || !collectionId.IsValid() {
		return 0, fmt.Errorf("Note ID (%d) or collection ID (%d) not valid", noteId, collectionId)
	}

	row, err := d.findAssociationStmt.QueryRow(noteId, collectionId)
	if err != nil {
		return 0, err
	}

	var id sql.NullInt64
	err = row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		return 0, err
	default:
		return core.NoteCollectionId(id.Int64), nil
	}
}

func (d *CollectionDAO) createAssociation(noteId core.NoteId, collectionId core.CollectionId) (core.NoteCollectionId, error) {
	if !noteId.IsValid() || !collectionId.IsValid() {
		return 0, fmt.Errorf("Note ID (%d) or collection ID (%d) not valid", noteId, collectionId)
	}

	res, err := d.createAssociationStmt.Exec(noteId, collectionId)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return core.NoteCollectionId(id), nil
}

// RemoveAssociations deletes all associations with the given note.
func (d *CollectionDAO) RemoveAssociations(noteId core.NoteId) error {
	if !noteId.IsValid() {
		return fmt.Errorf("Note ID (%d) not valid", noteId)
	}

	_, err := d.removeAssociationsStmt.Exec(noteId)
	if err != nil {
		return errors.Wrapf(err, "failed to remove associations of note %d", noteId)
	}

	return nil
}
