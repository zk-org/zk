package core

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
)

// NoteIndex persists and grants access to indexed information about the notes.
type NoteIndex interface {
	// Find retrieves the notes matching the given filtering and sorting criteria.
	Find(opts NoteFindOpts) ([]ContextualNote, error)
	// FindMinimal retrieves lightweight metadata for the notes matching the
	// given filtering and sorting criteria.
	FindMinimal(opts NoteFindOpts) ([]MinimalNote, error)

	// FindCollections retrieves all the collections of the given kind.
	FindCollections(kind CollectionKind, sorters []CollectionSorter) ([]Collection, error)

	// Indexed returns the list of indexed note file metadata.
	IndexedPaths() (<-chan paths.Metadata, error)
	// Add indexes a new note.
	Add(note Note) (NoteID, error)
	// Update resets the metadata of an already indexed note.
	Update(note Note) error
	// Remove deletes a note from the index.
	Remove(path string) error

	// Commit performs a set of operations atomically.
	Commit(transaction func(idx NoteIndex) error) error

	// NeedsReindexing returns whether all notes should be reindexed.
	NeedsReindexing() (bool, error)
	// SetNeedsReindexing indicates whether all notes should be reindexed.
	SetNeedsReindexing(needsReindexing bool) error
}

// NoteIndexingStats holds statistics about a notebook indexing process.
type NoteIndexingStats struct {
	// Number of notes in the source.
	SourceCount int `json:"sourceCount"`
	// Number of newly indexed notes.
	AddedCount int `json:"addedCount"`
	// Number of notes modified since last indexing.
	ModifiedCount int `json:"modifiedCount"`
	// Number of notes removed since last indexing.
	RemovedCount int `json:"removedCount"`
	// Duration of the indexing process.
	Duration time.Duration `json:"duration"`
}

// String implements Stringer
func (s NoteIndexingStats) String() string {
	return fmt.Sprintf(`Indexed %d %v in %v
  + %d added
  ~ %d modified
  - %d removed`,
		s.SourceCount,
		strutil.Pluralize("note", s.SourceCount),
		s.Duration.Round(500*time.Millisecond),
		s.AddedCount, s.ModifiedCount, s.RemovedCount,
	)
}

// indexTask indexes the notes in the given directory with the NoteIndex.
type indexTask struct {
	path   string
	config Config
	force  bool
	index  NoteIndex
	parser NoteParser
	logger util.Logger
}

func (t *indexTask) execute(callback func(change paths.DiffChange)) (NoteIndexingStats, error) {
	wrap := errors.Wrapper("indexing failed")

	stats := NoteIndexingStats{}
	startTime := time.Now()

	needsReindexing, err := t.index.NeedsReindexing()
	if err != nil {
		return stats, wrap(err)
	}

	force := t.force || needsReindexing

	shouldIgnorePath := func(path string) (bool, error) {
		group, err := t.config.GroupConfigForPath(path)
		if err != nil {
			return true, err
		}

		if filepath.Ext(path) != "."+group.Note.Extension {
			return true, nil
		}

		for _, ignoreGlob := range group.IgnoreGlobs() {
			matches, err := filepath.Match(ignoreGlob, path)
			if err != nil {
				return true, errors.Wrapf(err, "failed to match ignore glob %s to %s", ignoreGlob, path)
			}
			if matches {
				return true, nil
			}
		}

		return false, nil
	}

	source := paths.Walk(t.path, t.logger, shouldIgnorePath)

	target, err := t.index.IndexedPaths()
	if err != nil {
		return stats, wrap(err)
	}

	// FIXME: Use the FS?
	count, err := paths.Diff(source, target, force, func(change paths.DiffChange) error {
		callback(change)
		absPath := filepath.Join(change.Path)

		switch change.Kind {
		case paths.DiffAdded:
			stats.AddedCount += 1
			note, err := t.parser.ParseNoteAt(absPath)
			if note != nil {
				_, err = t.index.Add(*note)
			}
			t.logger.Err(err)

		case paths.DiffModified:
			stats.ModifiedCount += 1
			note, err := t.parser.ParseNoteAt(absPath)
			if note != nil {
				err = t.index.Update(*note)
			}
			t.logger.Err(err)

		case paths.DiffRemoved:
			stats.RemovedCount += 1
			err := t.index.Remove(change.Path)
			t.logger.Err(err)
		}
		return nil
	})

	stats.SourceCount = count
	stats.Duration = time.Since(startTime)

	if needsReindexing {
		err = t.index.SetNeedsReindexing(false)
	}

	return stats, wrap(err)
}
