package core

// FileStorage is a port providing read and write access to a file storage.
type FileStorage interface {

	// Abs makes the given file path absolute if needed, using the FileStorage
	// working directory.
	Abs(path string) (string, error)

	// Rel makes the given absolute file path relative to the current working
	// directory.
	Rel(path string) (string, error)

	// FileExists returns whether a file exists at the given file path.
	FileExists(path string) (bool, error)

	// DirExists returns whether a directory exists at the given file path.
	DirExists(path string) (bool, error)

	// IsDescendantOf returns whether the given path is dir or one of its descendants.
	IsDescendantOf(dir string, path string) (bool, error)

	// Read returns the bytes content of the file at the given file path.
	Read(path string) ([]byte, error)

	// Write creates or overwrite the content at the given file path, creating
	// any intermediate directories if needed.
	Write(path string, content []byte) error
}
