package sqlite

import (
	"testing"
	"time"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNoteIndexFindByHrefPrefixBug(t *testing.T) {
	db, index := testNoteIndex(t)
	defer db.Close()

	// Create notes matching the ACTUAL scenario:
	// Files: journal/2024-08-27.md and journal/2024-08-27_ajct.md
	// Link: [text](journal/2024-08-27) - no .md extension
	shorterNote := core.Note{
		Path:     "journal/2024-08-27.md",
		Title:    "Shorter note",
		Body:     "This is the shorter note",
		Modified: time.Date(2024, 8, 27, 10, 0, 0, 0, time.UTC),
	}
	longerNote := core.Note{
		Path:     "journal/2024-08-27_ajct.md",
		Title:    "Longer note with suffix",
		Body:     "This is the longer note with suffix",
		Modified: time.Date(2024, 8, 27, 11, 0, 0, 0, time.UTC),
	}

	// Add notes to the index
	shorterId, err := index.Add(shorterNote)
	assert.Nil(t, err)
	longerId, err := index.Add(longerNote)
	assert.Nil(t, err)

	t.Logf("Shorter note ID: %d, Longer note ID: %d", shorterId, longerId)

	// Test the ACTUAL scenario: markdown link without extension
	// Link: [text](journal/2024-08-27) should find journal/2024-08-27.md

	// First try exact matching (what LSP does for markdown links initially)
	exactNote, err := index.FindMinimal(core.NoteFindOpts{
		IncludeHrefs:      []string{"journal/2024-08-27"},
		AllowPartialHrefs: false,
	})
	assert.Nil(t, err)

	t.Logf("Exact search for 'journal/2024-08-27' returned %d results:", len(exactNote))
	for i, n := range exactNote {
		t.Logf("  %d. %s", i, n.Path)
	}

	// This will likely fail because there's no exact file named "journal/2024-08-27"
	if len(exactNote) == 0 {
		t.Logf("No exact match found for 'journal/2024-08-27' (expected)")
	} else if exactNote[0].Path != "journal/2024-08-27.md" {
		t.Errorf("Exact matching: Expected 'journal/2024-08-27.md' but got '%s'", exactNote[0].Path)
	}

	// If exact matching fails, +	// markdown links should resolve to the right file.

	partialNote, err := index.FindMinimal(core.NoteFindOpts{
		IncludeHrefs:      []string{"journal/2024-08-27"},
		AllowPartialHrefs: true,
	})
	assert.Nil(t, err)
	assert.NotEqual(t, len(partialNote), 0)

	t.Logf("Partial search for 'journal/2024-08-27' returned %d results:", len(partialNote))
	for i, n := range partialNote {
		t.Logf("  %d. %s", i, n.Path)
	}

	// Partial matching should find the shorter file first
	if partialNote[0].Path != "journal/2024-08-27.md" {
		t.Errorf("Expected 'journal/2024-08-27.md' but got '%s' when searching for 'journal/2024-08-27'", partialNote[0].Path)
	}

	// Test with limit=1 (as in notebook.FindByHref)
	singleNote, err := index.FindMinimal(core.NoteFindOpts{
		IncludeHrefs:      []string{"journal/2024-08-27"},
		AllowPartialHrefs: true,
		Limit:             1,
	})
	assert.Nil(t, err)
	assert.NotEqual(t, len(singleNote), 0)

	t.Logf("Limited search for 'journal/2024-08-27' returned 1 result: %s", singleNote[0].Path)

	if singleNote[0].Path != "journal/2024-08-27.md" {
		t.Errorf("Expected 'journal/2024-08-27.md' but got '%s'. This simulates the actual LSP behavior.", singleNote[0].Path)
	}
}