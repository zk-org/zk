package note

import (
	"fmt"
	"strings"
)

type QueryFilter string

type Match struct {
	ID      int
	Snippet string
	Metadata
}

type Finder interface {
	Find(callback func(Match) error, filters ...Filter) error
}

func List(finder Finder, filters ...Filter) error {
	return finder.Find(func(note Match) error {
		fmt.Printf("%v\n", strings.ReplaceAll(note.Snippet, "\\033", "\033"))
		return nil
	}, filters...)
}

// Filter is a sealed interface implemented by Finder filter criteria.
type Filter interface{ sealed() }

func (f QueryFilter) sealed() {}
