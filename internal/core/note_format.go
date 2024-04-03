package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
    "strings"
)

// NoteFormatter formats notes to be printed on the screen.
type NoteFormatter func(note ContextualNote) (string, error)

func newNoteFormatter(basePath string, template Template, linkFormatter LinkFormatter, env map[string]string, fs FileStorage) (NoteFormatter, error) {
	termRepl, err := template.Styler().Style("$1", StyleTerm)
	if err != nil {
		return nil, err
	}

	return func(note ContextualNote) (string, error) {
		path := NotebookPath{
			Path:       note.Path,
			BasePath:   basePath,
			WorkingDir: fs.WorkingDir(),
		}
		relPath, err := path.PathRelToWorkingDir()
		if err != nil {
			return "", err
		}

		snippets := make([]string, 0)
		for _, snippet := range note.Snippets {
			snippets = append(snippets, noteTermRegex.ReplaceAllString(snippet, termRepl))
		}

        //FIXME: if notes have `"` in their titles, they will break when 
        //executing `zk graph --format json` as `Link:...` gets unescaped quotes 
        //from `newLazyStringer`. Issue: https://github.com/zk-org/zk/issues/389
        // escaping the quotes breaks tesh tests, but is the test perhaps 
        // guarding a less robust solution?
		return template.Render(noteFormatRenderContext{
			Filename:     note.Filename(),
			FilenameStem: note.FilenameStem(),
			Path:         relPath,
			AbsPath:      path.AbsPath(),
			Title:        note.Title,
			Link: newLazyStringer(func() string {
				context, err := NewLinkFormatterContext(path, note.Title, note.Metadata)
				if err != nil {
					return ""
				}
				link, _ := linkFormatter(context)
                link = strings.ReplaceAll(link, `"`, `\"`) // breaks tesh test
				return link
			}),
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
			Env:        env,
		})
	}, nil
}

var noteTermRegex = regexp.MustCompile(`<zk:match>(.*?)</zk:match>`)

// noteFormatRenderContext holds the variables available to the note formatting
// templates.
type noteFormatRenderContext struct {
	Filename     string                 `json:"filename"`
	FilenameStem string                 `json:"filenameStem" handlebars:"filename-stem"`
	Path         string                 `json:"path"`
	AbsPath      string                 `json:"absPath" handlebars:"abs-path"`
	Title        string                 `json:"title"`
	Link         fmt.Stringer           `json:"link"`
	Lead         string                 `json:"lead"`
	Body         string                 `json:"body"`
	Snippets     []string               `json:"snippets"`
	RawContent   string                 `json:"rawContent" handlebars:"raw-content"`
	WordCount    int                    `json:"wordCount" handlebars:"word-count"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	Created      time.Time              `json:"created"`
	Modified     time.Time              `json:"modified"`
	Checksum     string                 `json:"checksum"`
	Env          map[string]string      `json:"-"`
}

func (c noteFormatRenderContext) Equal(other noteFormatRenderContext) bool {
	json1, err := json.Marshal(c)
	if err != nil {
		return false
	}
	json2, err := json.Marshal(other)
	if err != nil {
		return false
	}
	return string(json1) == string(json2)
}
