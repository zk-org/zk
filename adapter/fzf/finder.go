package fzf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mickael-menu/zk/adapter/term"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/style"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/opt"
	stringsutil "github.com/mickael-menu/zk/util/strings"
)

// NoteFinder wraps a note.Finder and filters its result interactively using fzf.
type NoteFinder struct {
	opts     NoteFinderOpts
	finder   note.Finder
	terminal *term.Terminal
}

// NoteFinderOpts holds the configuration for the fzf notes finder.
//
// The absolute path to the slip box (BasePath) and the working directory
// (CurrentPath) are used to make the path of each note relative to the working
// directory.
type NoteFinderOpts struct {
	// Indicates whether fzf is opened for every query, even if empty.
	AlwaysFilter bool
	// Preview command to run when selecting a note.
	PreviewCmd opt.String
	// When non nil, a "create new note from query" binding will be added to
	// fzf to create a note in this directory.
	NewNoteDir *zk.Dir
	// Absolute path to the slip box.
	BasePath string
	// Path to the working directory.
	CurrentPath string
}

func NewNoteFinder(opts NoteFinderOpts, finder note.Finder, terminal *term.Terminal) *NoteFinder {
	return &NoteFinder{
		opts:     opts,
		finder:   finder,
		terminal: terminal,
	}
}

func (f *NoteFinder) Find(opts note.FinderOpts) ([]note.Match, error) {
	isInteractive, opts := popInteractiveFilter(opts)
	selectedMatches := make([]note.Match, 0)
	matches, err := f.finder.Find(opts)
	relPaths := []string{}

	if !isInteractive || !f.terminal.IsInteractive() || err != nil || (!f.opts.AlwaysFilter && len(matches) == 0) {
		return matches, err
	}

	for _, match := range matches {
		path, err := filepath.Rel(f.opts.CurrentPath, filepath.Join(f.opts.BasePath, match.Path))
		if err != nil {
			return selectedMatches, err
		}
		relPaths = append(relPaths, path)
	}

	zkBin, err := os.Executable()
	if err != nil {
		return selectedMatches, err
	}

	bindings := []Binding{}

	if dir := f.opts.NewNoteDir; dir != nil {
		suffix := ""
		if dir.Name != "" {
			suffix = " in " + dir.Name + "/"
		}

		bindings = append(bindings, Binding{
			Keys:        "Ctrl-N",
			Description: "create a note with the query as title" + suffix,
			Action:      fmt.Sprintf("abort+execute(%s new %s --title {q} < /dev/tty > /dev/tty)", zkBin, dir.Path),
		})
	}

	fzf, err := New(Opts{
		PreviewCmd: f.opts.PreviewCmd.OrString("cat {1}").NonEmpty(),
		Padding:    2,
		Bindings:   bindings,
	})
	if err != nil {
		return selectedMatches, err
	}

	for i, match := range matches {
		fzf.Add([]string{
			relPaths[i],
			f.terminal.MustStyle(match.Title, style.RuleYellow),
			f.terminal.MustStyle(stringsutil.JoinLines(match.Body), style.RuleBlack),
		})
	}

	selection, err := fzf.Selection()
	if err != nil {
		return selectedMatches, err
	}

	for _, s := range selection {
		path := s[0]
		for i, m := range matches {
			if relPaths[i] == path {
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
