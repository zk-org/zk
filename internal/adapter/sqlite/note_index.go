package sqlite

import (
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
)

// NoteIndex persists note indexing results in the SQLite database.
// It implements the port core.NoteIndex and acts as a facade to the DAOs.
type NoteIndex struct {
	db          *DB
	notes       *NoteDAO
	collections *CollectionDAO
	logger      util.Logger
}

func NewNoteIndex(db *DB, logger util.Logger) *NoteIndex {
	return &NoteIndex{
		db:     db,
		logger: logger,
	}
}

// Find implements core.NoteIndex.
func (ni *NoteIndex) Find(opts core.NoteFindOpts) (notes []core.ContextualNote, err error) {
	err = ni.commit(func(dao *NoteDAO, _ *CollectionDAO) error {
		notes, err = dao.Find(opts)
		return err
	})
	return
}

// FindMinimal implements core.NoteIndex.
func (ni *NoteIndex) FindMinimal(opts core.NoteFindOpts) (notes []core.MinimalNote, err error) {
	panic("not implemented")
}

// IndexedPaths implements core.NoteIndex.
func (ni *NoteIndex) IndexedPaths() (metadata <-chan paths.Metadata, err error) {
	err = ni.commit(func(notes *NoteDAO, collections *CollectionDAO) error {
		metadata, err = notes.Indexed()
		return err
	})
	err = errors.Wrap(err, "failed to get indexed notes")
	return
}

// Add implements core.NoteIndex.
func (ni *NoteIndex) Add(note core.Note) (id core.NoteID, err error) {
	err = ni.commit(func(notes *NoteDAO, collections *CollectionDAO) error {
		id, err = notes.Add(note)
		if err != nil {
			return err
		}

		return ni.associateTags(collections, id, note.Tags)
	})

	err = errors.Wrapf(err, "%v: failed to index the note", note.Path)
	return
}

// Update implements core.NoteIndex.
func (ni *NoteIndex) Update(note core.Note) error {
	err := ni.commit(func(notes *NoteDAO, collections *CollectionDAO) error {
		noteId, err := notes.Update(note)
		if err != nil {
			return err
		}

		err = collections.RemoveAssociations(noteId)
		if err != nil {
			return err
		}

		return ni.associateTags(collections, noteId, note.Tags)
	})

	return errors.Wrapf(err, "%v: failed to update note index", note.Path)
}

func (ni *NoteIndex) associateTags(collections *CollectionDAO, noteId core.NoteID, tags []string) error {
	for _, tag := range tags {
		tagId, err := collections.FindOrCreate(core.CollectionKindTag, tag)
		if err != nil {
			return err
		}
		_, err = collections.Associate(noteId, tagId)
		if err != nil {
			return err
		}
	}

	return nil
}

// Remove implements core.NoteIndex
func (ni *NoteIndex) Remove(path string) error {
	err := ni.commit(func(notes *NoteDAO, collections *CollectionDAO) error {
		return notes.Remove(path)
	})
	return errors.Wrapf(err, "%v: failed to remove note from index", path)
}

// Commit implements core.NoteIndex.
func (ni *NoteIndex) Commit(transaction func(idx core.NoteIndex) error) error {
	return ni.commit(func(notes *NoteDAO, collections *CollectionDAO) error {
		return transaction(&NoteIndex{
			db:          ni.db,
			notes:       notes,
			collections: collections,
			logger:      ni.logger,
		})
	})
}

func (ni *NoteIndex) commit(transaction func(notes *NoteDAO, collections *CollectionDAO) error) error {
	if ni.notes != nil && ni.collections != nil {
		return transaction(ni.notes, ni.collections)
	} else {
		return ni.db.WithTransaction(func(tx Transaction) error {
			notes := NewNoteDAO(tx, ni.logger)
			collections := NewCollectionDAO(tx, ni.logger)
			return transaction(notes, collections)
		})
	}
}
