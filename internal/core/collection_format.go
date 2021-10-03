package core

import (
	"encoding/json"
)

// CollectionFormatter formats collections to be printed on the screen.
type CollectionFormatter func(collection Collection) (string, error)

func newCollectionFormatter(template Template) (CollectionFormatter, error) {
	return func(collection Collection) (string, error) {
		return template.Render(collectionFormatRenderContext{
			ID:        collection.ID,
			Kind:      collection.Kind,
			Name:      collection.Name,
			NoteCount: collection.NoteCount,
		})
	}, nil
}

// collectionFormatRenderContext holds the variables available to the
// collection formatting templates.
type collectionFormatRenderContext struct {
	// Unique ID of this collection in the Notebook.
	ID CollectionID `json:"id"`
	// Kind of this note collection, such as a tag.
	Kind CollectionKind `json:"kind"`
	// Name of this collection.
	Name string `json:"name"`
	// Number of notes associated with this collection.
	NoteCount int `json:"noteCount" handlebars:"note-count"`
}

func (c collectionFormatRenderContext) Equal(other collectionFormatRenderContext) bool {
	json1, err := json.Marshal(c)
	if err != nil {
		return false
	}
	json2, err := json.Marshal(other)
	if err != nil {
		return false
	}
	return string(json1) == string(json2)
}
