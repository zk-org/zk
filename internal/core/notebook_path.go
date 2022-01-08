package core

import "path/filepath"

type NotebookPath struct {
	// Path of a notebook file, relative to the notebook dir.
	Path string
	// Root directory of the notebook.
	BasePath string
	// Current working directory.
	WorkingDir string
}

// Filename returns the filename of the notebook file.
func (p NotebookPath) Filename() string {
	return filepath.Base(p.Path)
}

// AbsPath returns the absolute path to the notebook file.
func (p NotebookPath) AbsPath() string {
	return filepath.Join(p.BasePath, p.Path)
}

// PathRelToWorkingDir returns the path to the notebook file relative to the
// working dir. If the working dir is not set, returns the path relative to the
// notebook dir.
func (p NotebookPath) PathRelToWorkingDir() (string, error) {
	if p.WorkingDir == "" {
		return p.Path, nil
	}
	return filepath.Rel(p.WorkingDir, p.AbsPath())
}
