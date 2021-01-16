package note

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mickael-menu/zk/core/templ"
	"github.com/mickael-menu/zk/util/opt"
)

type ListOpts struct {
	Format opt.String
	FinderOpts
}

type ListDeps struct {
	BasePath  string
	Finder    Finder
	Templates templ.Loader
}

// List finds notes matching given criteria and formats them according to user
// preference.
func List(opts ListOpts, deps ListDeps, out io.Writer) (int, error) {
	templ := matchTemplate(opts.Format)
	template, err := deps.Templates.Load(templ)
	if err != nil {
		return 0, err
	}

	return deps.Finder.Find(opts.FinderOpts, func(note Match) error {
		ft, err := format(note, deps.BasePath, deps.Templates)
		if err != nil {
			return err
		}
		res, err := template.Render(ft)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(out, res)
		return err
	})
}

var templates = map[string]string{
	"path": `{{path}}`,

	"oneline": `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})`,

	"short": `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{prepend "  " snippet}}
`,

	"medium": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}

{{prepend "  " snippet}}
`,

	"long": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{prepend "  " snippet}}
`,

	"full": `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{prepend "  " body}}
`,
}

func matchTemplate(format opt.String) string {
	templ, ok := templates[format.OrDefault("short")]
	if !ok {
		templ = format.String()
		// Replace raw \n and \t by actual newlines and tabs in user format.
		templ = strings.ReplaceAll(templ, "\\n", "\n")
		templ = strings.ReplaceAll(templ, "\\t", "\t")
	}
	return templ
}

func format(match Match, basePath string, templates templ.Loader) (*matchRenderContext, error) {
	re := regexp.MustCompile(`<zk:match>(.*?)</zk:match>`)

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	path, err := filepath.Rel(wd, filepath.Join(basePath, match.Path))
	if err != nil {
		return nil, err
	}

	snippet := strings.TrimSpace(re.ReplaceAllString(match.Snippet, `{{#style "match"}}$1{{/style}}`))
	snippetTempl, err := templates.Load(snippet)
	if err != nil {
		return nil, err
	}
	snippet, err = snippetTempl.Render(nil)
	if err != nil {
		return nil, err
	}

	return &matchRenderContext{
		Path:       path,
		Title:      match.Title,
		Lead:       match.Lead,
		Body:       match.Body,
		RawContent: match.RawContent,
		WordCount:  match.WordCount,
		Snippet:    snippet,
		Created:    match.Created,
		Modified:   match.Modified,
	}, err
}

type matchRenderContext struct {
	Path       string
	Title      string
	Lead       string
	Body       string
	RawContent string
	WordCount  int
	Snippet    string
	Created    time.Time
	Modified   time.Time
}
