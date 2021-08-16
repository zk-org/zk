package core

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mickael-menu/zk/internal/util/opt"
)

// NoteFindOpts holds a set of filtering options used to find notes.
type NoteFindOpts struct {
	// Filter used to match the notes with FTS predicates.
	Match opt.String
	// Search for exact occurrences of the Match string.
	ExactMatch bool
	// Filter by note paths.
	IncludePaths []string
	// Filter excluding notes at the given paths.
	ExcludePaths []string
	// Indicates whether IncludePaths and ExcludePaths are using regexes.
	EnablePathRegexes bool
	// Filter excluding notes with the given IDs.
	ExcludeIDs []NoteID
	// Filter by tags found in the notes.
	Tags []string
	// Filter the notes mentioning the given ones.
	Mention []string
	// Filter the notes mentioned by the given ones.
	MentionedBy []string
	// Filter to select notes being linked by another one.
	LinkedBy *LinkFilter
	// Filter to select notes linking to another one.
	LinkTo *LinkFilter
	// Filter to select notes which could might be related to the given notes paths.
	Related []string
	// Filter to select notes having no other notes linking to them.
	Orphan bool
	// Filter notes created after the given date.
	CreatedStart *time.Time
	// Filter notes created before the given date.
	CreatedEnd *time.Time
	// Filter notes modified after the given date.
	ModifiedStart *time.Time
	// Filter notes modified before the given date.
	ModifiedEnd *time.Time
	// Limits the number of results
	Limit int
	// Sorting criteria
	Sorters []NoteSorter
}

// ExcludingID creates a new FinderOpts after adding the given ID to the list
// of excluded note IDs.
func (o NoteFindOpts) ExcludingID(id NoteID) NoteFindOpts {
	if o.ExcludeIDs == nil {
		o.ExcludeIDs = []NoteID{}
	}

	o.ExcludeIDs = append(o.ExcludeIDs, id)
	return o
}

// LinkFilter is a note filter used to select notes linking to other ones.
type LinkFilter struct {
	Paths       []string
	Negate      bool
	Recursive   bool
	MaxDistance int
}

// NoteSorter represents an order term used to sort a list of notes.
type NoteSorter struct {
	Field     NoteSortField
	Ascending bool
}

// NoteSortField represents a note field used to sort a list of notes.
type NoteSortField int

const (
	// Sort by creation date.
	NoteSortCreated NoteSortField = iota + 1
	// Sort by modification date.
	NoteSortModified
	// Sort by the file paths.
	NoteSortPath
	// Sort randomly.
	NoteSortRandom
	// Sort by the note titles.
	NoteSortTitle
	// Sort by the number of words in the note bodies.
	NoteSortWordCount
	// Sort by the length of the note path.
	// This is not accessible to the user but used for technical reasons, to
	// find the best match when searching a path prefix.
	NoteSortPathLength
)

// NoteSortersFromStrings returns a list of NoteSorter from their string
// representation.
func NoteSortersFromStrings(strs []string) ([]NoteSorter, error) {
	sorters := make([]NoteSorter, 0)

	// Iterates in reverse order to be able to override sort criteria set in a
	// config alias with a `--sort` flag.
	for i := len(strs) - 1; i >= 0; i-- {
		sorter, err := NoteSorterFromString(strs[i])
		if err != nil {
			return sorters, err
		}
		sorters = append(sorters, sorter)
	}
	return sorters, nil
}

// NoteSorterFromString returns a NoteSorter from its string representation.
//
// If the input str has for suffix `+`, then the order will be ascending, while
// descending for `-`. If no suffix is given, then the default order for the
// sorting field will be used.
func NoteSorterFromString(str string) (NoteSorter, error) {
	orderSymbol, _ := utf8.DecodeLastRuneInString(str)
	str = strings.TrimRight(str, "+-")

	var sorter NoteSorter
	switch str {
	case "created", "c":
		sorter = NoteSorter{Field: NoteSortCreated, Ascending: false}
	case "modified", "m":
		sorter = NoteSorter{Field: NoteSortModified, Ascending: false}
	case "path", "p":
		sorter = NoteSorter{Field: NoteSortPath, Ascending: true}
	case "title", "t":
		sorter = NoteSorter{Field: NoteSortTitle, Ascending: true}
	case "random", "r":
		sorter = NoteSorter{Field: NoteSortRandom, Ascending: true}
	case "word-count", "wc":
		sorter = NoteSorter{Field: NoteSortWordCount, Ascending: true}
	default:
		return sorter, fmt.Errorf("%s: unknown sorting term\ntry created, modified, path, title, random or word-count", str)
	}

	switch orderSymbol {
	case '+':
		sorter.Ascending = true
	case '-':
		sorter.Ascending = false
	}

	return sorter, nil
}
