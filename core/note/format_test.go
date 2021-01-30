package note

import (
	"testing"
	"time"

	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/test/assert"
)

func TestEmptyFormat(t *testing.T) {
	f, _ := newFormatter(t, opt.NewString(""))
	res, err := f.Format(Match{})
	assert.Nil(t, err)
	assert.Equal(t, res, "")
}

func TestDefaultFormat(t *testing.T) {
	f, _ := newFormatter(t, opt.NullString)
	res, err := f.Format(Match{})
	assert.Nil(t, err)
	assert.Equal(t, res, `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{#each snippets}}
{{prepend "  " (concat "‣ " .)}}
{{/each}}
`)
}

func TestFormats(t *testing.T) {
	test := func(format string, expected string) {
		f, _ := newFormatter(t, opt.NewString(format))
		actual, err := f.Format(Match{})
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	}

	// Known formats
	test("path", `{{path}}`)

	test("oneline", `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})`)

	test("short", `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{#each snippets}}
{{prepend "  " (concat "‣ " .)}}
{{/each}}
`)

	test("medium", `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}

{{#each snippets}}
{{prepend "  " (concat "‣ " .)}}
{{/each}}
`)

	test("long", `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{#each snippets}}
{{prepend "  " (concat "‣ " .)}}
{{/each}}
`)

	test("full", `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{prepend "  " body}}
`)

	// Known formats are case sensitive.
	test("Path", "Path")

	// Custom formats are used literally.
	test("{{title}}", "{{title}}")

	// \n and \t in custom formats are expanded.
	test(`{{title}}\t{{path}}\n{{snippet}}`, "{{title}}\t{{path}}\n{{snippet}}")
}

func TestFormatRenderContext(t *testing.T) {
	f, templs := newFormatter(t, opt.NewString("path"))

	_, err := f.Format(Match{
		Snippets: []string{"Note snippet"},
		Metadata: Metadata{
			Path:       "dir/note.md",
			Title:      "Note title",
			Lead:       "Lead paragraph",
			Body:       "Note body",
			RawContent: "Raw content",
			WordCount:  42,
			Created:    Now,
			Modified:   Now.Add(48 * time.Hour),
			Checksum:   "Note checksum",
		},
	})
	assert.Nil(t, err)

	// Check that the template was provided with the proper information in the
	// render context.
	assert.Equal(t, templs.Contexts, []interface{}{
		formatRenderContext{
			Path:       "dir/note.md",
			Title:      "Note title",
			Lead:       "Lead paragraph",
			Body:       "Note body",
			Snippets:   []string{"Note snippet"},
			RawContent: "Raw content",
			WordCount:  42,
			Created:    Now,
			Modified:   Now.Add(48 * time.Hour),
			Checksum:   "Note checksum",
		},
	})
}

func TestFormatPath(t *testing.T) {
	test := func(basePath, currentPath, path string, expected string) {
		f, templs := newFormatterWithPaths(t, basePath, currentPath, opt.NullString)
		_, err := f.Format(Match{
			Metadata: Metadata{Path: path},
		})
		assert.Nil(t, err)
		assert.Equal(t, templs.Contexts, []interface{}{
			formatRenderContext{
				Path:     expected,
				Snippets: []string{},
			},
		})
	}

	// Check that the path is relative to the current directory.
	test("", "", "note.md", "note.md")
	test("", "", "dir/note.md", "dir/note.md")
	test("/abs/zk", "/abs/zk", "note.md", "note.md")
	test("/abs/zk", "/abs/zk", "dir/note.md", "dir/note.md")
	test("/abs/zk", "/abs/zk/dir", "note.md", "../note.md")
	test("/abs/zk", "/abs/zk/dir", "dir/note.md", "note.md")
	test("/abs/zk", "/abs", "note.md", "zk/note.md")
	test("/abs/zk", "/abs", "dir/note.md", "zk/dir/note.md")
}

func TestFormatStylesSnippetTerm(t *testing.T) {
	test := func(snippet string, expected string) {
		f, templs := newFormatter(t, opt.NullString)
		_, err := f.Format(Match{
			Snippets: []string{snippet},
		})
		assert.Nil(t, err)
		assert.Equal(t, templs.Contexts, []interface{}{
			formatRenderContext{
				Path:     ".",
				Snippets: []string{expected},
			},
		})
	}

	test("Hello world!", "Hello world!")
	test("Hello <zk:match>world</zk:match>!", "Hello term(world)!")
	test("Hello <zk:match>world</zk:match> with <zk:match>several matches</zk:match>!", "Hello term(world) with term(several matches)!")
	test("Hello <zk:match>world</zk:match> with <zk:match>several<zk:match> matches</zk:match>!", "Hello term(world) with term(several<zk:match> matches)!")
}

func newFormatter(t *testing.T, format opt.String) (*Formatter, *TemplLoaderSpy) {
	return newFormatterWithPaths(t, "", "", format)
}

func newFormatterWithPaths(t *testing.T, basePath, currentPath string, format opt.String) (*Formatter, *TemplLoaderSpy) {
	loader := NewTemplLoaderSpy()
	styler := &StylerMock{}
	formatter, err := NewFormatter(basePath, currentPath, format, loader, styler)
	assert.Nil(t, err)
	return formatter, loader
}
