package core

import (
	"path/filepath"
	"regexp"
	"time"
)

// NoteFormatter formats notes to be printed on the screen.
type NoteFormatter func(note ContextualNote) (string, error)

func newNoteFormatter(basePath string, template Template, fs FileStorage) (NoteFormatter, error) {
	termRepl, err := template.Styler().Style("$1", StyleTerm)
	if err != nil {
		return nil, err
	}

	return func(note ContextualNote) (string, error) {
		path, err := fs.Rel(filepath.Join(basePath, note.Path))
		if err != nil {
			return "", err
		}

		snippets := make([]string, 0)
		for _, snippet := range note.Snippets {
			snippets = append(snippets, noteTermRegex.ReplaceAllString(snippet, termRepl))
		}

		return template.Render(noteFormatRenderContext{
			Path:       path,
			Title:      note.Title,
			Lead:       note.Lead,
			Body:       note.Body,
			Snippets:   snippets,
			Tags:       note.Tags,
			RawContent: note.RawContent,
			WordCount:  note.WordCount,
			Metadata:   note.Metadata,
			Created:    note.Created,
			Modified:   note.Modified,
			Checksum:   note.Checksum,
		})
	}, nil
}

var noteTermRegex = regexp.MustCompile(`<zk:match>(.*?)</zk:match>`)

// noteFormatRenderContext holds the variables available to the note formatting
// templates.
type noteFormatRenderContext struct {
	Path       string
	Title      string
	Lead       string
	Body       string
	Snippets   []string
	RawContent string `handlebars:"raw-content"`
	WordCount  int    `handlebars:"word-count"`
	Tags       []string
	Metadata   map[string]interface{}
	Created    time.Time
	Modified   time.Time
	Checksum   string
	Env        map[string]string
}
