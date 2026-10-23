package cli

import (
	"fmt"
	"slices"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
	dateutil "github.com/zk-org/zk/internal/util/date"
)

// Filtering holds filtering options to select notes.
type Filtering struct {
	Path []string `json:"hrefs"`

	Interactive     bool     `json:"-"`
	Limit           int      `json:"limit"`
	Match           []string `json:"match"`
	MatchStrategy   string   `json:"matchStrategy"`
	Exclude         []string `json:"excludeHrefs"`
	Tag             []string `json:"tags"`
	Mention         []string `json:"mention"`
	MentionedBy     []string `json:"mentionedBy"`
	LinkTo          []string `json:"linkTo"`
	NoLinkTo        []string `json:"-"`
	LinkedBy        []string `json:"linkedBy"`
	NoLinkedBy      []string `json:"-"`
	Orphan          bool     `json:"orphan"`
	Tagless         bool     `json:"tagless"`
	MissingBacklink bool     `json:"missingBacklink"`
	BrokenLinks     bool     `json:"brokenLinks"`
	Related         []string `json:"related"`
	MaxDistance     int      `json:"maxDistance"`
	Recursive       bool     `json:"recursive"`
	Created         string   `json:"created"`
	CreatedBefore   string   `json:"createdBefore"`
	CreatedAfter    string   `json:"createdAfter"`
	Modified        string   `json:"modified"`
	ModifiedBefore  string   `json:"modifiedBefore"`
	ModifiedAfter   string   `json:"modifiedAfter"`

	Sort []string `json:"sort"`

	// Deprecated
	ExactMatch bool `json:"exactMatch"`
}

// Flags describes the filtering options, for a front-end to present and
// fill in. The <path> arguments are declared by each command, as positional
// arguments.
func (f *Filtering) Flags() []spec.Flag {
	return []spec.Flag{
		{Name: "interactive", Short: "i", Group: spec.GroupFilter, Target: &f.Interactive,
			Help: "Select notes interactively with fzf."},
		{Name: "limit", Short: "n", Group: spec.GroupFilter, Placeholder: "COUNT", Target: &f.Limit,
			Help: "Limit the number of notes found."},
		{Name: "match", Short: "m", Group: spec.GroupFilter, Placeholder: "QUERY", Target: &f.Match,
			Help: "Terms to search for in the notes."},
		{Name: "match-strategy", Short: "M", Group: spec.GroupFilter, Placeholder: "STRATEGY", Default: "fts", Target: &f.MatchStrategy,
			Help: "Text matching strategy among: fts, re, exact."},
		{Name: "exclude", Short: "x", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.Exclude,
			Help: "Ignore notes matching the given path, including its descendants."},
		{Name: "tag", Short: "t", Group: spec.GroupFilter, Target: &f.Tag,
			Help: "Find notes tagged with the given tags."},
		{Name: "mention", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.Mention,
			Help: "Find notes mentioning the title of the given ones."},
		{Name: "mentioned-by", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.MentionedBy,
			Help: "Find notes whose title is mentioned in the given ones."},
		{Name: "link-to", Short: "l", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.LinkTo,
			Help: "Find notes which are linking to the given ones."},
		{Name: "no-link-to", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.NoLinkTo,
			Help: "Find notes which are not linking to the given notes."},
		{Name: "linked-by", Short: "L", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.LinkedBy,
			Help: "Find notes which are linked by the given ones."},
		{Name: "no-linked-by", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.NoLinkedBy,
			Help: "Find notes which are not linked by the given ones."},
		{Name: "orphan", Group: spec.GroupFilter, Target: &f.Orphan,
			Help: "Find notes which are not linked by any other note."},
		{Name: "tagless", Group: spec.GroupFilter, Target: &f.Tagless,
			Help: "Find notes which have no tags."},
		{Name: "missing-backlink", Group: spec.GroupFilter, Target: &f.MissingBacklink,
			Help: "Find notes with at least one missing backlink."},
		{Name: "broken-links", Group: spec.GroupFilter, Target: &f.BrokenLinks,
			Help: "Find notes with at least one broken internal link."},
		{Name: "related", Group: spec.GroupFilter, Placeholder: "PATH", Target: &f.Related,
			Help: "Find notes which might be related to the given ones."},
		{Name: "max-distance", Group: spec.GroupFilter, Placeholder: "COUNT", Target: &f.MaxDistance,
			Help: "Maximum distance between two linked notes."},
		{Name: "recursive", Short: "r", Group: spec.GroupFilter, Target: &f.Recursive,
			Help: "Follow links recursively."},
		{Name: "created", Group: spec.GroupFilter, Placeholder: "DATE", Target: &f.Created,
			Help: "Find notes created on the given date."},
		{Name: "created-before", Group: spec.GroupFilter, Placeholder: "DATE", Target: &f.CreatedBefore,
			Help: "Find notes created before the given date."},
		{Name: "created-after", Group: spec.GroupFilter, Placeholder: "DATE", Target: &f.CreatedAfter,
			Help: "Find notes created after the given date."},
		{Name: "modified", Group: spec.GroupFilter, Placeholder: "DATE", Target: &f.Modified,
			Help: "Find notes modified on the given date."},
		{Name: "modified-before", Group: spec.GroupFilter, Placeholder: "DATE", Target: &f.ModifiedBefore,
			Help: "Find notes modified before the given date."},
		{Name: "modified-after", Group: spec.GroupFilter, Placeholder: "DATE", Target: &f.ModifiedAfter,
			Help: "Find notes modified after the given date."},

		{Name: "sort", Short: "s", Group: spec.GroupSort, Placeholder: "TERM", Target: &f.Sort,
			Help: "Order the notes by the given criterion."},

		// Deprecated, kept to report a helpful error.
		{Name: "exact-match", Short: "e", Hidden: true, Target: &f.ExactMatch},
	}
}

// PathArg describes the <path> positional arguments every filtering command
// takes.
func (f *Filtering) PathArg() spec.Arg {
	return spec.Arg{
		Name:     "path",
		Help:     "Find notes matching the given path, including its descendants.",
		Optional: true,
		Variadic: true,
		Target:   &f.Path,
	}
}

// ParseFilter parses a named filter, defined in a notebook's configuration as
// the command line it mimics.
//
// It is a variable so that a front-end can install its own parser and keep a
// notebook's filters reading exactly like the commands it accepts.
var ParseFilter = func(args []string) (Filtering, error) {
	var f Filtering
	rest, err := spec.ParseFlags(f.Flags(), args)
	if err != nil {
		return f, err
	}
	f.Path = append(f.Path, rest...)
	return f, nil
}

// ParseFilterString parses a named filter written as a single string, splitting
// it the way a shell would before handing it to ParseFilter.
func ParseFilterString(filter string) (Filtering, error) {
	args, err := shellquote.Split(filter)
	if err != nil {
		return Filtering{}, err
	}
	return ParseFilter(args)
}

// ExpandNamedFilters expands recursively any named filter found in the Path field.
func (f Filtering) ExpandNamedFilters(filters map[string]string, expandedFilters []string) (Filtering, error) {
	actualPaths := []string{}

	for _, path := range f.Path {
		if filter, ok := filters[path]; ok && !slices.Contains(expandedFilters, path) {
			parsedFilter, err := ParseFilterString(filter)
			if err != nil {
				return f, fmt.Errorf("failed to expand named filter `%v`: %w", path, err)
			}

			// Expand recursively, but prevent infinite loops by registering
			// the current filter in the list of expanded filters.
			parsedFilter, err = parsedFilter.ExpandNamedFilters(filters, append(expandedFilters, path))
			if err != nil {
				return f, err
			}

			actualPaths = append(actualPaths, parsedFilter.Path...)
			f.Exclude = append(f.Exclude, parsedFilter.Exclude...)
			f.Tag = append(f.Tag, parsedFilter.Tag...)
			f.Mention = append(f.Mention, parsedFilter.Mention...)
			f.MentionedBy = append(f.MentionedBy, parsedFilter.MentionedBy...)
			f.LinkTo = append(f.LinkTo, parsedFilter.LinkTo...)
			f.NoLinkTo = append(f.NoLinkTo, parsedFilter.NoLinkTo...)
			f.LinkedBy = append(f.LinkedBy, parsedFilter.LinkedBy...)
			f.NoLinkedBy = append(f.NoLinkedBy, parsedFilter.NoLinkedBy...)
			f.Related = append(f.Related, parsedFilter.Related...)
			f.Sort = append(f.Sort, parsedFilter.Sort...)

			f.ExactMatch = f.ExactMatch || parsedFilter.ExactMatch
			f.Interactive = f.Interactive || parsedFilter.Interactive
			f.Orphan = f.Orphan || parsedFilter.Orphan
			f.Tagless = f.Tagless || parsedFilter.Tagless
			f.MissingBacklink = f.MissingBacklink || parsedFilter.MissingBacklink
			f.BrokenLinks = f.BrokenLinks || parsedFilter.BrokenLinks
			f.Recursive = f.Recursive || parsedFilter.Recursive

			if f.Limit == 0 {
				f.Limit = parsedFilter.Limit
			}
			if f.MaxDistance == 0 {
				f.MaxDistance = parsedFilter.MaxDistance
			}
			if f.Created == "" {
				f.Created = parsedFilter.Created
			}
			if f.CreatedBefore == "" {
				f.CreatedBefore = parsedFilter.CreatedBefore
			}
			if f.CreatedAfter == "" {
				f.CreatedAfter = parsedFilter.CreatedAfter
			}
			if f.Modified == "" {
				f.Modified = parsedFilter.Modified
			}
			if f.ModifiedBefore == "" {
				f.ModifiedBefore = parsedFilter.ModifiedBefore
			}
			if f.ModifiedAfter == "" {
				f.ModifiedAfter = parsedFilter.ModifiedAfter
			}

			f.Match = append(f.Match, parsedFilter.Match...)
			if f.MatchStrategy == "" {
				f.MatchStrategy = parsedFilter.MatchStrategy
			}

		} else {
			actualPaths = append(actualPaths, path)
		}
	}

	f.Path = actualPaths
	return f, nil
}

// NewNoteFindOpts creates an instance of core.NoteFindOpts from a set of user flags.
func (f Filtering) NewNoteFindOpts(notebook *core.Notebook) (core.NoteFindOpts, error) {
	opts := core.NoteFindOpts{}

	f, err := f.ExpandNamedFilters(notebook.Config.Filters, []string{})
	if err != nil {
		return opts, err
	}

	if f.ExactMatch {
		return opts, fmt.Errorf("the --exact-match (-e) option is deprecated, use --match-strategy=exact (-Me) instead")
	}

	opts.Match = make([]string, len(f.Match))
	copy(opts.Match, f.Match)
	opts.MatchStrategy, err = core.MatchStrategyFromString(f.MatchStrategy)
	if err != nil {
		return opts, err
	}

	if paths, ok := relPaths(notebook, f.Path); ok {
		opts.IncludeHrefs = paths
	}

	if paths, ok := relPaths(notebook, f.Exclude); ok {
		opts.ExcludeHrefs = paths
	}

	if len(f.Tag) > 0 {
		opts.Tags = f.Tag
	}

	if len(f.Mention) > 0 {
		opts.Mention = f.Mention
	}

	if len(f.MentionedBy) > 0 {
		opts.MentionedBy = f.MentionedBy
	}

	opts.LinkedBy = linkFilterFromPaths(notebook, f.LinkedBy, f.NoLinkedBy, f.Recursive, f.MaxDistance)
	opts.LinkTo = linkFilterFromPaths(notebook, f.LinkTo, f.NoLinkTo, f.Recursive, f.MaxDistance)

	if paths, ok := relPaths(notebook, f.Related); ok {
		opts.Related = paths
	}

	opts.Orphan = f.Orphan
	opts.Tagless = f.Tagless
	opts.MissingBacklink = f.MissingBacklink
	opts.BrokenLinks = f.BrokenLinks

	opts.CreatedStart, opts.CreatedEnd, err = parseTimeRange(f.Created, f.CreatedBefore, f.CreatedAfter)
	if err != nil {
		return opts, err
	}

	opts.ModifiedStart, opts.ModifiedEnd, err = parseTimeRange(f.Modified, f.ModifiedBefore, f.ModifiedAfter)
	if err != nil {
		return opts, err
	}

	sorters, err := core.NoteSortersFromStrings(f.Sort)
	if err != nil {
		return opts, err
	}
	opts.Sorters = sorters

	opts.Limit = f.Limit

	return opts, nil
}

func relPaths(notebook *core.Notebook, paths []string) ([]string, bool) {
	relPaths := make([]string, 0)
	for _, p := range paths {
		path, err := notebook.RelPath(p)
		if err == nil {
			relPaths = append(relPaths, path)
		}
	}
	return relPaths, len(relPaths) > 0
}

// linkFilterFromPaths builds a LinkFilter from a positive and a negative list of hrefs. Positive items take precedence.
func linkFilterFromPaths(notebook *core.Notebook, hrefs, negatedHrefs []string, recursive bool, maxDistance int) *core.LinkFilter {
	if paths, ok := relPaths(notebook, hrefs); ok {
		return &core.LinkFilter{
			Hrefs:       paths,
			Negate:      false,
			Recursive:   recursive,
			MaxDistance: maxDistance,
		}
	}

	if paths, ok := relPaths(notebook, negatedHrefs); ok {
		return &core.LinkFilter{
			Hrefs:  paths,
			Negate: true,
		}
	}

	return nil
}

func parseDayRange(date string) (start time.Time, end time.Time, err error) {
	day, err := dateutil.TimeFromNatural(date)
	if err != nil {
		return
	}

	// we add -1 second so that the day range ends at 23:59:59
	// i.e, the 'new day' begins at 00:00:00
	start = startOfDay(day).Add(time.Second * -1)
	end = start.AddDate(0, 0, 1)
	return start, end, nil
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// parseTimeRange converts the exact/before/after filter fields into optional start and end timestamps.
// An exact day takes precedence over the before/after fields.
func parseTimeRange(exact, before, after string) (start, end *time.Time, err error) {
	if exact != "" {
		s, e, err := parseDayRange(exact)
		if err != nil {
			return nil, nil, err
		}
		return &s, &e, nil
	}

	if before != "" {
		date, err := dateutil.TimeFromNatural(before)
		if err != nil {
			return nil, nil, err
		}
		end = &date
	}
	if after != "" {
		date, err := dateutil.TimeFromNatural(after)
		if err != nil {
			return nil, nil, err
		}
		start = &date
	}

	return start, end, nil
}
