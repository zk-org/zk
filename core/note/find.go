package note

import (
	"time"
)

// Finder retrieves notes matching the given options.
//
// Returns the number of matches found.
type Finder interface {
	Find(opts FinderOpts, callback func(Match) error) (int, error)
}

// FinderOpts holds the option used to filter and order a list of notes.
type FinderOpts struct {
	Filters []Filter
	Sorters []Sorter
	Limit   int
}

// Match holds information about a note matching the find options.
type Match struct {
	Metadata
	// Snippet is an excerpt of the note.
	Snippet string
}

// Filter is a sealed interface implemented by Finder filter criteria.
type Filter interface{ sealed() }

// MatchFilter is a note filter used to match its content with FTS predicates.
type MatchFilter string

// PathFilter is a note filter using path globs to match notes.
type PathFilter []string

// ExcludePathFilter is a note filter using path globs to exclude notes from the list.
type ExcludePathFilter []string

func (f MatchFilter) sealed()       {}
func (f PathFilter) sealed()        {}
func (f ExcludePathFilter) sealed() {}
func (f DateFilter) sealed()        {}

// DateFilter can be used to filter notes created or modified before, after or on a given date.
type DateFilter struct {
	Date      time.Time
	Direction DateDirection
	Field     DateField
}

type DateDirection int

const (
	DateOn DateDirection = iota + 1
	DateBefore
	DateAfter
)

type DateField int

const (
	DateCreated DateField = iota + 1
	DateModified
)

// Sorter represents an order term used to sort a list of notes.
type Sorter struct {
	Field     SortField
	Ascending bool
}

// SortField represents a note field used to sort a list of notes.
type SortField int

const (
	// Sort by creation date.
	SortCreated SortField = iota + 1
	// Sort by modification date.
	SortModified
	// Sort by the file paths.
	SortPath
	// Sort randomly.
	SortRandom
	// Sort by the note titles.
	SortTitle
	// Sort by the number of words in the note bodies.
	SortWordCount
)
