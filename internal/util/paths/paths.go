package paths

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/zk-org/zk/internal/util/errors"
)

// Metadata holds information about a file path.
type Metadata struct {
	Path     string
	Modified time.Time
}

// Exists returns whether the given path exists on the file system.
func Exists(path string) (bool, error) {
	fi, err := fileInfo(path)
	if err != nil {
		return false, err
	} else {
		return fi != nil, nil
	}
}

// DirExists returns whether the given path exists and is a directory.
func DirExists(path string) (bool, error) {
	fi, err := fileInfo(path)
	if err != nil {
		return false, err
	} else {
		return fi != nil && (*fi).Mode().IsDir(), nil
	}
}

func fileInfo(path string) (*os.FileInfo, error) {
	if fi, err := os.Stat(path); err == nil {
		return &fi, nil
	} else if os.IsNotExist(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

// FilenameStem returns the filename component of the given path,
// after removing its file extension.
func FilenameStem(path string) string {
	path = DropExt(path)
	return filepath.Base(path)
}

// DropExt returns the path after removing any file extension.
func DropExt(path string) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext)
}

// WriteString writes the given content into a new file at the given path,
// creating any intermediate directories if needed.
func WriteString(path string, content string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != ".." {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

// Expands environment variables and `~`, returning an absolute path.
func ExpandPath(path string) (string, error) {

	if strings.HasPrefix(path, "~") {
		// resolve necessary variables
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		home := usr.HomeDir
		if home == "" {
			err := errors.New("Could not find user's home directory.")
			return "", err
		}

		// construct as abs path
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(home, path[2:])
		}
	}

	path = os.ExpandEnv(path)

	return path, nil
}
