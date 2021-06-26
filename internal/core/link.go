package core

// Link represents a link in a note to another note or an external resource.
type Link struct {
	// Label of the link.
	Title string `json:"title"`
	// Destination URI of the link.
	Href string `json:"href"`
	// Indicates whether the target is a remote (e.g. HTTP) resource.
	IsExternal bool `json:"isExternal"`
	// Relationships between the note and the linked target.
	Rels []LinkRelation `json:"rels"`
	// Excerpt of the paragraph containing the note.
	Snippet string `json:"snippet"`
	// Start byte offset of the snippet in the note content.
	SnippetStart int `json:"snippetStart"`
	// End byte offset of the snippet in the note content.
	SnippetEnd int `json:"snippetEnd"`
}

type ResolvedLink struct {
	Link
	SourceID   NoteID `json:"sourceId"`
	SourcePath string `json:"sourcePath"`
	TargetID   NoteID `json:"targetId"`
	TargetPath string `json:"targetPath"`
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
