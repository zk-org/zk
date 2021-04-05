package core

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/util/opt"
)

// NoteFormetter formats notes to be printed on the screen.
type NoteFormatter struct {
	basePath    string
	currentPath string
	template    Template
	// Regex replacement for a term marked in a snippet.
	snippetTermReplacement string
}

// NewNoteFormatter creates a NoteFormatter from a given format template.
//
// The absolute path to the notebook (basePath) and the working directory
// (currentPath) are used to make the path of each note relative to the working
// directory.
func NewNoteFormatter(basePath string, currentPath string, format opt.String, templates TemplateLoader, styler Styler) (*NoteFormatter, error) {
	template, err := templates.LoadTemplate(resolveNoteFormat(format))
	if err != nil {
		return nil, err
	}

	termRepl, err := styler.Style("$1", StyleTerm)
	if err != nil {
		return nil, err
	}

	return &NoteFormatter{
		basePath:               basePath,
		currentPath:            currentPath,
		template:               template,
		snippetTermReplacement: termRepl,
	}, nil
}

func resolveNoteFormat(format opt.String) string {
	templ, ok := defaultNoteFormats[format.OrString("short").Unwrap()]
	if !ok {
		templ = format.String()
		// Replace raw \n and \t by actual newlines and tabs in user format.
		templ = strings.ReplaceAll(templ, "\\n", "\n")
		templ = strings.ReplaceAll(templ, "\\t", "\t")
	}
	return templ
}

var defaultNoteFormats = map[string]string{
	"path": `{{path}}`,

	"oneline": `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})`,

	"short": `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{list snippets}}`,

	"medium": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}

{{list snippets}}`,

	"long": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{list snippets}}`,

	"full": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}
Tags: {{join tags ", "}}

{{prepend "  " body}}
`,
}

var noteTermRegex = regexp.MustCompile(`<zk:match>(.*?)</zk:match>`)

// Format formats a note to be printed on the screen.
func (f *NoteFormatter) Format(note ContextualNote) (string, error) {
	path, err := filepath.Rel(f.currentPath, filepath.Join(f.basePath, note.Path))
	if err != nil {
		return "", err
	}

	snippets := make([]string, 0)
	for _, snippet := range note.Snippets {
		snippets = append(snippets, noteTermRegex.ReplaceAllString(snippet, f.snippetTermReplacement))
	}

	return f.template.Render(noteFormatRenderContext{
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
}

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
