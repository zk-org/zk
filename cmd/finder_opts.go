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
	Path []string `arg optional placeholder:"<glob>"`

	Match          string   `short:"m" placeholder:"<query>" help:"Terms to search for in the notes."`
	Limit          int      `short:"n" placeholder:"<count>" help:"Limit the number of results."`
	Created        string   `          placeholder:"<date>"  help:"Only the notes created on the given date."`
	CreatedBefore  string   `          placeholder:"<date>"  help:"Only the notes created before the given date."`
	CreatedAfter   string   `          placeholder:"<date>"  help:"Only the notes created after the given date."`
	Modified       string   `          placeholder:"<date>"  help:"Only the notes modified on the given date."`
	ModifiedBefore string   `          placeholder:"<date>"  help:"Only the notes modified before the given date."`
	ModifiedAfter  string   `          placeholder:"<date>"  help:"Only the notes modified after the given date."`
	Related        []string `                                help:"Only the notes which might be related to the given notes." xor:"link"`
	LinkedBy       []string `short:"l" placeholder:"<path>"  help:"Only the notes linked by the given notes."                 xor:"link"`
	LinkingTo      []string `short:"L" placeholder:"<path>"  help:"Only the notes linking to the given notes."                xor:"link"`
	NotLinkedBy    []string `          placeholder:"<path>"  help:"Only the notes not linked by the given notes."             xor:"link"`
	NotLinkingTo   []string `          placeholder:"<path>"  help:"Only the notes not linking to the given notes."            xor:"link"`
	MaxDistance    int      `                                help:"Maximum distance between two linked notes."`
	Orphan         bool     `                                help:"Only the notes which don't have any other note linking to them."`
	Exclude        []string `short:"x" placeholder:"<glob>"  help:"Excludes notes matching the given file path pattern from the list."`
	Recursive      bool     `short:"r"                       help:"Follow links recursively."`
	Interactive    bool     `short:"i"                       help:"Further filter the list of notes interactively."`
}

// Sorting holds sorting options to order notes.
type Sorting struct {
	Sort []string `short:"s" placeholder:"<term>" help:"Sort the notes by the given criterion."`
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
