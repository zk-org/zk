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
	db     *DB
	dao    *dao
	logger util.Logger
}

type dao struct {
	notes       *NoteDAO
	collections *CollectionDAO
	metadata    *MetadataDAO
}

func NewNoteIndex(db *DB, logger util.Logger) *NoteIndex {
	return &NoteIndex{
		db:     db,
		logger: logger,
	}
}

// Find implements core.NoteIndex.
func (ni *NoteIndex) Find(opts core.NoteFindOpts) (notes []core.ContextualNote, err error) {
	err = ni.commit(func(dao *dao) error {
		notes, err = dao.notes.Find(opts)
		return err
	})
	return
}

// FindMinimal implements core.NoteIndex.
func (ni *NoteIndex) FindMinimal(opts core.NoteFindOpts) (notes []core.MinimalNote, err error) {
	err = ni.commit(func(dao *dao) error {
		notes, err = dao.notes.FindMinimal(opts)
		return err
	})
	return
}

// FindCollections implements core.NoteIndex.
func (ni *NoteIndex) FindCollections(kind core.CollectionKind, sorters []core.CollectionSorter) (collections []core.Collection, err error) {
	err = ni.commit(func(dao *dao) error {
		collections, err = dao.collections.FindAll(kind, sorters)
		return err
	})
	return
}

// IndexedPaths implements core.NoteIndex.
func (ni *NoteIndex) IndexedPaths() (metadata <-chan paths.Metadata, err error) {
	err = ni.commit(func(dao *dao) error {
		metadata, err = dao.notes.Indexed()
		return err
	})
	err = errors.Wrap(err, "failed to get indexed notes")
	return
}

// Add implements core.NoteIndex.
func (ni *NoteIndex) Add(note core.Note) (id core.NoteID, err error) {
	err = ni.commit(func(dao *dao) error {
		id, err = dao.notes.Add(note)
		if err != nil {
			return err
		}

		return ni.associateTags(dao.collections, id, note.Tags)
	})

	err = errors.Wrapf(err, "%v: failed to index the note", note.Path)
	return
}

// Update implements core.NoteIndex.
func (ni *NoteIndex) Update(note core.Note) error {
	err := ni.commit(func(dao *dao) error {
		noteId, err := dao.notes.Update(note)
		if err != nil {
			return err
		}

		err = dao.collections.RemoveAssociations(noteId)
		if err != nil {
			return err
		}

		return ni.associateTags(dao.collections, noteId, note.Tags)
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
	err := ni.commit(func(dao *dao) error {
		return dao.notes.Remove(path)
	})
	return errors.Wrapf(err, "%v: failed to remove note from index", path)
}

// Commit implements core.NoteIndex.
func (ni *NoteIndex) Commit(transaction func(idx core.NoteIndex) error) error {
	return ni.commit(func(dao *dao) error {
		return transaction(&NoteIndex{
			db:     ni.db,
			dao:    dao,
			logger: ni.logger,
		})
	})
}

// NeedsReindexing implements core.NoteIndex.
func (ni *NoteIndex) NeedsReindexing() (needsReindexing bool, err error) {
	err = ni.commit(func(dao *dao) error {
		res, err := dao.metadata.Get(reindexingRequiredKey)
		needsReindexing = (res == "true")
		return err
	})
	return
}

// SetNeedsReindexing implements core.NoteIndex.
func (ni *NoteIndex) SetNeedsReindexing(needsReindexing bool) error {
	return ni.commit(func(dao *dao) error {
		value := "false"
		if needsReindexing {
			value = "true"
		}

		return dao.metadata.Set(reindexingRequiredKey, value)
	})
}

func (ni *NoteIndex) commit(transaction func(dao *dao) error) error {
	if ni.dao != nil {
		return transaction(ni.dao)
	} else {
		return ni.db.WithTransaction(func(tx Transaction) error {
			dao := dao{
				notes:       NewNoteDAO(tx, ni.logger),
				collections: NewCollectionDAO(tx, ni.logger),
				metadata:    NewMetadataDAO(tx),
			}
			return transaction(&dao)
		})
	}
}
