package paths

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mickael-menu/zk/util"
)

// Walk emits the metadata of each file stored in the directory with the given extension.
// Hidden files and directories are ignored.
func Walk(basePath string, extension string, logger util.Logger) <-chan Metadata {
	extension = "." + extension

	c := make(chan Metadata, 50)
	go func() {
		defer close(c)

		err := filepath.Walk(basePath, func(abs string, info os.FileInfo, err error) error {
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
				if isHidden || filepath.Ext(filename) != extension {
					return nil
				}

				path, err := filepath.Rel(basePath, abs)
				if err != nil {
					logger.Println(err)
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
