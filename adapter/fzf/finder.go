package fzf

import (
	"os"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/style"
	"github.com/mickael-menu/zk/util/opt"
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

	if !isInteractive || err != nil || len(matches) == 0 {
		return matches, err
	}

	selectedMatches := make([]note.Match, 0)

	zkBin, err := os.Executable()
	if err != nil {
		return selectedMatches, err
	}

	fzf, err := New(Opts{
		// PreviewCmd: opt.NewString("bat -p --theme Nord --color always {1}"),
		PreviewCmd: opt.NewString(zkBin + " list -f {{raw-content}} {1}"),
		Padding:    2,
	})
	if err != nil {
		return selectedMatches, err
	}

	for _, match := range matches {
		fzf.Add([]string{
			match.Path,
			f.styler.MustStyle(match.Title, style.Rule("yellow")),
			f.styler.MustStyle(stringsutil.JoinLines(match.Body), style.Rule("faint")),
		})
	}

	selection, err := fzf.Selection()
	if err != nil {
		return selectedMatches, err
	}

	for _, s := range selection {
		path := s[0]
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
