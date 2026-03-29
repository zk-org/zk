package core

import (
	"testing"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/paths"
	"github.com/zk-org/zk/internal/util/test/assert"
)

type noteIndexMock struct {
	notes             map[string]*ContextualNote
	findMinimalResult []MinimalNote
}

func (m *noteIndexMock) Find(opts NoteFindOpts) ([]ContextualNote, error) {
	return []ContextualNote{}, nil
}

func (m *noteIndexMock) FindMinimal(opts NoteFindOpts) ([]MinimalNote, error) {
	return m.findMinimalResult, nil
}

func (m *noteIndexMock) FindLinksBetweenNotes(ids []NoteID) ([]ResolvedLink, error) {
	return []ResolvedLink{}, nil
}

func (m *noteIndexMock) FindCollections(kind CollectionKind, sorters []CollectionSorter) ([]Collection, error) {
	return []Collection{}, nil
}

func (m *noteIndexMock) IndexedPaths() (<-chan paths.Metadata, error) {
	return nil, nil
}

func (m *noteIndexMock) Add(note Note) (NoteID, error) {
	return 1, nil
}

func (m *noteIndexMock) Update(note Note) error {
	return nil
}

func (m *noteIndexMock) Remove(path string) error {
	return nil
}

func (m *noteIndexMock) Commit(transaction func(idx NoteIndex) error) error {
	return transaction(m)
}

func (m *noteIndexMock) NeedsReindexing() (bool, error) {
	return false, nil
}

func (m *noteIndexMock) SetNeedsReindexing(needsReindexing bool) error {
	return nil
}

type notebookTest struct {
	rootDir  string
	dirs     []string
	fs       *fileStorageMock
	index    *noteIndexMock
	parser   *noteContentParserMock
	config   Config
	notebook *Notebook
}

func (t *notebookTest) setup() {
	if t.rootDir == "" {
		t.rootDir = "/notebook"
	}
	if t.dirs == nil {
		t.dirs = []string{}
	}
	t.dirs = append(t.dirs, t.rootDir)

	t.fs = newFileStorageMock(t.rootDir, t.dirs)
	if t.index == nil {
		t.index = &noteIndexMock{}
	}
	t.parser = newNoteContentParserMock(map[string]*NoteContent{})

	if t.config.Note.Lang == "" {
		t.config = NewDefaultConfig()
	}

	t.notebook = NewNotebook(t.rootDir, t.config, NotebookPorts{
		NoteIndex:         t.index,
		NoteContentParser: t.parser,
		TemplateLoaderFactory: func(language string) (TemplateLoader, error) {
			return newTemplateLoaderMock(), nil
		},
		IDGeneratorFactory: func(opts IDOptions) func() string {
			return func() string { return "id" }
		},
		FS:     t.fs,
		Logger: &util.NullLogger,
		OSEnv: func() map[string]string {
			return map[string]string{}
		},
	})
}

func TestNotebookFindMinimalNotes(t *testing.T) {
	var tests = []struct {
		name  string
		notes []MinimalNote
	}{{
		name:  "empty",
		notes: []MinimalNote{},
	}, {
		name: "some",
		notes: []MinimalNote{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := &noteIndexMock{
				findMinimalResult: tt.notes,
			}
			test := notebookTest{
				rootDir: "/notebook",
				index:   index,
			}
			test.setup()
			notes, err := test.notebook.FindMinimalNotes(NoteFindOpts{})
			assert.Nil(t, err)
			assert.Equal(t, len(notes), len(tt.notes))
		})
	}
}
