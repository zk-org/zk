package core

import (
	"os"
	"path/filepath"
)

// fileStorageMock implements an in-memory FileStorage for testing purposes.
type fileStorageMock struct {
	// Working directory used to calculate relative paths.
	workingDir string
	// File content indexed by their path in this file storage.
	files map[string]string
	// Existing directories
	dirs []string
}

func newFileStorageMock(workingDir string, dirs []string) *fileStorageMock {
	return &fileStorageMock{
		workingDir: workingDir,
		files:      map[string]string{},
		dirs:       dirs,
	}
}

func (fs *fileStorageMock) WorkingDir() string {
	return fs.workingDir
}

func (fs *fileStorageMock) Abs(path string) (string, error) {
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

func (fs *fileStorageMock) Rel(path string) (string, error) {
	return filepath.Rel(fs.workingDir, path)
}

func (fs *fileStorageMock) Canonical(path string) string {
	return path
}

func (fs *fileStorageMock) FileExists(path string) (bool, error) {
	_, ok := fs.files[path]
	return ok, nil
}

func (fs *fileStorageMock) DirExists(path string) (bool, error) {
	for _, dir := range fs.dirs {
		if dir == path {
			return true, nil
		}
	}
	return false, nil
}

func (fs *fileStorageMock) fileInfo(path string) (*os.FileInfo, error) {
	panic("not implemented")
}

func (fs *fileStorageMock) IsDescendantOf(dir string, path string) (bool, error) {
	panic("not implemented")
}

func (fs *fileStorageMock) Read(path string) ([]byte, error) {
	content, _ := fs.files[path]
	return []byte(content), nil
}

func (fs *fileStorageMock) Write(path string, content []byte) error {
	fs.files[path] = string(content)
	return nil
}
