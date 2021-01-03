package zk

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
	"gopkg.in/djherbis/times.v1"
)

// NoteMetadata holds information about a particular note.
type NoteMetadata struct {
	Path      Path
	Title     string
	Body      string
	WordCount int
	Created   time.Time
	Modified  time.Time
	Checksum  string
}

type Indexer interface {
	Indexed() (<-chan FileMetadata, error)
	Add(metadata NoteMetadata) error
	Update(metadata NoteMetadata) error
	Remove(path Path) error
}

// Index indexes the content of the notes in the given directory.
func Index(dir Dir, indexer Indexer, logger util.Logger) error {
	wrap := errors.Wrapper("indexation failed")

	source := dir.Walk(logger)
	target, err := indexer.Indexed()
	if err != nil {
		return wrap(err)
	}

	return Diff(source, target, func(change DiffChange) error {
		switch change.Kind {
		case DiffAdded:
			metadata, err := noteMetadata(change.Path)
			if err == nil {
				err = indexer.Add(metadata)
			}
			logger.Err(err)

		case DiffModified:
			metadata, err := noteMetadata(change.Path)
			if err == nil {
				err = indexer.Update(metadata)
			}
			logger.Err(err)

		case DiffRemoved:
			indexer.Remove(change.Path)
		}
		return nil
	})
}

func noteMetadata(path Path) (NoteMetadata, error) {
	metadata := NoteMetadata{
		Path: path,
	}

	content, err := ioutil.ReadFile(path.Abs)
	if err != nil {
		return metadata, err
	}
	contentStr := string(content)
	metadata.Body = contentStr
	metadata.WordCount = len(strings.Fields(contentStr))
	metadata.Checksum = fmt.Sprintf("%x", sha256.Sum256(content))

	times, err := times.Stat(path.Abs)
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
