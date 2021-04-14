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
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
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
