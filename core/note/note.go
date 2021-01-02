package note

import (
	"time"
)

// Path holds a note path relative to its slip box.
type Path struct {
	Dir      string
	Filename string
}

// FileMetadata holds information about a note file.
type FileMetadata struct {
	Path     Path
	Created  time.Time
	Modified time.Time
}

// Metadata holds information about a particular note.
type Metadata struct {
	Title     string
	Content   string
	WordCount int
}
