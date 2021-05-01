package core

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
	"github.com/relvacode/iso8601"
	"gopkg.in/djherbis/times.v1"
)

// NoteIndex persists and grants access to indexed information about the notes.
type NoteIndex interface {

	// Find retrieves the notes matching the given filtering and sorting criteria.
	Find(opts NoteFindOpts) ([]ContextualNote, error)
	// FindMinimal retrieves lightweight metadata for the notes matching the
	// given filtering and sorting criteria.
	FindMinimal(opts NoteFindOpts) ([]MinimalNote, error)

	// FindCollections retrieves all the collections of the given kind.
	FindCollections(kind CollectionKind) ([]Collection, error)

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
	notebook *Notebook
	force    bool
	index    NoteIndex
	parser   NoteParser
	logger   util.Logger
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

	// FIXME: Use Extension defined in each DirConfig.
	source := paths.Walk(t.notebook.Path, t.notebook.Config.Note.Extension, t.logger)
	target, err := t.index.IndexedPaths()
	if err != nil {
		return stats, wrap(err)
	}

	// FIXME: Use the FS?
	count, err := paths.Diff(source, target, force, func(change paths.DiffChange) error {
		callback(change)

		switch change.Kind {
		case paths.DiffAdded:
			stats.AddedCount += 1
			note, err := t.noteAt(change.Path)
			if err == nil {
				_, err = t.index.Add(note)
			}
			t.logger.Err(err)

		case paths.DiffModified:
			stats.ModifiedCount += 1
			note, err := t.noteAt(change.Path)
			if err == nil {
				err = t.index.Update(note)
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

// noteAt parses a Note at the given path.
func (t *indexTask) noteAt(path string) (Note, error) {
	wrap := errors.Wrapper(path)

	note := Note{
		Path:  path,
		Links: []Link{},
		Tags:  []string{},
	}

	absPath := filepath.Join(t.notebook.Path, path)
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return note, wrap(err)
	}
	contentStr := string(content)
	contentParts, err := t.parser.Parse(contentStr)
	if err != nil {
		return note, wrap(err)
	}
	note.Title = contentParts.Title.String()
	note.Lead = contentParts.Lead.String()
	note.Body = contentParts.Body.String()
	note.RawContent = contentStr
	note.WordCount = len(strings.Fields(contentStr))
	note.Links = make([]Link, 0)
	note.Tags = contentParts.Tags
	note.Metadata = contentParts.Metadata
	note.Checksum = fmt.Sprintf("%x", sha256.Sum256(content))

	for _, link := range contentParts.Links {
		if !strutil.IsURL(link.Href) {
			// Make the href relative to the notebook root.
			href := filepath.Join(filepath.Dir(absPath), link.Href)
			link.Href, err = t.notebook.RelPath(href)
			if err != nil {
				return note, wrap(err)
			}
		}
		note.Links = append(note.Links, link)
	}

	times, err := times.Stat(absPath)
	if err != nil {
		return note, wrap(err)
	}

	note.Modified = times.ModTime().UTC()
	note.Created = t.creationDateFrom(note.Metadata, times)

	return note, nil
}

func (t *indexTask) creationDateFrom(metadata map[string]interface{}, times times.Timespec) time.Time {
	// Read the creation date from the YAML frontmatter `date` key.
	if dateVal, ok := metadata["date"]; ok {
		if dateStr, ok := dateVal.(string); ok {
			if time, err := iso8601.ParseString(dateStr); err == nil {
				return time
			}
			// Omitting the `T` is common
			if time, err := time.Parse("2006-01-02 15:04:05", dateStr); err == nil {
				return time
			}
			if time, err := time.Parse("2006-01-02 15:04", dateStr); err == nil {
				return time
			}
		}
	}

	if times.HasBirthTime() {
		return times.BirthTime().UTC()
	}

	return time.Now().UTC()
}
