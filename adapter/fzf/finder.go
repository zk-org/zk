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
// The absolute path to the notebook (BasePath) and the working directory
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
	// Absolute path to the notebook.
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
	selectedMatches := make([]note.Match, 0)
	matches, err := f.finder.Find(opts)
	relPaths := []string{}

	if !opts.Interactive || !f.terminal.IsInteractive() || err != nil || (!f.opts.AlwaysFilter && len(matches) == 0) {
		return matches, err
	}

	for _, match := range matches {
		absPath := filepath.Join(f.opts.BasePath, match.Path)
		relPath, err := filepath.Rel(f.opts.CurrentPath, absPath)
		if err != nil {
			return selectedMatches, err
		}
		relPaths = append(relPaths, relPath)
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

	previewCmd := f.opts.PreviewCmd.OrString("cat {-1}").Unwrap()
	if previewCmd != "" {
		// The note paths will be relative to the current path, so we need to
		// move there otherwise the preview command will fail.
		previewCmd = `cd "` + f.opts.CurrentPath + `" && ` + previewCmd
	}

	fzf, err := New(Opts{
		PreviewCmd: opt.NewNotEmptyString(previewCmd),
		Padding:    2,
		Bindings:   bindings,
	})
	if err != nil {
		return selectedMatches, err
	}

	for i, match := range matches {
		title := match.Title
		if title == "" {
			title = relPaths[i]
		}
		fzf.Add([]string{
			f.terminal.MustStyle(title, style.RuleYellow),
			f.terminal.MustStyle(stringsutil.JoinLines(match.Body), style.RuleUnderstate),
			f.terminal.MustStyle(relPaths[i], style.RuleUnderstate),
		})
	}

	selection, err := fzf.Selection()
	if err != nil {
		return selectedMatches, err
	}

	for _, s := range selection {
		path := s[len(s)-1]
		for i, m := range matches {
			if relPaths[i] == path {
				selectedMatches = append(selectedMatches, m)
			}
		}
	}

	return selectedMatches, nil
}
