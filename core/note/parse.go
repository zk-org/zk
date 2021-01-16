package note

import (
	"github.com/mickael-menu/zk/util/opt"
)

type Content struct {
	// Title is the heading of the note.
	Title opt.String
	// Lead is the opening paragraph or section of the note.
	Lead opt.String
	// Body is the content of the note, including the Lead but without the Title.
	Body opt.String
}

type Parser interface {
	Parse(source string) (Content, error)
}
