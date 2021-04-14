package core

// Link represents a link in a note to another note or an external resource.
type Link struct {
	// Label of the link.
	Title string
	// Destination URI of the link.
	Href string
	// Indicates whether the target is a remote (e.g. HTTP) resource.
	IsExternal bool
	// Relationships between the note and the linked target.
	Rels []LinkRelation
	// Excerpt of the paragraph containing the note.
	Snippet string
	// Start byte offset of the snippet in the note content.
	SnippetStart int
	// End byte offset of the snippet in the note content.
	SnippetEnd int
}

// LinkRelation defines the relationship between a link's source and target.
type LinkRelation string

const (
	// LinkRelationDown defines the target note as a child of the source.
	LinkRelationDown LinkRelation = "down"
	// LinkRelationDown defines the target note as a parent of the source.
	LinkRelationUp LinkRelation = "up"
)

// LinkRels creates a slice of LinkRelation from a list of strings.
func LinkRels(rel ...string) []LinkRelation {
	rels := []LinkRelation{}
	for _, r := range rel {
		rels = append(rels, LinkRelation(r))
	}
	return rels
}
