package fzf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/adapter/term"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/opt"
	stringsutil "github.com/mickael-menu/zk/internal/util/strings"
)

// NoteFilter uses fzf to filter interactively a set of notes.
type NoteFilter struct {
	opts     NoteFilterOpts
	fs       core.FileStorage
	terminal *term.Terminal
}

// NoteFilterOpts holds the configuration for the fzf notes filtering.
//
// The absolute path to the notebook (NotebookDir) and the working directory
// (WorkingDir) are used to make the path of each note relative to the working
// directory.
type NoteFilterOpts struct {
	// Indicates whether the filtering is interactive. If not, fzf is bypassed.
	Interactive bool
	// Indicates whether fzf is opened for every query, even if empty.
	AlwaysFilter bool
	// Preview command to run when selecting a note.
	PreviewCmd opt.String
	// When non null, a "create new note from query" binding will be added to
	// fzf to create a note in this directory.
	NewNoteDir *core.Dir
	// Absolute path to the notebook.
	NotebookDir string
}

func NewNoteFilter(opts NoteFilterOpts, fs core.FileStorage, terminal *term.Terminal) *NoteFilter {
	return &NoteFilter{
		opts:     opts,
		fs:       fs,
		terminal: terminal,
	}
}

// Apply filters the given notes with fzf.
func (f *NoteFilter) Apply(notes []core.ContextualNote) ([]core.ContextualNote, error) {
	selectedNotes := make([]core.ContextualNote, 0)
	relPaths := []string{}
	absPaths := []string{}

	if !f.opts.Interactive || !f.terminal.IsInteractive() || (!f.opts.AlwaysFilter && len(notes) == 0) {
		return notes, nil
	}

	for _, note := range notes {
		absPaths = append(absPaths, filepath.Join(f.opts.NotebookDir, note.Path))
		relPaths = append(relPaths, note.Path)
	}

	zkBin, err := os.Executable()
	if err != nil {
		return selectedNotes, err
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
			Action:      fmt.Sprintf(`abort+execute("%s" new "%s" --title {q} < /dev/tty > /dev/tty)`, zkBin, dir.Path),
		})
	}

	previewCmd := f.opts.PreviewCmd.OrString("cat {-1}").Unwrap()

	fzf, err := New(Opts{
		PreviewCmd: opt.NewNotEmptyString(previewCmd),
		Padding:    2,
		Bindings:   bindings,
	})
	if err != nil {
		return selectedNotes, err
	}

	for i, note := range notes {
		title := note.Title
		if title == "" {
			title = relPaths[i]
		}
		fzf.Add([]string{
			f.terminal.MustStyle(title, core.StyleYellow),
			f.terminal.MustStyle(stringsutil.JoinLines(note.Body), core.StyleUnderstate),
			f.terminal.MustStyle(absPaths[i], core.StyleUnderstate),
		})
	}

	selection, err := fzf.Selection()
	if err != nil {
		return selectedNotes, err
	}

	for _, s := range selection {
		path := s[len(s)-1]
		for i, m := range notes {
			if absPaths[i] == path {
				selectedNotes = append(selectedNotes, m)
			}
		}
	}

	return selectedNotes, nil
}
