package note

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Finder retrieves notes matching the given options.
//
// Returns the number of matches found.
type Finder interface {
	Find(opts FinderOpts) ([]Match, error)
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
	// Snippets are relevant excerpts in the note.
	Snippets []string
}

// Filter is a sealed interface implemented by Finder filter criteria.
type Filter interface{ sealed() }

// MatchFilter is a note filter used to match its content with FTS predicates.
type MatchFilter string

// PathFilter is a note filter using path globs to match notes.
type PathFilter []string

// ExcludePathFilter is a note filter using path globs to exclude notes from the list.
type ExcludePathFilter []string

// LinkedByFilter is a note filter used to select notes being linked by another one.
type LinkedByFilter struct {
	Paths       []string
	Negate      bool
	Recursive   bool
	MaxDistance int
}

// LinkingToFilter is a note filter used to select notes being linked by another one.
type LinkingToFilter struct {
	Paths       []string
	Negate      bool
	Recursive   bool
	MaxDistance int
}

// RelatedFilter is a note filter used to select notes which could might be
// related to the given notes.
type RelatedFilter []string

// OrphanFilter is a note filter used to select notes having no other notes linking to them.
type OrphanFilter struct{}

// DateFilter can be used to filter notes created or modified before, after or on a given date.
type DateFilter struct {
	Date      time.Time
	Direction DateDirection
	Field     DateField
}

// InteractiveFilter lets the user select manually the notes.
type InteractiveFilter bool

func (f MatchFilter) sealed()       {}
func (f PathFilter) sealed()        {}
func (f ExcludePathFilter) sealed() {}
func (f LinkedByFilter) sealed()    {}
func (f LinkingToFilter) sealed()   {}
func (f RelatedFilter) sealed()     {}
func (f OrphanFilter) sealed()      {}
func (f DateFilter) sealed()        {}
func (f InteractiveFilter) sealed() {}

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

// SorterFromString returns a Sorter from its string representation.
//
// If the input str has for suffix `+`, then the order will be ascending, while
// descending for `-`. If no suffix is given, then the default order for the
// sorting field will be used.
func SorterFromString(str string) (Sorter, error) {
	orderSymbol, _ := utf8.DecodeLastRuneInString(str)
	str = strings.TrimRight(str, "+-")

	var sorter Sorter
	switch str {
	case "created", "c":
		sorter = Sorter{Field: SortCreated, Ascending: false}
	case "modified", "m":
		sorter = Sorter{Field: SortModified, Ascending: false}
	case "path", "p":
		sorter = Sorter{Field: SortPath, Ascending: true}
	case "title", "t":
		sorter = Sorter{Field: SortTitle, Ascending: true}
	case "random", "r":
		sorter = Sorter{Field: SortRandom, Ascending: true}
	case "word-count", "wc":
		sorter = Sorter{Field: SortWordCount, Ascending: true}
	default:
		return sorter, fmt.Errorf("%s: unknown sorting term", str)
	}

	switch orderSymbol {
	case '+':
		sorter.Ascending = true
	case '-':
		sorter.Ascending = false
	}

	return sorter, nil
}

// SortersFromStrings returns a list of Sorter from their string representation.
func SortersFromStrings(strs []string) ([]Sorter, error) {
	sorters := make([]Sorter, 0)
	for _, str := range strs {
		sorter, err := SorterFromString(str)
		if err != nil {
			return sorters, err
		}
		sorters = append(sorters, sorter)
	}
	return sorters, nil
}
