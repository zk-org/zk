package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Collection represents a collection, such as a tag.
type Collection struct {
	// Unique ID of this collection in the Notebook.
	ID CollectionID `json:"id"`
	// Kind of this note collection, such as a tag.
	Kind CollectionKind `json:"kind"`
	// Name of this collection.
	Name string `json:"name"`
	// Number of notes associated with this collection.
	NoteCount int `json:"note_count"`
}

// CollectionID represents the unique ID of a collection relative to a given
// NoteIndex implementation.
type CollectionID int64

func (id CollectionID) IsValid() bool {
	return id > 0
}

// NoteCollectionID represents the unique ID of an association between a note
// and a collection in a NoteIndex implementation.
type NoteCollectionID int64

func (id NoteCollectionID) IsValid() bool {
	return id > 0
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
	// for the given kind, ordered with the given sorters.
	FindCollections(kind CollectionKind, sorters []CollectionSorter) ([]Collection, error)

	// AssociateNoteCollection creates a new association between a note and a
	// collection, if it does not already exist.
	AssociateNoteCollection(noteID NoteID, collectionID CollectionID) (NoteCollectionID, error)

	// RemoveNoteCollections deletes all collection associations with the given
	// note.
	RemoveNoteAssociations(noteID NoteID) error
}

// CollectionSorter represents an order term used to sort a list of collections.
type CollectionSorter struct {
	Field     CollectionSortField
	Ascending bool
}

// CollectionSortField represents a collection field used to sort a list of collections.
type CollectionSortField int

const (
	// Sort by the collection names.
	CollectionSortName CollectionSortField = iota + 1
	// Sort by the number of notes part of the collection.
	CollectionSortNoteCount
)

// CollectionSortersFromStrings returns a list of CollectionSorter from their string
// representation.
func CollectionSortersFromStrings(strs []string) ([]CollectionSorter, error) {
	sorters := make([]CollectionSorter, 0)

	// Iterates in reverse order to be able to override sort criteria set in a
	// config alias with a `--sort` flag.
	for i := len(strs) - 1; i >= 0; i-- {
		sorter, err := CollectionSorterFromString(strs[i])
		if err != nil {
			return sorters, err
		}
		sorters = append(sorters, sorter)
	}
	return sorters, nil
}

// CollectionSorterFromString returns a CollectionSorter from its string representation.
//
// If the input str has for suffix `+`, then the order will be ascending, while
// descending for `-`. If no suffix is given, then the default order for the
// sorting field will be used.
func CollectionSorterFromString(str string) (CollectionSorter, error) {
	orderSymbol, _ := utf8.DecodeLastRuneInString(str)
	str = strings.TrimRight(str, "+-")

	var sorter CollectionSorter
	switch str {
	case "name", "n":
		sorter = CollectionSorter{Field: CollectionSortName, Ascending: true}
	case "note-count", "nc":
		sorter = CollectionSorter{Field: CollectionSortNoteCount, Ascending: false}
	default:
		return sorter, fmt.Errorf("%s: unknown sorting term\ntry name or note-count", str)
	}

	switch orderSymbol {
	case '+':
		sorter.Ascending = true
	case '-':
		sorter.Ascending = false
	}

	return sorter, nil
}
