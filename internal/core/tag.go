package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type TaggedNoteDetail struct {
	Tag       string `json:"-"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	Positions []int  `json:"position"`
}

type TaggedNotes struct {
	Tag   string             `json:"tag"`
	Notes []TaggedNoteDetail `json:"notes"`
}

func GroupTaggedNotes(details []TaggedNoteDetail) []TaggedNotes {
	seen := []string{}
	groups := map[string][]TaggedNoteDetail{}

	for _, d := range details {
		if _, exists := groups[d.Tag]; !exists {
			seen = append(seen, d.Tag)
		}
		groups[d.Tag] = append(groups[d.Tag], d)
	}

	result := make([]TaggedNotes, 0, len(seen))
	for _, tag := range seen {
		result = append(result, TaggedNotes{
			Tag:   tag,
			Notes: groups[tag],
		})
	}
	return result
}

type TaggedNoteSorter struct {
	Field     TaggedNoteSortField
	Ascending bool
}

type TaggedNoteSortField int

const (
	// Sort by the tag name.
	TagSortName TaggedNoteSortField = 1 << iota
	// Sort by the title of the note.
	TagSortTitle
)

func TaggedNoteSorterFromStrings(strs []string) ([]TaggedNoteSorter, error) {
	sorters := make([]TaggedNoteSorter, 0)

	for _, sorterString := range strs {
		sorter, err := TaggedNoteSorterFromString(sorterString)
		if err != nil {
			return sorters, err
		}
		sorters = append(sorters, sorter)
	}
	return sorters, nil
}

func TaggedNoteSorterFromString(str string) (TaggedNoteSorter, error) {
	orderSymbol, _ := utf8.DecodeLastRuneInString(str)
	str = strings.TrimRight(str, "+-")

	var sorter TaggedNoteSorter
	switch str {
	case "name", "n":
		sorter = TaggedNoteSorter{Field: TagSortName, Ascending: true}
	case "title", "t":
		sorter = TaggedNoteSorter{Field: TagSortTitle, Ascending: true}
	default:
		return sorter, fmt.Errorf("%s: unknown sorting term\ntry name or title", str)
	}

	switch orderSymbol {
	case '+':
		sorter.Ascending = true
	case '-':
		sorter.Ascending = false
	}

	return sorter, nil
}
