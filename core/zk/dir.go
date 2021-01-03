package zk

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/util"
)

// Dir represents a directory inside a slip box.
type Dir struct {
	// Name of the directory, which is the path relative to the slip box's root.
	Name string
	// Absolute path to the directory.
	Path string
	// User configuration for this directory.
	Config DirConfig
}

// FileMetadata holds information about a note file.
type FileMetadata struct {
	Path     Path
	Modified time.Time
}

// Walk emits the metadata of each note stored in the directory.
func (d Dir) Walk(logger util.Logger) <-chan FileMetadata {
	c := make(chan FileMetadata, 50)
	go func() {
		defer close(c)

		err := filepath.Walk(d.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			filename := info.Name()
			isHidden := strings.HasPrefix(filename, ".")

			if info.IsDir() {
				if isHidden {
					return filepath.SkipDir
				}

			} else {
				// FIXME: Customize extension in config
				if isHidden || filepath.Ext(filename) != ".md" {
					return nil
				}

				path, err := filepath.Rel(d.Path, path)
				if err != nil {
					logger.Println(err)
					return nil
				}

				curDir := filepath.Dir(path)
				if curDir == "." {
					curDir = ""
				}

				c <- FileMetadata{
					Path: Path{
						Dir:      filepath.Join(d.Name, curDir),
						Filename: filename,
					},
					Modified: info.ModTime(),
				}
			}

			return nil
		})

		if err != nil {
			logger.Println(err)
		}
	}()

	return c
}
