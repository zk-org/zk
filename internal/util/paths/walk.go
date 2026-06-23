package paths

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zk-org/zk/internal/util"
)

// Walk emits the metadata of each file stored in the directory if they pass
// the given shouldIgnorePath closure. Hidden files and directories are ignored,
// as are directories rejected by shouldIgnorePath, which are pruned from the
// walk instead of being traversed file by file.
func Walk(basePath string, logger util.Logger, notebookRoot string, shouldIgnorePath func(string, bool) (bool, error)) <-chan Metadata {
	c := make(chan Metadata, 50)
	go func() {
		defer close(c)

		err := filepath.Walk(basePath, func(abs string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			filename := info.Name()
			isHidden := strings.HasPrefix(filename, ".")
			isNotebookRoot := filename == notebookRoot

			path, err := filepath.Rel(basePath, abs)
			if err != nil {
				logger.Println(err)
				return nil
			}

			if info.IsDir() {
				if isHidden && !isNotebookRoot {
					return filepath.SkipDir
				}
				// Prune excluded directories.
				if !isNotebookRoot {
					shouldIgnore, err := shouldIgnorePath(path, true)
					if err != nil {
						logger.Println(err)
						return nil
					}
					if shouldIgnore {
						return filepath.SkipDir
					}
				}

			} else {
				shouldIgnore, err := shouldIgnorePath(path, false)
				if err != nil {
					logger.Println(err)
					return nil
				}
				if isHidden || shouldIgnore {
					return nil
				}

				c <- Metadata{
					Path:     path,
					Modified: info.ModTime().UTC(),
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
