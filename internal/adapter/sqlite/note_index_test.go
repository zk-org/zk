package sqlite

import (
	"fmt"
	"testing"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/test/assert"
)

// FIXME: Missing tests

func TestNoteIndexAddWithLinks(t *testing.T) {
	db, index := testNoteIndex(t)

	id, err := index.Add(core.Note{
		Path: "log/added.md",
		Links: []core.Link{
			{
				Title: "Same dir",
				Href:  "log/2021-01-04",
				Rels:  core.LinkRels("rel-1", "rel-2"),
			},
			{
				Title:        "Relative",
				Href:         "f39c8",
				Snippet:      "[Relative](f39c8) link",
				SnippetStart: 50,
				SnippetEnd:   100,
			},
			{
				Title: "Second is added",
				Href:  "f39c8#anchor",
				Rels:  core.LinkRels("second"),
			},
			{
				Title: "Unknown",
				Href:  "unknown",
			},
			{
				Title:      "URL",
				Href:       "http://example.com",
				IsExternal: true,
				Snippet:    "External [URL](http://example.com)",
			},
		},
	})
	assert.Nil(t, err)

	rows := queryLinkRows(t, db.db, fmt.Sprintf("source_id = %d", id))
	assert.Equal(t, rows, []linkRow{
		{
			SourceID: id,
			TargetID: idPointer(2),
			Title:    "Same dir",
			Href:     "log/2021-01-04",
			Rels:     "\x01rel-1\x01rel-2\x01",
		},
		{
			SourceID:     id,
			TargetID:     idPointer(4),
			Title:        "Relative",
			Href:         "f39c8",
			Rels:         "",
			Snippet:      "[Relative](f39c8) link",
			SnippetStart: 50,
			SnippetEnd:   100,
		},
		{
			SourceID: id,
			TargetID: idPointer(4),
			Title:    "Second is added",
			Href:     "f39c8#anchor",
			Rels:     "\x01second\x01",
		},
		{
			SourceID: id,
			TargetID: nil,
			Title:    "Unknown",
			Href:     "unknown",
			Rels:     "",
		},
		{
			SourceID:   id,
			TargetID:   nil,
			Title:      "URL",
			Href:       "http://example.com",
			IsExternal: true,
			Rels:       "",
			Snippet:    "External [URL](http://example.com)",
		},
	})
}

func TestNoteIndexAddFillsLinksMissingTargetId(t *testing.T) {
	db, index := testNoteIndex(t)

	id, err := index.Add(core.Note{
		Path: "missing_target.md",
	})
	assert.Nil(t, err)

	rows := queryLinkRows(t, db.db, fmt.Sprintf("target_id = %d", id))
	assert.Equal(t, rows, []linkRow{
		{
			SourceID: 3,
			TargetID: &id,
			Title:    "Missing target",
			Href:     "missing",
			Snippet:  "There's a Missing target",
		},
	})
}

func TestNoteIndexUpdateWithLinks(t *testing.T) {
	db, index := testNoteIndex(t)

	links := queryLinkRows(t, db.db, "source_id = 1")
	assert.Equal(t, links, []linkRow{
		{
			SourceID: 1,
			TargetID: idPointer(2),
			Title:    "An internal link",
			Href:     "log/2021-01-04.md",
			Snippet:  "[[An internal link]]",
		},
		{
			SourceID:   1,
			TargetID:   nil,
			Title:      "An external link",
			Href:       "https://domain.com",
			IsExternal: true,
			Snippet:    "[[An external link]]",
		},
	})

	err := index.Update(core.Note{
		Path: "log/2021-01-03.md",
		Links: []core.Link{
			{
				Title:      "A new link",
				Href:       "index",
				Type:       core.LinkTypeWikiLink,
				IsExternal: false,
				Rels:       core.LinkRels("rel"),
				Snippet:    "[[A new link]]",
			},
			{
				Title:      "An external link",
				Href:       "https://domain.com",
				Type:       core.LinkTypeMarkdown,
				IsExternal: true,
				Snippet:    "[[An external link]]",
			},
		},
	})
	assert.Nil(t, err)

	links = queryLinkRows(t, db.db, "source_id = 1")
	assert.Equal(t, links, []linkRow{
		{
			SourceID: 1,
			TargetID: idPointer(3),
			Title:    "A new link",
			Href:     "index",
			Type:     "wiki-link",
			Rels:     "\x01rel\x01",
			Snippet:  "[[A new link]]",
		},
		{
			SourceID:   1,
			TargetID:   nil,
			Title:      "An external link",
			Href:       "https://domain.com",
			Type:       "markdown",
			IsExternal: true,
			Snippet:    "[[An external link]]",
		},
	})
}

func TestNoteIndexAddWithTags(t *testing.T) {
	db, index := testNoteIndex(t)

	assertSQL := func(after bool) {
		assertTagExistsOrNot(t, db, true, "fiction")
		assertTagExistsOrNot(t, db, after, "new-tag")
	}

	assertSQL(false)
	id, err := index.Add(core.Note{
		Path: "log/added.md",
		Tags: []string{"new-tag", "fiction"},
	})
	assert.Nil(t, err)
	assertSQL(true)
	assertTaggedOrNot(t, db, true, id, "new-tag")
	assertTaggedOrNot(t, db, true, id, "fiction")
}

func TestNoteIndexUpdateWithTags(t *testing.T) {
	db, index := testNoteIndex(t)
	id := core.NoteID(1)

	assertSQL := func(after bool) {
		assertTaggedOrNot(t, db, true, id, "fiction")
		assertTaggedOrNot(t, db, after, id, "new-tag")
		assertTaggedOrNot(t, db, after, id, "fantasy")
	}

	assertSQL(false)
	err := index.Update(core.Note{
		Path: "log/2021-01-03.md",
		Tags: []string{"new-tag", "fiction", "fantasy"},
	})
	assert.Nil(t, err)
	assertSQL(true)
}

func testNoteIndex(t *testing.T) (*DB, *NoteIndex) {
	db := testDB(t)
	return db, NewNoteIndex("", db, &util.NullLogger)
}

func assertTagExistsOrNot(t *testing.T, db *DB, shouldExist bool, tag string) {
	assertExistOrNot(t, db, shouldExist, "SELECT id FROM collections WHERE kind = 'tag' AND name = ?", tag)
}

func assertTaggedOrNot(t *testing.T, db *DB, shouldBeTagged bool, noteID core.NoteID, tag string) {
	assertExistOrNot(t, db, shouldBeTagged, "SELECT id FROM notes_collections WHERE note_id = ? AND collection_id IS (SELECT id FROM collections WHERE kind = 'tag' AND name = ?)", noteID, tag)
}
