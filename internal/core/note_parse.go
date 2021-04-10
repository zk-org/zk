package core

import (
	"github.com/mickael-menu/zk/internal/util/opt"
)

// NoteParser parses a note's raw content into its components.
type NoteParser interface {
	Parse(content string) (*ParsedNote, error)
}

// ParsedNote holds the data parsed from the note content.
type ParsedNote struct {
	// Title is the heading of the note.
	Title opt.String
	// Lead is the opening paragraph or section of the note.
	Lead opt.String
	// Body is the content of the note, including the Lead but without the Title.
	Body opt.String
	// Tags is the list of tags found in the note content.
	Tags []string
	// Links is the list of outbound links found in the note.
	Links []Link
	// Additional metadata. For example, extracted from a YAML frontmatter.
	Metadata map[string]interface{}
}
