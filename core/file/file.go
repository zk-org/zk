package file

import (
	"time"
)

// Metadata holds information about a slip box file.
type Metadata struct {
	Path     string
	Modified time.Time
}
