package core

import (
	"time"

	"github.com/mickael-menu/zk/internal/util/paths"
)

// NoteIndex persists and grants access to indexed information about the notes.
type NoteIndex interface {

	// Find retrieves the notes matching the given filtering and sorting criteria.
	Find(opts NoteFindOpts) ([]ContextualNote, error)
	// FindMinimal retrieves lightweight metadata for the notes matching the
	// given filtering and sorting criteria.
	FindMinimal(opts NoteFindOpts) ([]MinimalNote, error)

	// Indexed returns the list of indexed note file metadata.
	IndexedPaths() (<-chan paths.Metadata, error)
	// Add indexes a new note from its metadata.
	Add(note Note) (NoteID, error)
	// Update resets the metadata of an already indexed note.
	Update(note Note) error
	// Remove deletes a note from the index.
	Remove(path string) error

	// Commit performs a set of operations atomically.
	Commit(transaction func(idx NoteIndex) error) error
}

// NoteIndexingStats holds statistics about a notebook indexing process.
type NoteIndexingStats struct {
	// Number of notes in the source.
	SourceCount int
	// Number of newly indexed notes.
	AddedCount int
	// Number of notes modified since last indexing.
	ModifiedCount int
	// Number of notes removed since last indexing.
	RemovedCount int
	// Duration of the indexing process.
	Duration time.Duration
}
