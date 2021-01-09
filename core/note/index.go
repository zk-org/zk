package note

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/paths"
	"gopkg.in/djherbis/times.v1"
)

// Metadata holds information about a particular note.
type Metadata struct {
	Path      string
	Title     string
	Body      string
	WordCount int
	Created   time.Time
	Modified  time.Time
	Checksum  string
}

// Indexer persists the notes index.
type Indexer interface {
	// Indexed returns the list of indexed note file metadata.
	Indexed() (<-chan paths.Metadata, error)
	// Add indexes a new note from its metadata.
	Add(metadata Metadata) error
	// Update updates the metadata of an already indexed note.
	Update(metadata Metadata) error
	// Remove deletes a note from the index.
	Remove(path string) error
}

// Index indexes the content of the notes in the given directory.
func Index(dir zk.Dir, indexer Indexer, logger util.Logger) error {
	wrap := errors.Wrapper("indexation failed")

	source := paths.Walk(dir.Path, dir.Config.Extension, logger)
	target, err := indexer.Indexed()
	if err != nil {
		return wrap(err)
	}

	err = paths.Diff(source, target, func(change paths.DiffChange) error {
		switch change.Kind {
		case paths.DiffAdded:
			metadata, err := metadata(change.Path, dir.Path)
			if err == nil {
				err = indexer.Add(metadata)
			}
			logger.Err(err)

		case paths.DiffModified:
			metadata, err := metadata(change.Path, dir.Path)
			if err == nil {
				err = indexer.Update(metadata)
			}
			logger.Err(err)

		case paths.DiffRemoved:
			err := indexer.Remove(change.Path)
			logger.Err(err)
		}
		return nil
	})

	return wrap(err)
}

// metadata retrieves note metadata for the given file.
func metadata(path string, basePath string) (Metadata, error) {
	metadata := Metadata{
		Path: path,
	}

	absPath := filepath.Join(basePath, path)
	content, err := ioutil.ReadFile(absPath)
	if err != nil {
		return metadata, err
	}
	contentStr := string(content)
	contentParts := Parse(contentStr)
	metadata.Title = contentParts.Title
	metadata.Body = contentParts.Body
	metadata.WordCount = len(strings.Fields(contentStr))
	metadata.Checksum = fmt.Sprintf("%x", sha256.Sum256(content))

	times, err := times.Stat(absPath)
	if err != nil {
		return metadata, err
	}

	metadata.Modified = times.ModTime()
	if times.HasBirthTime() {
		metadata.Created = times.BirthTime()
	} else {
		metadata.Created = time.Now()
	}

	return metadata, nil
}
