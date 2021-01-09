package note

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mickael-menu/zk/core/templ"
	"github.com/mickael-menu/zk/util/opt"
)

// MatchFilter is a note filter used to match its content with FTS predicates.
type MatchFilter string

// Match holds information about a note matching the list filters.
type Match struct {
	// Snippet is an excerpt of the note.
	Snippet string
	Metadata
}

// Finder retrieves notes matching the given Filter.
type Finder interface {
	Find(callback func(Match) error, filters ...Filter) error
}

// Unit represents a type of component of note, for example its path or its title.
type Unit int

const (
	UnitPath Unit = iota + 1
	UnitTitle
	UnitBody
	UnitSnippet
	UnitMatch
	UnitWordCount
	UnitDate
	UnitChecksum
)

type ListOpts struct {
	Format  opt.String
	Filters []Filter
}

type ListDeps struct {
	BasePath  string
	Finder    Finder
	Templates templ.Loader
}

// List finds notes matching given criteria and formats them according to user
// preference.
func List(opts ListOpts, deps ListDeps, callback func(formattedNote string) error) error {
	templ := matchTemplate(opts.Format)
	template, err := deps.Templates.Load(templ)
	if err != nil {
		return err
	}

	return deps.Finder.Find(func(note Match) error {
		ft, err := format(note, deps.BasePath)
		if err != nil {
			return err
		}
		res, err := template.Render(ft)
		if err != nil {
			return err
		}
		return callback(res)
	}, opts.Filters...)
}

var templates = map[string]string{
	"path": `{{path}}`,

	"oneline": `{{path}} {{title}} ({{date created "elapsed"}})`,

	"short": `{{path}} {{title}} ({{date created "elapsed"}})

{{prepend "  " snippet}}
`,

	"medium": `{{path}} {{title}}
Created: {{date created "short"}}

{{prepend "  " snippet}}
`,

	"long": `{{path}} {{title}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{prepend "  " snippet}}
`,

	"full": `{{path}} {{title}}
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

func format(match Match, basePath string) (*matchRenderContext, error) {
	color.NoColor = false // Otherwise the colors are not displayed in `less -r`.
	path := color.New(color.FgCyan).SprintFunc()
	title := color.New(color.FgYellow).SprintFunc()
	term := color.New(color.FgRed).SprintFunc()

	re := regexp.MustCompile(`<zk:match>(.*?)</zk:match>`)

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	pth, err := filepath.Rel(wd, filepath.Join(basePath, match.Path))
	if err != nil {
		return nil, err
	}

	return &matchRenderContext{
		Path:      path(pth),
		Title:     title(match.Title),
		Body:      match.Body,
		WordCount: match.WordCount,
		Snippet:   strings.TrimSpace(re.ReplaceAllString(match.Snippet, term("$1"))),
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
