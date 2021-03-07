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
	// Tags is the list of tags found in the note content.
	Tags []string
	// Links is the list of outbound links found in the note.
	Links []Link
	// Additional metadata. For example, extracted from a YAML frontmatter.
	Metadata map[string]interface{}
}

// Link links a note to another note or an external resource.
type Link struct {
	Title    string
	Href     string
	External bool
	Rels     []string
	Snippet  string
}

// LinkRelation defines the relationship between a link's source and target.
type LinkRelation string

const (
	// LinkRelationDown defines the target note as a child of the source.
	LinkRelationDown LinkRelation = "down"
	// LinkRelationDown defines the target note as a parent of the source.
	LinkRelationUp LinkRelation = "up"
)

type Parser interface {
	Parse(source string) (*Content, error)
}

// CollectionKind defines a kind of note collection, such as tags.
type CollectionKind string

const (
	CollectionKindTag CollectionKind = "tag"
)
