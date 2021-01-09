package file

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
)

// Walk emits the metadata of each file stored in the directory.
func Walk(dir zk.Dir, extension string, logger util.Logger) <-chan Metadata {
	extension = "." + extension

	c := make(chan Metadata, 50)
	go func() {
		defer close(c)

		err := filepath.Walk(dir.Path, func(abs string, info os.FileInfo, err error) error {
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

				path, err := filepath.Rel(dir.Path, abs)
				if err != nil {
					logger.Println(err)
					return nil
				}

				curDir := filepath.Dir(path)
				if curDir == "." {
					curDir = ""
				}

				c <- Metadata{
					Path:     filepath.Join(dir.Name, curDir, filename),
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
