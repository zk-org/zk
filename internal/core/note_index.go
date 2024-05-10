package core

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/paths"
	strutil "github.com/zk-org/zk/internal/util/strings"
)

// NoteIndex persists and grants access to indexed information about the notes.
type NoteIndex interface {
	// Find retrieves the notes matching the given filtering and sorting criteria.
	Find(opts NoteFindOpts) ([]ContextualNote, error)
	// FindMinimal retrieves lightweight metadata for the notes matching the
	// given filtering and sorting criteria.
	FindMinimal(opts NoteFindOpts) ([]MinimalNote, error)

	// Find link match returns the best note match for a given link href,
	// relative to baseDir.
	FindLinkMatch(baseDir string, href string, linkType LinkType) (NoteID, error)

	// FindLinksBetweenNotes retrieves the links between the given notes.
	FindLinksBetweenNotes(ids []NoteID) ([]ResolvedLink, error)

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

// NoteIndexOpts holds the options for the indexing process.
type NoteIndexOpts struct {
	// When true, existing notes will be reindexed.
	Force   bool
	Verbose bool
}

// indexTask indexes the notes in the given directory with the NoteIndex.
type indexTask struct {
	path    string
	config  Config
	force   bool
	verbose bool
	index   NoteIndex
	parser  NoteParser
	logger  util.Logger
}

func (t *indexTask) execute(callback func(change paths.DiffChange)) (NoteIndexingStats, error) {
	wrap := errors.Wrapper("indexing failed")

	stats := NoteIndexingStats{}
	startTime := time.Now()

	needsReindexing, err := t.index.NeedsReindexing()
	if err != nil {
		return stats, wrap(err)
	}

	print := func(message string) {
		if t.verbose {
			fmt.Println(message)
		}
	}

	force := t.force || needsReindexing

	type IgnoredFile struct {
		Path   string
		Reason string
	}
	ignoredFiles := []IgnoredFile{}

	shouldIgnorePath := func(path string) (bool, error) {
		notifyIgnored := func(reason string) {
			ignoredFiles = append(ignoredFiles, IgnoredFile{
				Path:   path,
				Reason: reason,
			})
		}

		group, err := t.config.GroupConfigForPath(path)
		if err != nil {
			return true, err
		}

		if filepath.Ext(path) != "."+group.Note.Extension {
			notifyIgnored("expected extension \"" + group.Note.Extension + "\"")
			return true, nil
		}

		for _, ignoreGlob := range group.ExcludeGlobs() {
			matches, err := doublestar.PathMatch(ignoreGlob, path)
			if err != nil {
				return true, errors.Wrapf(err, "failed to match exclude glob %s to %s", ignoreGlob, path)
			}
			if matches {
				notifyIgnored("matched exclude glob \"" + ignoreGlob + "\"")
				return true, nil
			}
		}

		return false, nil
	}

	notebookPath := &NotebookPath{Path: t.path}
	source := paths.Walk(t.path, t.logger, notebookPath.Filename(), shouldIgnorePath)

	target, err := t.index.IndexedPaths()
	if err != nil {
		return stats, wrap(err)
	}

	// FIXME: Use the FS?
	count, err := paths.Diff(source, target, force, func(change paths.DiffChange) error {
		callback(change)
		print("- " + change.Kind.String() + " " + change.Path)
		absPath := filepath.Join(t.path, change.Path)

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

	for _, ignored := range ignoredFiles {
		print("- ignored " + ignored.Path + ": " + ignored.Reason)
	}

	stats.SourceCount = count
	stats.Duration = time.Since(startTime)

	if needsReindexing {
		err = t.index.SetNeedsReindexing(false)
	}

	print("")
	return stats, wrap(err)
}
