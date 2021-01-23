package fzf

import (
	"fmt"
	"io"
	"strings"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/style"
	stringsutil "github.com/mickael-menu/zk/util/strings"
)

// NoteFinder wraps a note.Finder and filters its result interactively using fzf.
type NoteFinder struct {
	finder note.Finder
	styler style.Styler
}

func NewNoteFinder(finder note.Finder, styler style.Styler) *NoteFinder {
	return &NoteFinder{finder, styler}
}

func (f *NoteFinder) Find(opts note.FinderOpts) ([]note.Match, error) {
	isInteractive, opts := popInteractiveFilter(opts)
	matches, err := f.finder.Find(opts)

	if !isInteractive || err != nil {
		return matches, err
	}

	selectedMatches := make([]note.Match, 0)

	selection, err := withFzf(func(fzf io.Writer) error {
		for _, match := range matches {
			fmt.Fprintf(fzf, "%v\x01  %v  %v\n",
				match.Path,
				f.styler.MustStyle(match.Title, style.Rule("yellow")),
				f.styler.MustStyle(stringsutil.JoinLines(match.Body), style.Rule("faint")),
			)
		}
		return nil
	})
	if err != nil {
		return selectedMatches, err
	}

	for _, s := range selection {
		path := strings.Split(s, "\x01")[0]
		for _, m := range matches {
			if m.Path == path {
				selectedMatches = append(selectedMatches, m)
			}
		}
	}

	return selectedMatches, nil
}

func popInteractiveFilter(opts note.FinderOpts) (bool, note.FinderOpts) {
	isInteractive := false
	filters := make([]note.Filter, 0)

	for _, filter := range opts.Filters {
		if f, ok := filter.(note.InteractiveFilter); ok {
			isInteractive = bool(f)
		} else {
			filters = append(filters, filter)
		}
	}

	opts.Filters = filters
	return isInteractive, opts
}
