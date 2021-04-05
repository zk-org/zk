package core

// Collection represents a collection, such as a tag.
type Collection struct {
	// Unique ID of this collection in the CollectionRepository.
	ID CollectionID
	// Kind of this note collection, such as a tag.
	Kind CollectionKind
	// Name of this collection.
	Name string
	// Number of notes associated with this collection.
	NoteCount int
}

// CollectionID represents the unique ID of a collection relative to a given
// CollectionRepository implementation.
type CollectionID interface {
	IsValid() bool
}

// NoteCollectionID represents the unique ID of an association between a note
// and a collection in a CollectionRepository implementation.
type NoteCollectionID interface {
	IsValid() bool
}

// CollectionKind defines a kind of note collection, such as tags.
type CollectionKind string

const (
	CollectionKindTag CollectionKind = "tag"
)

// CollectionRepository persists note collection across sessions.
type CollectionRepository interface {

	// FindOrCreate returns the ID of the collection with given kind and name.
	// If the collection does not exist, creates a new one.
	FindOrCreateCollection(name string, kind CollectionKind) (CollectionID, error)

	// FindCollections returns the list of all collections in the repository
	// for the given kind.
	FindCollections(kind CollectionKind) ([]Collection, error)

	// AssociateNoteCollection creates a new association between a note and a
	// collection, if it does not already exist.
	AssociateNoteCollection(noteID NoteID, collectionID CollectionID) (NoteCollectionID, error)

	// RemoveNoteCollections deletes all collection associations with the given
	// note.
	RemoveNoteAssociations(noteId NoteID) error
}
