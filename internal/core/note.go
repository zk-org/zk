package core

import (
	"path/filepath"
	"time"

	"github.com/zk-org/zk/internal/util/paths"
)

// NoteID represents the unique ID of a note collection relative to a given
// NoteIndex implementation.
type NoteID int64

func (id NoteID) IsValid() bool {
	return id > 0
}

// MinimalNote holds a Note's title and path information, for display purposes.
type MinimalNote struct {
	// Unique ID of this note in a notebook.
	ID NoteID
	// Path relative to the root of the notebook.
	Path string
	// Title of the note.
	Title string
	// JSON dictionary of raw metadata extracted from the frontmatter.
	Metadata map[string]interface{}
}

// Note holds the metadata and content of a single note.
type Note struct {
	// Unique ID of this note in a NoteRepository.
	ID NoteID
	// Path relative to the root of the notebook.
	Path string
	// Title of the note.
	Title string
	// First paragraph from the note body.
	Lead string
	// Content of the note, after any frontmatter and title heading.
	Body string
	// Whole raw content of the note.
	RawContent string
	// Number of words found in the content.
	WordCount int
	// List of outgoing links (internal or external) found in the content.
	Links []Link
	// List of tags found in the content.
	Tags []string
	// JSON dictionary of raw metadata extracted from the frontmatter.
	Metadata map[string]interface{}
	// Date of creation.
	Created time.Time
	// Date of last modification.
	Modified time.Time
	// Checksum of the note content.
	Checksum string
}

func (n Note) AsMinimalNote() MinimalNote {
	return MinimalNote{
		ID:       n.ID,
		Path:     n.Path,
		Title:    n.Title,
		Metadata: n.Metadata,
	}
}

// Filename returns the filename portion of the note path.
func (n Note) Filename() string {
	return filepath.Base(n.Path)
}

// FilenameStem returns the filename portion of the note path, excluding its
// file extension.
func (n Note) FilenameStem() string {
	return paths.FilenameStem(n.Path)
}

// ContextualNote holds a Note and context-sensitive content snippets.
//
// This is used for example:
//   - to show an excerpt with highlighted search terms
//   - when following links, to print the source paragraph
type ContextualNote struct {
	Note
	// List of context-sensitive excerpts from the note.
	Snippets []string
}
