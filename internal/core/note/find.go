package note

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/opt"
)

// Finder retrieves notes matching the given options.
type Finder interface {
	Find(opts FinderOpts) ([]Match, error)
	FindByHref(href string) (*Match, error)
}

// FinderOpts holds the option used to filter and order a list of notes.
type FinderOpts struct {
	// Filter used to match the notes with FTS predicates.
	Match opt.String
	// Filter by note paths.
	IncludePaths []string
	// Filter excluding notes at the given paths.
	ExcludePaths []string
	// Filter excluding notes with the given IDs.
	ExcludeIds []core.NoteID
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
	// Indicates that the user should select manually the notes.
	Interactive bool
	// Limits the number of results
	Limit int
	// Sorting criteria
	Sorters []Sorter
}

// ExcludingId creates a new FinderOpts after adding the given id to the list
// of excluded note ids.
func (o FinderOpts) ExcludingId(id core.NoteID) FinderOpts {
	if o.ExcludeIds == nil {
		o.ExcludeIds = []core.NoteID{}
	}

	o.ExcludeIds = append(o.ExcludeIds, id)
	return o
}

// Match holds information about a note matching the find options.
type Match struct {
	Metadata
	// Snippets are relevant excerpts in the note.
	Snippets []string
}

// LinkFilter is a note filter used to select notes linking to other ones.
type LinkFilter struct {
	Paths       []string
	Negate      bool
	Recursive   bool
	MaxDistance int
}

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

// SortersFromStrings returns a list of Sorter from their string representation.
func SortersFromStrings(strs []string) ([]Sorter, error) {
	sorters := make([]Sorter, 0)

	// Iterates in reverse order to be able to override sort criteria set in a
	// config alias with a `--sort` flag.
	for i := len(strs) - 1; i >= 0; i-- {
		sorter, err := SorterFromString(strs[i])
		if err != nil {
			return sorters, err
		}
		sorters = append(sorters, sorter)
	}
	return sorters, nil
}
