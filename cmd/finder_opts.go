package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/tj/go-naturaldate"
)

// Filtering holds filtering options to select notes.
type Filtering struct {
	Path []string `group:filter arg optional placeholder:PATH help:"Find notes matching the given path, including its descendants."`

	Interactive    bool     `group:filter short:i                     help:"Select notes interactively with fzf."`
	Limit          int      `group:filter short:n   placeholder:COUNT help:"Limit the number of notes found."`
	Match          string   `group:filter short:m   placeholder:QUERY help:"Terms to search for in the notes."`
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
		if filter, ok := filters[path]; ok {
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

// NewFinderOpts creates an instance of note.FinderOpts from a set of user flags.
func NewFinderOpts(zk *zk.Zk, filtering Filtering) (*note.FinderOpts, error) {
	opts := note.FinderOpts{}

	opts.Match = opt.NewNotEmptyString(filtering.Match)

	if paths, ok := relPaths(zk, filtering.Path); ok {
		opts.IncludePaths = paths
	}

	if paths, ok := relPaths(zk, filtering.Exclude); ok {
		opts.ExcludePaths = paths
	}

	if len(filtering.Tag) > 0 {
		opts.Tags = filtering.Tag
	}

	if len(filtering.Mention) > 0 {
		opts.Mention = filtering.Mention
	}

	if len(filtering.MentionedBy) > 0 {
		opts.MentionedBy = filtering.MentionedBy
	}

	if paths, ok := relPaths(zk, filtering.LinkedBy); ok {
		opts.LinkedBy = &note.LinkFilter{
			Paths:       paths,
			Negate:      false,
			Recursive:   filtering.Recursive,
			MaxDistance: filtering.MaxDistance,
		}
	} else if paths, ok := relPaths(zk, filtering.NoLinkedBy); ok {
		opts.LinkedBy = &note.LinkFilter{
			Paths:  paths,
			Negate: true,
		}
	}

	if paths, ok := relPaths(zk, filtering.LinkTo); ok {
		opts.LinkTo = &note.LinkFilter{
			Paths:       paths,
			Negate:      false,
			Recursive:   filtering.Recursive,
			MaxDistance: filtering.MaxDistance,
		}
	} else if paths, ok := relPaths(zk, filtering.NoLinkTo); ok {
		opts.LinkTo = &note.LinkFilter{
			Paths:  paths,
			Negate: true,
		}
	}

	if paths, ok := relPaths(zk, filtering.Related); ok {
		opts.Related = paths
	}

	opts.Orphan = filtering.Orphan

	if filtering.Created != "" {
		start, end, err := parseDayRange(filtering.Created)
		if err != nil {
			return nil, err
		}
		opts.CreatedStart = &start
		opts.CreatedEnd = &end
	} else {
		if filtering.CreatedBefore != "" {
			date, err := parseDate(filtering.CreatedBefore)
			if err != nil {
				return nil, err
			}
			opts.CreatedEnd = &date
		}
		if filtering.CreatedAfter != "" {
			date, err := parseDate(filtering.CreatedAfter)
			if err != nil {
				return nil, err
			}
			opts.CreatedStart = &date
		}
	}

	if filtering.Modified != "" {
		start, end, err := parseDayRange(filtering.Modified)
		if err != nil {
			return nil, err
		}
		opts.ModifiedStart = &start
		opts.ModifiedEnd = &end
	} else {
		if filtering.ModifiedBefore != "" {
			date, err := parseDate(filtering.ModifiedBefore)
			if err != nil {
				return nil, err
			}
			opts.ModifiedEnd = &date
		}
		if filtering.ModifiedAfter != "" {
			date, err := parseDate(filtering.ModifiedAfter)
			if err != nil {
				return nil, err
			}
			opts.ModifiedStart = &date
		}
	}

	opts.Interactive = filtering.Interactive

	sorters, err := note.SortersFromStrings(filtering.Sort)
	if err != nil {
		return nil, err
	}
	opts.Sorters = sorters

	opts.Limit = filtering.Limit

	return &opts, nil
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

func parseDayRange(date string) (start time.Time, end time.Time, err error) {
	day, err := parseDate(date)
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
