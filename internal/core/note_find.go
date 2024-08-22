package core

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// NoteFindOpts holds a set of filtering options used to find notes.
type NoteFindOpts struct {
	// Filter used to match the notes with the given MatchStrategy.
	Match []string
	// Text matching strategy used with Match.
	MatchStrategy MatchStrategy
	// Filter by note hrefs.
	IncludeHrefs []string
	// Filter excluding notes at the given hrefs.
	ExcludeHrefs []string
	// Indicates whether href options can match any portion of a path.
	// This is used for wiki links.
	AllowPartialHrefs bool
	// Filter including notes with the given IDs.
	IncludeIDs []NoteID
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
	// Filter to select notes which could might be related to the given notes hrefs.
	Related []string
	// Filter to select notes having no other notes linking to them.
	Orphan bool
	// Filter to select notes having no tags.
	Tagless bool
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

// IncludingIDs creates a new FinderOpts after adding the given IDs to the list
// of excluded note IDs.
func (o NoteFindOpts) IncludingIDs(ids []NoteID) NoteFindOpts {
	if o.IncludeIDs == nil {
		o.IncludeIDs = []NoteID{}
	}

	o.IncludeIDs = append(o.IncludeIDs, ids...)
	return o
}

// ExcludingIDs creates a new FinderOpts after adding the given IDs to the list
// of excluded note IDs.
func (o NoteFindOpts) ExcludingIDs(ids []NoteID) NoteFindOpts {
	if o.ExcludeIDs == nil {
		o.ExcludeIDs = []NoteID{}
	}

	o.ExcludeIDs = append(o.ExcludeIDs, ids...)
	return o
}

// LinkFilter is a note filter used to select notes linking to other ones.
type LinkFilter struct {
	Hrefs       []string
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

// MatchStrategy represents a text matching strategy used when filtering notes with `--match`.
type MatchStrategy int

const (
	// Full text search.
	MatchStrategyFts MatchStrategy = iota + 1
	// Exact text matching.
	MatchStrategyExact
	// Regular expression.
	MatchStrategyRe
)

// MatchStrategyFromString returns a MatchStrategy from its string representation.
func MatchStrategyFromString(str string) (MatchStrategy, error) {
	switch str {
	case "fts", "f", "":
		return MatchStrategyFts, nil
	case "re", "grep", "r":
		return MatchStrategyRe, nil
	case "exact", "e":
		return MatchStrategyExact, nil
	default:
		return 0, fmt.Errorf("%s: unknown match strategy\ntry fts (full-text search), re (regular expression) or exact", str)
	}
}
