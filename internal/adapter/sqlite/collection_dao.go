package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
)

// CollectionDAO persists collections (e.g. tags) in the SQLite database.
type CollectionDAO struct {
	tx     Transaction
	logger util.Logger

	// Prepared SQL statements
	createCollectionStmt   *LazyStmt
	findCollectionStmt     *LazyStmt
	findAllCollectionsStmt *LazyStmt
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

		// Find all collections.
		findAllCollectionsStmt: tx.PrepareLazy(`
			SELECT c.name, COUNT(nc.id) as count
			  FROM collections c
			 INNER JOIN notes_collections nc ON nc.collection_id = c.id
			 WHERE kind = ?
			 GROUP BY c.id
			 ORDER BY c.name
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
func (d *CollectionDAO) FindOrCreate(kind core.CollectionKind, name string) (core.CollectionID, error) {
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

func (d *CollectionDAO) FindAll(kind core.CollectionKind) ([]core.Collection, error) {
	rows, err := d.findAllCollectionsStmt.Query(kind)
	if err != nil {
		return []core.Collection{}, err
	}
	defer rows.Close()

	collections := []core.Collection{}

	for rows.Next() {
		var name string
		var count int
		err := rows.Scan(&name, &count)
		if err != nil {
			return collections, err
		}

		collections = append(collections, core.Collection{
			Kind:      kind,
			Name:      name,
			NoteCount: count,
		})
	}

	return collections, nil
}

func (d *CollectionDAO) findCollection(kind core.CollectionKind, name string) (core.CollectionID, error) {
	wrap := errors.Wrapperf("failed to get %s named %s", kind, name)

	row, err := d.findCollectionStmt.QueryRow(kind, name)
	if err != nil {
		return 0, wrap(err)
	}

	var id sql.NullInt64
	err = row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		return 0, nil
	case err != nil:
		return 0, wrap(err)
	default:
		return core.CollectionID(id.Int64), nil
	}
}

func (d *CollectionDAO) create(kind core.CollectionKind, name string) (core.CollectionID, error) {
	wrap := errors.Wrapperf("failed to create new %s named %s", kind, name)

	res, err := d.createCollectionStmt.Exec(kind, name)
	if err != nil {
		return 0, wrap(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, wrap(err)
	}

	return core.CollectionID(id), nil
}

// Associate creates a new association between a note and a collection, if it
// does not already exist.
func (d *CollectionDAO) Associate(noteId core.NoteID, collectionId core.CollectionID) (core.NoteCollectionID, error) {
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

func (d *CollectionDAO) findAssociation(noteId core.NoteID, collectionId core.CollectionID) (core.NoteCollectionID, error) {
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
		return core.NoteCollectionID(id.Int64), nil
	}
}

func (d *CollectionDAO) createAssociation(noteId core.NoteID, collectionId core.CollectionID) (core.NoteCollectionID, error) {
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

	return core.NoteCollectionID(id), nil
}

// RemoveAssociations deletes all associations with the given note.
func (d *CollectionDAO) RemoveAssociations(noteId core.NoteID) error {
	if !noteId.IsValid() {
		return fmt.Errorf("Note ID (%d) not valid", noteId)
	}

	_, err := d.removeAssociationsStmt.Exec(noteId)
	if err != nil {
		return errors.Wrapf(err, "failed to remove associations of note %d", noteId)
	}

	return nil
}
