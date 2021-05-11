package cli

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"
	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/internal/core"
	dateutil "github.com/mickael-menu/zk/internal/util/date"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/strings"
)

// Filtering holds filtering options to select notes.
type Filtering struct {
	Path []string `group:filter arg optional placeholder:PATH help:"Find notes matching the given path, including its descendants."`

	Interactive    bool     `group:filter short:i                     help:"Select notes interactively with fzf."`
	Limit          int      `group:filter short:n   placeholder:COUNT help:"Limit the number of notes found."`
	Match          string   `group:filter short:m   placeholder:QUERY help:"Terms to search for in the notes."`
	ExactMatch     bool     `group:filter short:e                     help:"Search for exact occurrences of the --match argument (case insensitive)."`
	Exclude        []string `group:filter short:x   placeholder:PATH  help:"Ignore notes matching the given path, including its descendants."`
	Tag            []string `group:filter short:t                     help:"Find notes tagged with the given tags."`
	Mention        []string `group:filter           placeholder:PATH  help:"Find notes mentioning the title of the given ones."`
	MentionedBy    []string `group:filter           placeholder:PATH  help:"Find notes whose title is mentioned in the given ones."`
	LinkTo         []string `group:filter short:l   placeholder:PATH  help:"Find notes which are linking to the given ones."`
	NoLinkTo       []string `group:filter           placeholder:PATH  help:"Find notes which are not linking to the given notes."`
	LinkedBy       []string `group:filter short:L   placeholder:PATH  help:"Find notes which are linked by the given ones."`
	NoLinkedBy     []string `group:filter           placeholder:PATH  help:"Find notes which are not linked by the given ones."`
	Orphan         bool     `group:filter                             help:"Find notes which are not linked by any other note."`
	Related        []string `group:filter           placeholder:PATH  help:"Find notes which might be related to the given ones."`
	MaxDistance    int      `group:filter           placeholder:COUNT help:"Maximum distance between two linked notes."`
	Recursive      bool     `group:filter short:r                     help:"Follow links recursively."`
	Created        string   `group:filter           placeholder:DATE  help:"Find notes created on the given date."`
	CreatedBefore  string   `group:filter           placeholder:DATE  help:"Find notes created before the given date."`
	CreatedAfter   string   `group:filter           placeholder:DATE  help:"Find notes created after the given date."`
	Modified       string   `group:filter           placeholder:DATE  help:"Find notes modified on the given date."`
	ModifiedBefore string   `group:filter           placeholder:DATE  help:"Find notes modified before the given date."`
	ModifiedAfter  string   `group:filter           placeholder:DATE  help:"Find notes modified after the given date."`

	Sort []string `group:sort short:s placeholder:TERM help:"Order the notes by the given criterion."`
}

// ExpandNamedFilters expands recursively any named filter found in the Path field.
func (f Filtering) ExpandNamedFilters(filters map[string]string, expandedFilters []string) (Filtering, error) {
	actualPaths := []string{}

	for _, path := range f.Path {
		if filter, ok := filters[path]; ok && !strings.InList(expandedFilters, path) {
			wrap := errors.Wrapperf("failed to expand named filter `%v`", path)

			var parsedFilter Filtering
			parser, err := kong.New(&parsedFilter)
			if err != nil {
				return f, wrap(err)
			}
			args, err := shellquote.Split(filter)
			if err != nil {
				return f, wrap(err)
			}
			_, err = parser.Parse(args)
			if err != nil {
				return f, wrap(err)
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

			if f.Match == "" {
				f.Match = parsedFilter.Match
			} else if parsedFilter.Match != "" {
				f.Match = fmt.Sprintf("(%s) AND (%s)", f.Match, parsedFilter.Match)
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

	opts.Match = opt.NewNotEmptyString(f.Match)
	opts.ExactMatch = f.ExactMatch

	if paths, ok := relPaths(notebook, f.Path); ok {
		opts.IncludePaths = paths
	}

	if paths, ok := relPaths(notebook, f.Exclude); ok {
		opts.ExcludePaths = paths
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

	if paths, ok := relPaths(notebook, f.LinkedBy); ok {
		opts.LinkedBy = &core.LinkFilter{
			Paths:       paths,
			Negate:      false,
			Recursive:   f.Recursive,
			MaxDistance: f.MaxDistance,
		}
	} else if paths, ok := relPaths(notebook, f.NoLinkedBy); ok {
		opts.LinkedBy = &core.LinkFilter{
			Paths:  paths,
			Negate: true,
		}
	}

	if paths, ok := relPaths(notebook, f.LinkTo); ok {
		opts.LinkTo = &core.LinkFilter{
			Paths:       paths,
			Negate:      false,
			Recursive:   f.Recursive,
			MaxDistance: f.MaxDistance,
		}
	} else if paths, ok := relPaths(notebook, f.NoLinkTo); ok {
		opts.LinkTo = &core.LinkFilter{
			Paths:  paths,
			Negate: true,
		}
	}

	if paths, ok := relPaths(notebook, f.Related); ok {
		opts.Related = paths
	}

	opts.Orphan = f.Orphan

	if f.Created != "" {
		start, end, err := parseDayRange(f.Created)
		if err != nil {
			return opts, err
		}
		opts.CreatedStart = &start
		opts.CreatedEnd = &end
	} else {
		if f.CreatedBefore != "" {
			date, err := dateutil.TimeFromNatural(f.CreatedBefore)
			if err != nil {
				return opts, err
			}
			opts.CreatedEnd = &date
		}
		if f.CreatedAfter != "" {
			date, err := dateutil.TimeFromNatural(f.CreatedAfter)
			if err != nil {
				return opts, err
			}
			opts.CreatedStart = &date
		}
	}

	if f.Modified != "" {
		start, end, err := parseDayRange(f.Modified)
		if err != nil {
			return opts, err
		}
		opts.ModifiedStart = &start
		opts.ModifiedEnd = &end
	} else {
		if f.ModifiedBefore != "" {
			date, err := dateutil.TimeFromNatural(f.ModifiedBefore)
			if err != nil {
				return opts, err
			}
			opts.ModifiedEnd = &date
		}
		if f.ModifiedAfter != "" {
			date, err := dateutil.TimeFromNatural(f.ModifiedAfter)
			if err != nil {
				return opts, err
			}
			opts.ModifiedStart = &date
		}
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

func parseDayRange(date string) (start time.Time, end time.Time, err error) {
	day, err := dateutil.TimeFromNatural(date)
	if err != nil {
		return
	}

	start = startOfDay(day)
	end = start.AddDate(0, 0, 1)
	return start, end, nil
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
