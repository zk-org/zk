package note

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/core/style"
	"github.com/mickael-menu/zk/internal/util/opt"
)

// Formatter formats notes to be printed on the screen.
type Formatter struct {
	basePath    string
	currentPath string
	renderer    core.Template
	// Regex replacement for a term marked in a snippet.
	snippetTermReplacement string
}

// NewFormatter creates a Formatter from a given format template.
//
// The absolute path to the notebook (basePath) and the working directory
// (currentPath) are used to make the path of each note relative to the working
// directory.
func NewFormatter(basePath string, currentPath string, format opt.String, templates core.TemplateLoader, styler style.Styler) (*Formatter, error) {
	template := resolveFormatTemplate(format)
	renderer, err := templates.LoadTemplate(template)
	if err != nil {
		return nil, err
	}

	termRepl, err := styler.Style("$1", style.RuleTerm)
	if err != nil {
		return nil, err
	}

	return &Formatter{
		basePath:               basePath,
		currentPath:            currentPath,
		renderer:               renderer,
		snippetTermReplacement: termRepl,
	}, nil
}

func resolveFormatTemplate(format opt.String) string {
	templ, ok := formatTemplates[format.OrString("short").Unwrap()]
	if !ok {
		templ = format.String()
		// Replace raw \n and \t by actual newlines and tabs in user format.
		templ = strings.ReplaceAll(templ, "\\n", "\n")
		templ = strings.ReplaceAll(templ, "\\t", "\t")
	}
	return templ
}

var formatTemplates = map[string]string{
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

var termRegex = regexp.MustCompile(`<zk:match>(.*?)</zk:match>`)

// Format formats a note to be printed on the screen.
func (f *Formatter) Format(match Match) (string, error) {
	path, err := filepath.Rel(f.currentPath, filepath.Join(f.basePath, match.Path))
	if err != nil {
		return "", err
	}

	snippets := make([]string, 0)
	for _, snippet := range match.Snippets {
		snippets = append(snippets, termRegex.ReplaceAllString(snippet, f.snippetTermReplacement))
	}

	return f.renderer.Render(formatRenderContext{
		Path:       path,
		Title:      match.Title,
		Lead:       match.Lead,
		Body:       match.Body,
		Snippets:   snippets,
		Tags:       match.Tags,
		RawContent: match.RawContent,
		WordCount:  match.WordCount,
		Metadata:   match.Metadata.Metadata,
		Created:    match.Created,
		Modified:   match.Modified,
		Checksum:   match.Checksum,
	})
}

type formatRenderContext struct {
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
