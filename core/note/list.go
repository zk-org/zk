package note

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mickael-menu/zk/core/templ"
	"github.com/mickael-menu/zk/util/opt"
)

// MatchFilter is a note filter used to match its content with FTS predicates.
type MatchFilter string

// PathFilter is a note filter using path globs to match notes.
type PathFilter []string

// Match holds information about a note matching the list filters.
type Match struct {
	// Snippet is an excerpt of the note.
	Snippet string
	Metadata
}

func (m Match) String() string {
	return fmt.Sprintf(`note.Match{
	Snippet: "%v",
	Metadata: %v,
}`, m.Snippet, m.Metadata)
}

// Finder retrieves notes matching the given Filter.
// Returns the number of matches.
type Finder interface {
	Find(opts FinderOpts, callback func(Match) error) (int, error)
}

type FinderOpts struct {
	Filters []Filter
	Limit   int
}

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
func List(opts ListOpts, deps ListDeps, callback func(formattedNote string) error) (int, error) {
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
		return callback(res)
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
		Path:      path,
		Title:     match.Title,
		Body:      match.Body,
		WordCount: match.WordCount,
		Snippet:   snippet,
		Created:   match.Created,
		Modified:  match.Modified,
	}, err
}

type matchRenderContext struct {
	Path      string
	Title     string
	Body      string
	WordCount int
	Snippet   string
	Created   time.Time
	Modified  time.Time
}

// Filter is a sealed interface implemented by Finder filter criteria.
type Filter interface{ sealed() }

func (f MatchFilter) sealed() {}
func (f PathFilter) sealed()  {}
