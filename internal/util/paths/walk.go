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
//
// Symbolic links are followed, and the notes of a linked directory are
// emitted under the link's path.
func Walk(basePath string, logger util.Logger, notebookRoot string, shouldIgnorePath func(string, bool) (bool, error)) <-chan Metadata {
	c := make(chan Metadata, 50)
	go func() {
		defer close(c)

		// Directories are keyed by their resolved path and walked at most once,
		// so symlink cycles terminate and no note is emitted twice.
		visitedDirs := map[string]bool{}

		// Link targets are compared against the notebook, so both need their
		// resolved form.
		if resolved, err := filepath.EvalSymlinks(basePath); err == nil {
			basePath = resolved
		}

		var walk func(root string, prefix string) error
		walk = func(root string, prefix string) error {
			visitedDirs[root] = true

			return filepath.Walk(root, func(abs string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				// The link itself was already checked before walking its target.
				if abs == root {
					return nil
				}

				isLink := info.Mode()&os.ModeSymlink != 0
				if isLink {
					info, err = os.Stat(abs)
					if err != nil {
						logger.Println(err)
						return nil
					}
				}

				// filepath.Walk only prunes real directories: returning SkipDir
				// for a link would skip the remaining entries of its parent.
				skipDir := filepath.SkipDir
				if isLink {
					skipDir = nil
				}

				filename := info.Name()
				isHidden := strings.HasPrefix(filename, ".")
				isNotebookRoot := filename == notebookRoot

				path, err := filepath.Rel(root, abs)
				if err != nil {
					logger.Println(err)
					return nil
				}
				path = filepath.Join(prefix, path)

				if info.IsDir() {
					if isHidden && !isNotebookRoot {
						return skipDir
					}
					// Prune excluded directories.
					if !isNotebookRoot {
						shouldIgnore, err := shouldIgnorePath(path, true)
						if err != nil {
							logger.Println(err)
							return nil
						}
						if shouldIgnore {
							return skipDir
						}
					}

					if isLink {
						target, err := filepath.EvalSymlinks(abs)
						if err != nil {
							logger.Println(err)
							return nil
						}
						// A target inside the notebook is already reached
						// through its real path.
						if visitedDirs[target] || isDescendant(target, basePath) {
							return nil
						}
						return walk(target, path)
					}

					if visitedDirs[abs] {
						return filepath.SkipDir
					}
					visitedDirs[abs] = true

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
		}

		if err := walk(basePath, ""); err != nil {
			logger.Println(err)
		}
	}()

	return c
}

func isDescendant(path string, dir string) bool {
	return path == dir || strings.HasPrefix(path, dir+string(filepath.Separator))
}
