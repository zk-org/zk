package sqlite

import (
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/core/note"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
)

// NoteIndexer persists note indexing results in the SQLite database.
// It implements the core port note.Indexer and acts as a facade to the DAOs.
type NoteIndexer struct {
	tx          Transaction
	notes       *NoteDAO
	collections *CollectionDAO
	logger      util.Logger
}

func NewNoteIndexer(notes *NoteDAO, collections *CollectionDAO, logger util.Logger) *NoteIndexer {
	return &NoteIndexer{
		notes:       notes,
		collections: collections,
		logger:      logger,
	}
}

// Indexed returns the list of indexed note file metadata.
func (i *NoteIndexer) Indexed() (<-chan paths.Metadata, error) {
	c, err := i.notes.Indexed()
	return c, errors.Wrap(err, "failed to get indexed notes")
}

// Add indexes a new note from its metadata.
func (i *NoteIndexer) Add(metadata note.Metadata) (core.NoteID, error) {
	wrap := errors.Wrapperf("%v: failed to index the note", metadata.Path)
	noteId, err := i.notes.Add(metadata)
	if err != nil {
		return SQLNoteID(0), wrap(err)
	}

	err = i.associateTags(noteId, metadata.Tags)
	if err != nil {
		return SQLNoteID(0), wrap(err)
	}

	return noteId, nil
}

// Update updates the metadata of an already indexed note.
func (i *NoteIndexer) Update(metadata note.Metadata) error {
	wrap := errors.Wrapperf("%v: failed to update note index", metadata.Path)

	noteId, err := i.notes.Update(metadata)
	if err != nil {
		return wrap(err)
	}

	err = i.collections.RemoveAssociations(noteId)
	if err != nil {
		return wrap(err)
	}

	err = i.associateTags(noteId, metadata.Tags)
	if err != nil {
		return wrap(err)
	}

	return err
}

func (i *NoteIndexer) associateTags(noteId SQLNoteID, tags []string) error {
	for _, tag := range tags {
		tagId, err := i.collections.FindOrCreate(note.CollectionKindTag, tag)
		if err != nil {
			return err
		}
		_, err = i.collections.Associate(noteId, tagId)
		if err != nil {
			return err
		}
	}

	return nil
}

// Remove deletes a note from the index.
func (i *NoteIndexer) Remove(path string) error {
	err := i.notes.Remove(path)
	return errors.Wrapf(err, "%v: failed to remove note index", path)
}
