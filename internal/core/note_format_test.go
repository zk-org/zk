package core

import (
	"testing"
	"time"

	"github.com/mickael-menu/zk/internal/util/test/assert"
)

func TestNewNoteFormatter(t *testing.T) {
	test := formatTest{
		format: "format",
	}
	test.setup()

	var date1 = time.Date(2009, 1, 17, 20, 34, 58, 651387237, time.UTC)
	var date2 = time.Date(2009, 2, 17, 20, 34, 58, 651387237, time.UTC)
	var date3 = time.Date(2009, 3, 17, 20, 34, 58, 651387237, time.UTC)
	var date4 = time.Date(2009, 4, 17, 20, 34, 58, 651387237, time.UTC)

	formatter, err := test.run("format")
	assert.Nil(t, err)
	assert.Equal(t, test.receivedLang, "fr")

	res, err := formatter(ContextualNote{
		Note: Note{
			ID:         1,
			Path:       "note1",
			Title:      "Note 1",
			Lead:       "Lead 1",
			Body:       "Body 1",
			RawContent: "Content 1",
			WordCount:  1,
			Tags:       []string{"tag1", "tag2"},
			Metadata: map[string]interface{}{
				"metadata1": "val1",
				"metadata2": "val2",
			},
			Created:  date1,
			Modified: date2,
			Checksum: "checksum1",
		},
		Snippets: []string{"snippet1", "snippet2"},
	})
	assert.Nil(t, err)
	assert.Equal(t, res, "format")

	res, err = formatter(ContextualNote{
		Note: Note{
			ID:         2,
			Path:       "dir/note2",
			Title:      "Note 2",
			Lead:       "Lead 2",
			Body:       "Body 2",
			RawContent: "Content 2",
			WordCount:  2,
			Tags:       []string{},
			Metadata:   map[string]interface{}{},
			Created:    date3,
			Modified:   date4,
			Checksum:   "checksum2",
		},
		Snippets: []string{},
	})
	assert.Nil(t, err)
	assert.Equal(t, res, "format")

	// Check that the template received the proper contexts
	assert.Equal(t, test.template.Contexts, []interface{}{
		noteFormatRenderContext{
			Path:       "note1",
			Title:      "Note 1",
			Lead:       "Lead 1",
			Body:       "Body 1",
			Snippets:   []string{"snippet1", "snippet2"},
			RawContent: "Content 1",
			WordCount:  1,
			Tags:       []string{"tag1", "tag2"},
			Metadata: map[string]interface{}{
				"metadata1": "val1",
				"metadata2": "val2",
			},
			Created:  date1,
			Modified: date2,
			Checksum: "checksum1",
		},
		noteFormatRenderContext{
			Path:       "dir/note2",
			Title:      "Note 2",
			Lead:       "Lead 2",
			Body:       "Body 2",
			Snippets:   []string{},
			RawContent: "Content 2",
			WordCount:  2,
			Tags:       []string{},
			Metadata:   map[string]interface{}{},
			Created:    date3,
			Modified:   date4,
			Checksum:   "checksum2",
		},
	})
}

func TestNoteFormatterMakesPathRelative(t *testing.T) {
	test := func(basePath, currentPath, path string, expected string) {
		test := formatTest{
			rootDir:    basePath,
			workingDir: currentPath,
		}
		test.setup()
		formatter, err := test.run("format")
		assert.Nil(t, err)
		_, err = formatter(ContextualNote{
			Note: Note{Path: path},
		})
		assert.Nil(t, err)
		assert.Equal(t, test.template.Contexts, []interface{}{
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

func TestNoteFormatterStylesSnippetTerm(t *testing.T) {
	test := func(snippet string, expected string) {
		test := formatTest{}
		test.setup()
		formatter, err := test.run("format")
		assert.Nil(t, err)
		_, err = formatter(ContextualNote{
			Snippets: []string{snippet},
		})
		assert.Nil(t, err)
		assert.Equal(t, test.template.Contexts, []interface{}{
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

// formatTest builds and runs the SUT for note formatter test cases.
type formatTest struct {
	format         string
	rootDir        string
	workingDir     string
	fs             *fileStorageMock
	config         Config
	templateLoader *templateLoaderMock
	template       *templateSpy
	receivedLang   string
}

func (t *formatTest) setup() {
	if t.format == "" {
		t.format = "format"
	}

	if t.rootDir == "" {
		t.rootDir = "/notebook"
	}
	if t.workingDir == "" {
		t.workingDir = t.rootDir
	}
	t.fs = newFileStorageMock(t.workingDir, []string{})

	t.templateLoader = newTemplateLoaderMock()
	t.template = t.templateLoader.SpyString(t.format)

	t.config = NewDefaultConfig()
	t.config.Note.Lang = "fr"
}

func (t *formatTest) run(format string) (NoteFormatter, error) {
	notebook := NewNotebook(t.rootDir, t.config, NotebookPorts{
		TemplateLoaderFactory: func(language string) (TemplateLoader, error) {
			t.receivedLang = language
			return t.templateLoader, nil
		},
		FS: t.fs,
	})

	return notebook.NewNoteFormatter(format)
}
