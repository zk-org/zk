package note

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/core"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/paths"
	strutil "github.com/mickael-menu/zk/util/strings"
	"github.com/relvacode/iso8601"
	"gopkg.in/djherbis/times.v1"
)

// Metadata holds information about a particular note.
type Metadata struct {
	Path       string
	Title      string
	Lead       string
	Body       string
	RawContent string
	WordCount  int
	Links      []Link
	Tags       []string
	Metadata   map[string]interface{}
	Created    time.Time
	Modified   time.Time
	Checksum   string
}

// IndexingStats holds metrics about an indexing process.
type IndexingStats struct {
	SourceCount   int
	AddedCount    int
	ModifiedCount int
	RemovedCount  int
	Duration      time.Duration
}

// String implements Stringer
func (s IndexingStats) String() string {
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

// Indexer persists the notes index.
type Indexer interface {
	// Indexed returns the list of indexed note file metadata.
	Indexed() (<-chan paths.Metadata, error)
	// Add indexes a new note from its metadata.
	Add(metadata Metadata) (core.NoteId, error)
	// Update updates the metadata of an already indexed note.
	Update(metadata Metadata) error
	// Remove deletes a note from the index.
	Remove(path string) error
}

// Index indexes the content of the notes in the given notebook.
func Index(zk *zk.Zk, force bool, parser Parser, indexer Indexer, logger util.Logger, callback func(change paths.DiffChange)) (IndexingStats, error) {
	wrap := errors.Wrapper("indexing failed")

	stats := IndexingStats{}
	startTime := time.Now()

	// FIXME: Use Extension defined in each DirConfig.
	source := paths.Walk(zk.Path, zk.Config.Note.Extension, logger)
	target, err := indexer.Indexed()
	if err != nil {
		return stats, wrap(err)
	}

	count, err := paths.Diff(source, target, force, func(change paths.DiffChange) error {
		callback(change)

		switch change.Kind {
		case paths.DiffAdded:
			stats.AddedCount += 1
			metadata, err := metadata(change.Path, zk, parser)
			if err == nil {
				_, err = indexer.Add(metadata)
			}
			logger.Err(err)

		case paths.DiffModified:
			stats.ModifiedCount += 1
			metadata, err := metadata(change.Path, zk, parser)
			if err == nil {
				err = indexer.Update(metadata)
			}
			logger.Err(err)

		case paths.DiffRemoved:
			stats.RemovedCount += 1
			err := indexer.Remove(change.Path)
			logger.Err(err)
		}
		return nil
	})

	stats.SourceCount = count
	stats.Duration = time.Since(startTime)

	return stats, wrap(err)
}

// metadata retrieves note metadata for the given file.
func metadata(path string, zk *zk.Zk, parser Parser) (Metadata, error) {
	metadata := Metadata{
		Path:  path,
		Links: []Link{},
		Tags:  []string{},
	}

	absPath := filepath.Join(zk.Path, path)
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return metadata, err
	}
	contentStr := string(content)
	contentParts, err := parser.Parse(contentStr)
	if err != nil {
		return metadata, err
	}
	metadata.Title = contentParts.Title.String()
	metadata.Lead = contentParts.Lead.String()
	metadata.Body = contentParts.Body.String()
	metadata.RawContent = contentStr
	metadata.WordCount = len(strings.Fields(contentStr))
	metadata.Links = make([]Link, 0)
	metadata.Tags = contentParts.Tags
	metadata.Metadata = contentParts.Metadata
	metadata.Checksum = fmt.Sprintf("%x", sha256.Sum256(content))

	for _, link := range contentParts.Links {
		if !strutil.IsURL(link.Href) {
			// Make the href relative to the notebook root.
			href := filepath.Join(filepath.Dir(absPath), link.Href)
			link.Href, err = zk.RelPath(href)
			if err != nil {
				return metadata, err
			}
		}
		metadata.Links = append(metadata.Links, link)
	}

	times, err := times.Stat(absPath)
	if err != nil {
		return metadata, err
	}

	metadata.Modified = times.ModTime().UTC()
	metadata.Created = creationDateFrom(metadata.Metadata, times)

	return metadata, nil
}

func creationDateFrom(metadata map[string]interface{}, times times.Timespec) time.Time {
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
