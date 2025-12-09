package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/paths"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNotebookNewNote(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
	}
	test.setup()

	note, err := test.run(NewNoteOpts{
		Title:   opt.NewString("Note title"),
		Content: "Note content",
		Extra: map[string]any{
			"add-extra": "ec83da",
		},
		Date: now,
	})

	assert.NotNil(t, note)
	assert.Nil(t, err)
	assert.Equal(t, note.Path, "filename.ext")

	// Check created note.
	assert.Equal(t, test.fs.files["/notebook/filename.ext"], "body")

	assert.Equal(t, test.receivedLang, test.config.Note.Lang)
	assert.Equal(t, test.receivedIDOpts, test.config.Note.IDOptions)

	// Check that the templates received the proper render contexts.
	assert.Equal(t, test.filenameTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Note title",
			Content:      "Note content",
			Dir:          "",
			Filename:     "",
			FilenameStem: "",
			Extra:        map[string]any{"add-extra": "ec83da", "conf-extra": "38srnw"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
	assert.Equal(t, test.bodyTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Note title",
			Content:      "Note content",
			Dir:          "",
			Filename:     "filename.ext",
			FilenameStem: "filename",
			Extra:        map[string]any{"add-extra": "ec83da", "conf-extra": "38srnw"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
}

func TestNotebookNewNoteWithDefaultTitle(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
	}
	test.setup()

	_, err := test.run(NewNoteOpts{
		Date: now,
	})

	assert.Nil(t, err)
	assert.Equal(t, test.filenameTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:    "id",
			Title: "Titre par défaut",
			Extra: map[string]any{"conf-extra": "38srnw"},
			Now:   now,
			Env:   map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
}

func TestNotebookNewNoteInUnknownDir(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
	}
	test.setup()

	_, err := test.run(NewNoteOpts{
		Directory: opt.NewString("a-dir"),
	})

	assert.Err(t, err, "a-dir: directory not found")
}

func TestNotebookNewNoteInDir(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
		dirs:    []string{"/notebook/a-dir"},
	}
	test.setup()

	note, err := test.run(NewNoteOpts{
		Title:     opt.NewString("Note title"),
		Directory: opt.NewString("a-dir"),
		Date:      now,
	})

	assert.Nil(t, err)
	assert.Equal(t, note.Path, "a-dir/filename.ext")

	// Check created note.
	assert.Equal(t, test.fs.files["/notebook/a-dir/filename.ext"], "body")

	// Check that the templates received the proper render contexts.
	assert.Equal(t, test.filenameTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Note title",
			Content:      "",
			Dir:          "a-dir",
			Filename:     "",
			FilenameStem: "",
			Extra:        map[string]any{"conf-extra": "38srnw"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
	assert.Equal(t, test.bodyTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Note title",
			Content:      "",
			Dir:          "a-dir",
			Filename:     "filename.ext",
			FilenameStem: "filename",
			Extra:        map[string]any{"conf-extra": "38srnw"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
}

// Create a note in a directory belonging to a config group which will override
// the default config.
func TestNotebookNewNoteInDirWithGroup(t *testing.T) {
	groupConfig := GroupConfig{
		Paths: []string{"a-dir"},
		Note: NoteConfig{
			DefaultTitle:     "Group default title",
			FilenameTemplate: "group-filename",
			BodyTemplatePath: opt.NewString("group-body"),
			Extension:        "group-ext",
			Lang:             "de",
			IDOptions: IDOptions{
				Length:  29,
				Charset: []rune("group"),
				Case:    CaseMixed,
			},
		},
		Extra: map[string]any{
			"group-extra": "e48rs",
		},
	}

	test := newNoteTest{
		rootDir: "/notebook",
		dirs:    []string{"/notebook/a-dir"},
		groups: map[string]GroupConfig{
			"group-a": groupConfig,
		},
	}
	test.setup()

	filenameTemplate := test.templateLoader.SpyString("group-filename.group-ext")
	bodyTemplate := test.templateLoader.SpyFile("group-body", "group template body")

	note, err := test.run(NewNoteOpts{
		Directory: opt.NewString("a-dir"),
		Date:      now,
	})

	assert.Nil(t, err)
	assert.Equal(t, note.Path, "a-dir/group-filename.group-ext")

	assert.Equal(t, test.fs.files["/notebook/a-dir/group-filename.group-ext"], "group template body")

	assert.Equal(t, test.receivedLang, groupConfig.Note.Lang)
	assert.Equal(t, test.receivedIDOpts, groupConfig.Note.IDOptions)

	// Check that the templates received the proper render contexts.
	assert.Equal(t, filenameTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Group default title",
			Content:      "",
			Dir:          "a-dir",
			Filename:     "",
			FilenameStem: "",
			Extra:        map[string]any{"group-extra": "e48rs"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
	assert.Equal(t, bodyTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Group default title",
			Content:      "",
			Dir:          "a-dir",
			Filename:     "group-filename.group-ext",
			FilenameStem: "group-filename",
			Extra:        map[string]any{"group-extra": "e48rs"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
}

// Create a note with an explicit group overriding the default config.
func TestNotebookNewNoteWithGroup(t *testing.T) {
	groupConfig := GroupConfig{
		Paths: []string{"a-dir"},
		Note: NoteConfig{
			DefaultTitle:     "Group default title",
			FilenameTemplate: "group-filename",
			BodyTemplatePath: opt.NewString("group-body"),
			Extension:        "group-ext",
			Lang:             "de",
			IDOptions: IDOptions{
				Length:  29,
				Charset: []rune("group"),
				Case:    CaseMixed,
			},
		},
		Extra: map[string]any{
			"group-extra": "e48rs",
		},
	}

	test := newNoteTest{
		rootDir: "/notebook",
		groups: map[string]GroupConfig{
			"group-a": groupConfig,
		},
	}
	test.setup()

	filenameTemplate := test.templateLoader.SpyString("group-filename.group-ext")
	bodyTemplate := test.templateLoader.SpyFile("group-body", "group template body")

	note, err := test.run(NewNoteOpts{
		Group: opt.NewString("group-a"),
		Date:  now,
	})

	assert.Nil(t, err)
	assert.Equal(t, note.Path, "group-filename.group-ext")

	// Check created note.
	assert.Equal(t, test.fs.files["/notebook/group-filename.group-ext"], "group template body")

	assert.Equal(t, test.receivedLang, groupConfig.Note.Lang)
	assert.Equal(t, test.receivedIDOpts, groupConfig.Note.IDOptions)

	// Check that the templates received the proper render contexts.
	assert.Equal(t, filenameTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Group default title",
			Content:      "",
			Dir:          "",
			Filename:     "",
			FilenameStem: "",
			Extra:        map[string]any{"group-extra": "e48rs"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
	assert.Equal(t, bodyTemplate.Contexts, []interface{}{
		newNoteTemplateContext{
			ID:           "id",
			Title:        "Group default title",
			Content:      "",
			Dir:          "",
			Filename:     "group-filename.group-ext",
			FilenameStem: "group-filename",
			Extra:        map[string]any{"group-extra": "e48rs"},
			Now:          now,
			Env:          map[string]string{"KEY1": "foo", "KEY2": "bar"},
		},
	})
}

func TestNotebookNewNoteWithUnknownGroup(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
	}
	test.setup()

	_, err := test.run(NewNoteOpts{
		Group: opt.NewString("group-a"),
		Date:  now,
	})

	assert.Err(t, err, "no group named `group-a` found in the config")
}

func TestNotebookNewNoteWithCustomTemplate(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
	}
	test.setup()
	test.templateLoader.SpyFile("custom-body", "custom body template")

	note, err := test.run(NewNoteOpts{
		Template: opt.NewString("custom-body"),
		Date:     now,
	})

	assert.Nil(t, err)
	assert.Equal(t, test.fs.files["/notebook/"+note.Path], "custom body template")
}

// Tries to generate a filename until one is free.
func TestNotebookNewNoteTriesUntilFreePath(t *testing.T) {
	test := newNoteTest{
		rootDir: "/notebook",
		files: map[string]string{
			"/notebook/filename1.ext": "file1",
			"/notebook/filename2.ext": "file2",
			"/notebook/filename3.ext": "file3",
		},
		filenameTemplateRender: func(context newNoteTemplateContext) string {
			return "filename" + context.ID + ".ext"
		},
		idGeneratorFactory: incrementingID,
	}
	test.setup()

	note, err := test.run(NewNoteOpts{
		Date: now,
	})

	assert.Nil(t, err)
	assert.Equal(t, note.Path, "filename4.ext")

	// Check created note.
	assert.Equal(t, test.fs.files["/notebook/filename4.ext"], "body")
}

func TestNotebookNewNoteErrorWhenNoFreePath(t *testing.T) {
	files := map[string]string{}
	for i := 1; i < 51; i++ {
		files[fmt.Sprintf("/notebook/filename%d.ext", i)] = "body"
	}
	test := newNoteTest{
		rootDir: "/notebook",
		files:   files,
		filenameTemplateRender: func(context newNoteTemplateContext) string {
			return "filename" + context.ID + ".ext"
		},
		idGeneratorFactory: incrementingID,
	}
	test.setup()

	_, err := test.run(NewNoteOpts{
		Date: now,
	})

	assert.Err(t, err, "/notebook/filename50.ext: note already exists")
	assert.Equal(t, test.fs.files, files)
}

var now = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

// newNoteTest builds and runs the SUT for new note test cases.
type newNoteTest struct {
	rootDir                string
	files                  map[string]string
	dirs                   []string
	fs                     *fileStorageMock
	index                  *noteIndexAddMock
	parser                 *noteContentParserMock
	config                 Config
	groups                 map[string]GroupConfig
	templateLoader         *templateLoaderMock
	filenameTemplateRender func(context newNoteTemplateContext) string
	filenameTemplate       *templateSpy
	bodyTemplate           *templateSpy
	idGeneratorFactory     IDGeneratorFactory
	osEnv                  map[string]string

	receivedLang   string
	receivedIDOpts IDOptions
}

func (t *newNoteTest) setup() {
	if t.rootDir == "" {
		t.rootDir = "/notebook"
	}
	if t.dirs == nil {
		t.dirs = []string{}
	}
	t.dirs = append(t.dirs, t.rootDir)
	t.fs = newFileStorageMock(t.rootDir, t.dirs)
	if t.files != nil {
		t.fs.files = t.files
	}

	t.index = &noteIndexAddMock{ReturnedID: 42}
	t.parser = newNoteContentParserMock(map[string]*NoteContent{})

	t.templateLoader = newTemplateLoaderMock()
	if t.filenameTemplateRender != nil {
		t.filenameTemplate = t.templateLoader.Spy("filename.ext", func(context interface{}) string {
			return t.filenameTemplateRender(context.(newNoteTemplateContext))
		})
	} else {
		t.filenameTemplate = t.templateLoader.SpyString("filename.ext")
	}
	t.bodyTemplate = t.templateLoader.SpyFile("default", "body")

	if t.idGeneratorFactory == nil {
		t.idGeneratorFactory = func(opts IDOptions) func() string {
			return func() string { return "id" }
		}
	}

	if t.osEnv == nil {
		t.osEnv = map[string]string{
			"KEY1": "foo",
			"KEY2": "bar",
		}
	}

	if t.groups == nil {
		t.groups = map[string]GroupConfig{}
	}

	t.config = Config{
		Note: NoteConfig{
			FilenameTemplate: "filename",
			Extension:        "ext",
			BodyTemplatePath: opt.NewString("default"),
			Lang:             "fr",
			DefaultTitle:     "Titre par défaut",
			IDOptions: IDOptions{
				Length:  42,
				Charset: []rune("hello"),
				Case:    CaseUpper,
			},
		},
		Groups: t.groups,
		Extra: map[string]any{
			"conf-extra": "38srnw",
		},
	}
}

func (t *newNoteTest) parseContentAsNote(content string, note *NoteContent) {
	t.parser.results[content] = note
}

func (t *newNoteTest) run(opts NewNoteOpts) (*Note, error) {
	notebook := NewNotebook(t.rootDir, t.config, NotebookPorts{
		TemplateLoaderFactory: func(language string) (TemplateLoader, error) {
			t.receivedLang = language
			return t.templateLoader, nil
		},
		IDGeneratorFactory: func(opts IDOptions) func() string {
			t.receivedIDOpts = opts
			return t.idGeneratorFactory(opts)
		},
		FS:                t.fs,
		NoteIndex:         t.index,
		NoteContentParser: t.parser,
		Logger:            &util.NullLogger,
		OSEnv:             func() map[string]string { return t.osEnv },
	})

	return notebook.NewNote(opts)
}

// incrementingID returns a generator of incrementing string ID.
func incrementingID(opts IDOptions) func() string {
	i := 0
	return func() string {
		i++
		return fmt.Sprintf("%d", i)
	}
}

type noteIndexAddMock struct {
	ReturnedID NoteID
}

func (m *noteIndexAddMock) Find(opts NoteFindOpts) ([]ContextualNote, error)     { return nil, nil }
func (m *noteIndexAddMock) FindMinimal(opts NoteFindOpts) ([]MinimalNote, error) { return nil, nil }
func (m *noteIndexAddMock) FindLinkMatch(baseDir string, href string, linkType LinkType) (NoteID, error) {
	return 0, nil
}
func (m *noteIndexAddMock) FindLinksBetweenNotes(ids []NoteID) ([]ResolvedLink, error) {
	return nil, nil
}
func (m *noteIndexAddMock) FindCollections(kind CollectionKind, sorters []CollectionSorter) ([]Collection, error) {
	return nil, nil
}
func (m *noteIndexAddMock) IndexedPaths() (<-chan paths.Metadata, error)       { return nil, nil }
func (m *noteIndexAddMock) Add(note Note) (NoteID, error)                      { return m.ReturnedID, nil }
func (m *noteIndexAddMock) Update(note Note) error                             { return nil }
func (m *noteIndexAddMock) Remove(path string) error                           { return nil }
func (m *noteIndexAddMock) Commit(transaction func(idx NoteIndex) error) error { return nil }
func (m *noteIndexAddMock) NeedsReindexing() (bool, error)                     { return false, nil }
func (m *noteIndexAddMock) SetNeedsReindexing(needsReindexing bool) error      { return nil }
