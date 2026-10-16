package paths

import (
	"os"
	"path/filepath"

	"github.com/zk-org/zk/internal/util"
)

// Stat emits the current metadata of the given notes, which have to be sorted
// by path like the source handed to Diff.
//
// A note which does not exist anymore is left out of the source, which makes
// Diff report it as removed. Any other failure to read a note keeps its indexed
// modification time, so that a transient error does not remove it from the
// index.
//
// This is an alternative to Walk for notebooks where traversing the whole
// directory is too expensive (see the `notebook.index` setting).
func Stat(basePath string, notes []Metadata, logger util.Logger) <-chan Metadata {
	c := make(chan Metadata, 50)
	go func() {
		defer close(c)

		for _, note := range notes {
			info, err := os.Stat(filepath.Join(basePath, note.Path))
			switch {
			case os.IsNotExist(err):
				continue

			case err != nil:
				logger.Println(err)

			case info.IsDir():
				// The path is not a note anymore.
				continue

			default:
				note.Modified = info.ModTime().UTC()
			}

			c <- note
		}
	}()

	return c
}

// Collect drains a channel of metadata into a slice.
func Collect(c <-chan Metadata) []Metadata {
	metadata := []Metadata{}
	for m := range c {
		metadata = append(metadata, m)
	}
	return metadata
}
