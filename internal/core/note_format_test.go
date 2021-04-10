package core

// FIXME
/*
func TestEmptyNoteFormat(t *testing.T) {
	f, _ := newNoteFormatter(t, opt.NewString(""))
	res, err := f.Format(ContextualNote{})
	assert.Nil(t, err)
	assert.Equal(t, res, "")
}

func TestDefaultNoteFormat(t *testing.T) {
	f, _ := newNoteFormatter(t, opt.NullString)
	res, err := f.Format(ContextualNote{})
	assert.Nil(t, err)
	assert.Equal(t, res, `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{list snippets}}`)
}

func TestNoteFormats(t *testing.T) {
	test := func(format string, expected string) {
		f, _ := newNoteFormatter(t, opt.NewString(format))
		actual, err := f.Format(ContextualNote{})
		assert.Nil(t, err)
		assert.Equal(t, actual, expected)
	}

	// Known formats
	test("path", `{{path}}`)

	test("oneline", `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})`)

	test("short", `{{style "title" title}} {{style "path" path}} ({{date created "elapsed"}})

{{list snippets}}`)

	test("medium", `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}

{{list snippets}}`)

	test("long", `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}

{{list snippets}}`)

	test("full", `{{style "title" title}} {{style "path" path}}
Created: {{date created "short"}}
Modified: {{date created "short"}}
Tags: {{join tags ", "}}

{{prepend "  " body}}
`)

	// Known formats are case sensitive.
	test("Path", "Path")

	// Custom formats are used literally.
	test("{{title}}", "{{title}}")

	// \n and \t in custom formats are expanded.
	test(`{{title}}\t{{path}}\n{{snippet}}`, "{{title}}\t{{path}}\n{{snippet}}")
}

func TestNoteFormatRenderContext(t *testing.T) {
	f, templs := newNoteFormatter(t, opt.NewString("path"))

	_, err := f.Format(ContextualNote{
		Snippets: []string{"Note snippet"},
		Note: Note{
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
		noteFormatRenderContext{
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

func TestNoteFormatPath(t *testing.T) {
	test := func(basePath, currentPath, path string, expected string) {
		f, templs := newNoteFormatterWithPaths(t, basePath, currentPath, opt.NullString)
		_, err := f.Format(ContextualNote{
			Note: Note{Path: path},
		})
		assert.Nil(t, err)
		assert.Equal(t, templs.Contexts, []interface{}{
			noteFormatRenderContext{
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

func TestNoteFormatStylesSnippetTerm(t *testing.T) {
	test := func(snippet string, expected string) {
		f, templs := newNoteFormatter(t, opt.NullString)
		_, err := f.Format(ContextualNote{
			Snippets: []string{snippet},
		})
		assert.Nil(t, err)
		assert.Equal(t, templs.Contexts, []interface{}{
			noteFormatRenderContext{
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

func newNoteFormatter(t *testing.T, format opt.String) (*NoteFormatter, *TemplateLoaderSpy) {
	return newNoteFormatterWithPaths(t, "", "", format)
}

func newNoteFormatterWithPaths(t *testing.T, basePath, currentPath string, format opt.String) (*NoteFormatter, *TemplateLoaderSpy) {
	templates := NewTemplateLoaderSpy()
	styler := &StylerMock{}
	formatter, err := NewNoteFormatter(basePath, currentPath, format, templates, styler)
	assert.Nil(t, err)
	return formatter, templates
}
*/
