package paths

import (
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Expands home directory to absolute path.
// If no expansion is required or possible, the string is returned unaltered.
// If there is an error, an empty string is returned.
func ExpandHomeDir(path string) (string, error) {

	// Expand in case there are environment variables on the path
	path = os.ExpandEnv(path)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		path = home
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(home, path[2:])
	}

	return path, nil
}
