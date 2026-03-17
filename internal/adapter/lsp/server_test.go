package lsp

import (
	"net/url"
	"path/filepath"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/adapter/fs"
	"github.com/zk-org/zk/internal/adapter/markdown"
	"github.com/zk-org/zk/internal/adapter/sqlite"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/fixtures"
	"github.com/zk-org/zk/internal/util/test/assert"
)

type mockFileStorage struct {
	core.FileStorage
}

func (m *mockFileStorage) Canonical(path string) string {
	return path
}

func getNotebookFixture(name string) notebookFixture {
	p := fixtures.Path(name)
	return newNotebookFixture(p)
}

type notebookFixture struct {
	Path string
	FS   *fs.FileStorage
}

func newNotebookFixture(path string) notebookFixture {
	return notebookFixture{
		Path: path,
	}
}

// Read content and create parameter for DidOpen
func (notebook *notebookFixture) MakeDidOpenParam(noteName string) (protocol.DidOpenTextDocumentParams, error) {
	path := filepath.Join(notebook.Path, noteName)
	u := url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(path),
	}
	content, err := notebook.FS.Read(path)
	if err != nil {
		return protocol.DidOpenTextDocumentParams{}, err
	}

	return protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        u.String(),
			LanguageID: "markdown",
			Version:    1,
			Text:       string(content),
		},
	}, nil
}

func TestServer_buildInvokedCompletionList(t *testing.T) {
	fixture := getNotebookFixture("completion")

	fs, err := fs.NewFileStorage(fixture.Path, &util.NullLogger)
	if err != nil {
		t.Fatal(err)
	}
	fixture.FS = fs

	// Setup DB
	db, err := sqlite.OpenInMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := sqlite.NewNoteIndex(fixture.Path, db, &util.NullLogger)
	config := core.NewDefaultConfig()

	// Initialize markdown parser directly
	parser := markdown.NewParser(markdown.ParserOpts{
		HashtagEnabled: true,
	}, &util.NullLogger)

	notebook := core.NewNotebook(fixture.Path, config, core.NotebookPorts{
		NoteIndex:         index,
		NoteContentParser: parser,
		FS:                fs,
		TemplateLoaderFactory: func(lang string) (core.TemplateLoader, error) {
			return &core.NullTemplateLoader, nil
		},
		Logger: &util.NullLogger,
	})
	notebook.Index(core.NoteIndexOpts{
		Force:   true,
		Verbose: false,
	})

	docStore := newDocumentStore(fs, &util.NullLogger)
	server := &Server{
		documents: docStore,
	}

	didOpenParam, err := fixture.MakeDidOpenParam("Item1.md")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := docStore.DidOpen(didOpenParam, nil)
	if err != nil {
		t.Fatal(err)
	}

	didOpenParam, err = fixture.MakeDidOpenParam("Item2.md")
	if err != nil {
		t.Fatal(err)
	}

	docItem2, err := docStore.DidOpen(didOpenParam, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Panic protection (out of bounds)", func(t *testing.T) {
		// Use a defer to catch panics
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Function panicked: %v", r)
			}
		}()

		// Test with line out of bounds
		posLineOutOfBounds := protocol.Position{Line: 100, Character: 100}
		_, err1 := server.buildInvokedCompletionList(notebook, doc, posLineOutOfBounds)
		assert.Nil(t, err1)

		// Test with character out of bounds on a valid line
		posCharOutOfBounds := protocol.Position{Line: 0, Character: 100}
		_, err2 := server.buildInvokedCompletionList(notebook, doc, posCharOutOfBounds)
		assert.Nil(t, err2)
	})

	t.Run("Tag completion trigger", func(t *testing.T) {
		// Position at #world (line 0, char 7 is just after #)
		pos := protocol.Position{Line: 0, Character: 7}
		_, err := server.buildInvokedCompletionList(notebook, doc, pos)
		assert.Nil(t, err)
	})

	t.Run("Return all link completion", func(t *testing.T) {
		// Position after [[
		pos := protocol.Position{Line: 3, Character: 2}
		item, err := server.buildInvokedCompletionList(notebook, doc, pos)
		assert.Nil(t, err)
		if len(item) < 1 {
			t.Error("Number of completion items should be greater than 0")
		}
	})

	t.Run("Return all link completion even line starts with non-ascii characters", func(t *testing.T) {
		// Position after [[
		pos := protocol.Position{Line: 3, Character: 6}
		item, err := server.buildInvokedCompletionList(notebook, docItem2, pos)
		assert.Nil(t, err)
		if len(item) < 1 {
			t.Error("Number of completion items should be greater than 0")
		}
	})
}
