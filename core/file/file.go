package file

import (
	"path/filepath"
	"time"
)

// Metadata holds information about a slip box file.
type Metadata struct {
	Path     Path
	Modified time.Time
}

// Path holds a file path relative to a slip box.
type Path struct {
	Dir      string
	Filename string
	Abs      string
}

// Less returns whether ther receiver path is located before the given one,
// lexicographically.
func (p Path) Less(other Path) bool {
	switch {
	case p.Dir < other.Dir:
		return true
	case p.Dir > other.Dir:
		return false
	default:
		return p.Filename < other.Filename
	}
}

func (p Path) String() string {
	return filepath.Join(p.Dir, p.Filename)
}
