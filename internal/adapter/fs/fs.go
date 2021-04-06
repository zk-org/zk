package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileStorage implements the port core.FileStorage.
type FileStorage struct {
	// Current working directory.
	wd string
}

// NewFileStorage creates a new instance of FileStorage using the given working
// directory as reference point for relative paths.
func NewFileStorage(workingDir string) (*FileStorage, error) {
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	return &FileStorage{workingDir}, nil
}

func (fs *FileStorage) Abs(path string) (string, error) {
	var err error
	if !filepath.IsAbs(path) {
		path = filepath.Join(fs.wd, path)
		path, err = filepath.Abs(path)
		if err != nil {
			return path, err
		}
	}

	return path, nil
}

func (fs *FileStorage) FileExists(path string) (bool, error) {
	fi, err := fs.fileInfo(path)
	if err != nil {
		return false, err
	} else {
		return fi != nil && (*fi).Mode().IsRegular(), nil
	}
}

func (fs *FileStorage) DirExists(path string) (bool, error) {
	fi, err := fs.fileInfo(path)
	if err != nil {
		return false, err
	} else {
		return fi != nil && (*fi).Mode().IsDir(), nil
	}
}

func (fs *FileStorage) fileInfo(path string) (*os.FileInfo, error) {
	if fi, err := os.Stat(path); err == nil {
		return &fi, nil
	} else if os.IsNotExist(err) {
		return nil, nil
	} else {
		return nil, err
	}
}

func (fs *FileStorage) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (fs *FileStorage) Write(path string, content []byte) error {
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
	_, err = f.Write(content)
	return err
}
