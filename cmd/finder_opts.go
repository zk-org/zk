package cmd

import (
	"strconv"
	"time"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/tj/go-naturaldate"
)

// Filtering holds filtering options to select notes.
type Filtering struct {
	Path []string `group:filter arg optional placeholder:PATH help:"Find notes matching the given path, including its descendants."`

	Interactive    bool     `group:filter short:i                     help:"Select notes interactively with fzf."`
	Limit          int      `group:filter short:n   placeholder:COUNT help:"Limit the number of notes found."`
	Match          string   `group:filter short:m   placeholder:QUERY help:"Terms to search for in the notes."`
	Exclude        []string `group:filter short:x   placeholder:PATH  help:"Ignore notes matching the given path, including its descendants."`
	Orphan         bool     `group:filter                             help:"Find notes which are not linked by any other note."   xor:link`
	LinkedBy       []string `group:filter short:l   placeholder:PATH  help:"Find notes which are linked by the given ones."       xor:link`
	LinkingTo      []string `group:filter short:L   placeholder:PATH  help:"Find notes which are linking to the given ones."      xor:link`
	NotLinkedBy    []string `group:filter           placeholder:PATH  help:"Find notes which are not linked by the given ones."   xor:link`
	NotLinkingTo   []string `group:filter           placeholder:PATH  help:"Find notes which are not linking to the given notes." xor:link`
	Related        []string `group:filter           placeholder:PATH  help:"Find notes which might be related to the given ones." xor:link`
	MaxDistance    int      `group:filter           placeholder:COUNT help:"Maximum distance between two linked notes."`
	Recursive      bool     `group:filter short:r                     help:"Follow links recursively."`
	Created        string   `group:filter           placeholder:DATE  help:"Find notes created on the given date."`
	CreatedBefore  string   `group:filter           placeholder:DATE  help:"Find notes created before the given date."`
	CreatedAfter   string   `group:filter           placeholder:DATE  help:"Find notes created after the given date."`
	Modified       string   `group:filter           placeholder:DATE  help:"Find notes modified on the given date."`
	ModifiedBefore string   `group:filter           placeholder:DATE  help:"Find notes modified before the given date."`
	ModifiedAfter  string   `group:filter           placeholder:DATE  help:"Find notes modified after the given date."`
}

// Sorting holds sorting options to order notes.
type Sorting struct {
	Sort []string `group:sort short:s placeholder:TERM help:"Order the notes by the given criterion."`
}

// NewFinderOpts creates an instance of note.FinderOpts from a set of user flags.
func NewFinderOpts(zk *zk.Zk, filtering Filtering, sorting Sorting) (*note.FinderOpts, error) {
	filters := make([]note.Filter, 0)

	paths, ok := relPaths(zk, filtering.Path)
	if ok {
		filters = append(filters, note.PathFilter(paths))
	}

	excludePaths, ok := relPaths(zk, filtering.Exclude)
	if ok {
		filters = append(filters, note.ExcludePathFilter(excludePaths))
	}

	if filtering.Match != "" {
		filters = append(filters, note.MatchFilter(filtering.Match))
	}

	if filtering.Created != "" {
		date, err := parseDate(filtering.Created)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateCreated,
			Direction: note.DateOn,
		})
	}

	if filtering.CreatedBefore != "" {
		date, err := parseDate(filtering.CreatedBefore)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateCreated,
			Direction: note.DateBefore,
		})
	}

	if filtering.CreatedAfter != "" {
		date, err := parseDate(filtering.CreatedAfter)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateCreated,
			Direction: note.DateAfter,
		})
	}

	if filtering.Modified != "" {
		date, err := parseDate(filtering.Modified)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateModified,
			Direction: note.DateOn,
		})
	}

	if filtering.ModifiedBefore != "" {
		date, err := parseDate(filtering.ModifiedBefore)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateModified,
			Direction: note.DateBefore,
		})
	}

	if filtering.ModifiedAfter != "" {
		date, err := parseDate(filtering.ModifiedAfter)
		if err != nil {
			return nil, err
		}
		filters = append(filters, note.DateFilter{
			Date:      date,
			Field:     note.DateModified,
			Direction: note.DateAfter,
		})
	}

	linkedByPaths, ok := relPaths(zk, filtering.LinkedBy)
	if ok {
		filters = append(filters, note.LinkedByFilter{
			Paths:       linkedByPaths,
			Negate:      false,
			Recursive:   filtering.Recursive,
			MaxDistance: filtering.MaxDistance,
		})
	}

	linkingToPaths, ok := relPaths(zk, filtering.LinkingTo)
	if ok {
		filters = append(filters, note.LinkingToFilter{
			Paths:       linkingToPaths,
			Negate:      false,
			Recursive:   filtering.Recursive,
			MaxDistance: filtering.MaxDistance,
		})
	}

	notLinkedByPaths, ok := relPaths(zk, filtering.NotLinkedBy)
	if ok {
		filters = append(filters, note.LinkedByFilter{
			Paths:  notLinkedByPaths,
			Negate: true,
		})
	}

	notLinkingToPaths, ok := relPaths(zk, filtering.NotLinkingTo)
	if ok {
		filters = append(filters, note.LinkingToFilter{
			Paths:  notLinkingToPaths,
			Negate: true,
		})
	}

	relatedPaths, ok := relPaths(zk, filtering.Related)
	if ok {
		filters = append(filters, note.RelatedFilter(relatedPaths))
	}

	if filtering.Orphan {
		filters = append(filters, note.OrphanFilter{})
	}

	if filtering.Interactive {
		filters = append(filters, note.InteractiveFilter(true))
	}

	sorters, err := note.SortersFromStrings(sorting.Sort)
	if err != nil {
		return nil, err
	}

	return &note.FinderOpts{
		Filters: filters,
		Sorters: sorters,
		Limit:   filtering.Limit,
	}, nil
}

func relPaths(zk *zk.Zk, paths []string) ([]string, bool) {
	relPaths := make([]string, 0)
	for _, p := range paths {
		path, err := zk.RelPath(p)
		if err == nil {
			relPaths = append(relPaths, path)
		}
	}
	return relPaths, len(relPaths) > 0
}

func parseDate(date string) (time.Time, error) {
	if i, err := strconv.ParseInt(date, 10, 0); err == nil && i >= 1000 && i < 5000 {
		return time.Date(int(i), time.January, 0, 0, 0, 0, 0, time.UTC), nil
	}
	return naturaldate.Parse(date, time.Now().UTC(), naturaldate.WithDirection(naturaldate.Past))
}
