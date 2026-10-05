package fs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/paths"
)

// FileStorage implements the port core.FileStorage.
type FileStorage struct {
	// Current working directory.
	workingDir string
	logger     util.Logger
}

// NewFileStorage creates a new instance of FileStorage using the given working
// directory as reference point for relative paths.
func NewFileStorage(workingDir string, logger util.Logger) (*FileStorage, error) {
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	return &FileStorage{workingDir, logger}, nil
}

func (fs *FileStorage) WorkingDir() string {
	return fs.workingDir
}

func (fs *FileStorage) SetWorkingDir(path string) {
	fs.workingDir = path
}

func (fs *FileStorage) Abs(path string) (string, error) {
	var err error
	if !filepath.IsAbs(path) {
		path = filepath.Join(fs.workingDir, path)
		path, err = filepath.Abs(path)
		if err != nil {
			return path, err
		}
	}

	return path, nil
}

func (fs *FileStorage) Rel(path string) (string, error) {
	return filepath.Rel(fs.workingDir, path)
}

func (fs *FileStorage) Canonical(path string) string {
	path = filepath.Clean(path)

	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		fs.logger.Err(err)
	} else {
		path = resolvedPath
	}

	return path
}

func (fs *FileStorage) FileExists(path string) (bool, error) {
	return paths.FileExists(path)
}

func (fs *FileStorage) DirExists(path string) (bool, error) {
	return paths.DirExists(path)
}

func (fs *FileStorage) IsDescendantOf(dir string, path string) (bool, error) {
	dir, err := fs.Abs(dir)
	if err != nil {
		return false, err
	}
	dir = fs.Canonical(dir)

	path, err = fs.Abs(path)
	if err != nil {
		return false, err
	}
	path = fs.Canonical(path)

	path, err = filepath.Rel(dir, path)
	if err != nil {
		return false, err
	}

	return !strings.HasPrefix(path, ".."), nil
}

func (fs *FileStorage) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (fs *FileStorage) Write(path string, content []byte) error {
	return paths.WriteFile(path, content)
}
