package fzf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/adapter/term"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/opt"
	stringsutil "github.com/mickael-menu/zk/internal/util/strings"
)

// NoteFilter uses fzf to filter interactively a set of notes.
type NoteFilter struct {
	opts           NoteFilterOpts
	fs             core.FileStorage
	terminal       *term.Terminal
	templateLoader core.TemplateLoader
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
	// Format for a single line, taken from the config `fzf-line` property.
	LineTemplate opt.String
	// Optionally provide additional arguments, taken from the config `fzf-options` property.
	FzfOptions opt.String
	// Preview command to run when selecting a note.
	PreviewCmd opt.String
	// When non null, a "create new note from query" binding will be added to
	// fzf to create a note in this directory.
	NewNoteDir *core.Dir
	// Absolute path to the notebook.
	NotebookDir string
}

func NewNoteFilter(opts NoteFilterOpts, fs core.FileStorage, terminal *term.Terminal, templateLoader core.TemplateLoader) *NoteFilter {
	return &NoteFilter{
		opts:           opts,
		fs:             fs,
		terminal:       terminal,
		templateLoader: templateLoader,
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

	lineTemplate, err := f.templateLoader.LoadTemplate(f.opts.LineTemplate.OrString(defaultLineTemplate).String())
	if err != nil {
		return selectedNotes, err
	}

	for _, note := range notes {
		absPath := filepath.Join(f.opts.NotebookDir, note.Path)
		absPaths = append(absPaths, absPath)
		if relPath, err := f.fs.Rel(absPath); err == nil {
			relPaths = append(relPaths, relPath)
		} else {
			relPaths = append(relPaths, note.Path)
		}
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
		Options:    f.opts.FzfOptions.OrString(defaultOptions),
		PreviewCmd: opt.NewNotEmptyString(previewCmd),
		Padding:    2,
		Bindings:   bindings,
	})
	if err != nil {
		return selectedNotes, err
	}

	for i, note := range notes {
		context := lineRenderContext{
			Filename:     note.Filename(),
			FilenameStem: note.FilenameStem(),
			Path:         note.Path,
			AbsPath:      absPaths[i],
			RelPath:      relPaths[i],
			Title:        note.Title,
			TitleOrPath:  note.Title,
			Body:         stringsutil.JoinLines(note.Body),
			RawContent:   stringsutil.JoinLines(note.RawContent),
			WordCount:    note.WordCount,
			Tags:         note.Tags,
			Metadata:     note.Metadata,
			Created:      note.Created,
			Modified:     note.Modified,
			Checksum:     note.Checksum,
		}
		if context.TitleOrPath == "" {
			context.TitleOrPath = note.Path
		}

		line, err := lineTemplate.Render(context)
		if err != nil {
			return selectedNotes, err
		}

		// The absolute path is appended at the end of the line to be used in
		// the preview command.
		absPathField := f.terminal.MustStyle(context.AbsPath, core.StyleUnderstate)
		fzf.Add([]string{line, absPathField})
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

var defaultLineTemplate = `{{style "title" title-or-path}} {{style "understate" body}} {{style "understate" (json metadata)}}`

// defaultOptions are the default fzf options used when filtering notes.
var defaultOptions = strings.Join([]string{
	"--tiebreak begin",      // Prefer matches located at the beginning of the line
	"--exact",               // Look for exact matches instead of fuzzy ones by default
	"--tabstop 4",           // Length of tab characters
	"--height 100%",         // Height of the list relative to the terminal window
	"--layout reverse",      // Display the input field at the top
	"--no-hscroll",          // Make sure the path and titles are always visible
	"--color hl:-1,hl+:-1",  // Don't highlight search terms
	"--preview-window wrap", // Enable line wrapping in the preview window
}, " ")

type lineRenderContext struct {
	Filename     string
	FilenameStem string `handlebars:"filename-stem"`
	Path         string
	AbsPath      string `handlebars:"abs-path"`
	RelPath      string `handlebars:"rel-path"`
	Title        string
	TitleOrPath  string `handlebars:"title-or-path"`
	Body         string
	RawContent   string `handlebars:"raw-content"`
	WordCount    int    `handlebars:"word-count"`
	Tags         []string
	Metadata     map[string]interface{}
	Created      time.Time
	Modified     time.Time
	Checksum     string
}
