package lsp

import (
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/adapter/markdown"
	"github.com/zk-org/zk/internal/adapter/sqlite"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/test/assert"
)

type mockFileStorage struct {
	core.FileStorage
}

func (m *mockFileStorage) Canonical(path string) string {
	return path
}

func TestServer_buildInvokedCompletionList(t *testing.T) {
	// Setup DB
	db, err := sqlite.OpenInMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index := sqlite.NewNoteIndex("/tmp/notebook", db, &util.NullLogger)
	config := core.NewDefaultConfig()

	// Initialize markdown parser directly
	parser := markdown.NewParser(markdown.ParserOpts{
		HashtagEnabled: true,
	}, &util.NullLogger)

	fs := &mockFileStorage{}

	notebook := core.NewNotebook("/tmp/notebook", config, core.NotebookPorts{
		NoteIndex:         index,
		NoteContentParser: parser,
		FS:                fs,
		TemplateLoaderFactory: func(lang string) (core.TemplateLoader, error) {
			return &core.NullTemplateLoader, nil
		},
		Logger: &util.NullLogger,
	})

	docStore := newDocumentStore(fs, &util.NullLogger)
	server := &Server{
		documents: docStore,
	}

	uri := "file:///tmp/notebook/note.md"
	doc, err := docStore.DidOpen(protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: "markdown",
			Version:    1,
			Text:       "Hello #world\n[[",
		},
	}, nil)
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

	t.Run("Link completion trigger", func(t *testing.T) {
		// Position after [[ (line 1, char 2)
		pos := protocol.Position{Line: 1, Character: 2}
		_, err := server.buildInvokedCompletionList(notebook, doc, pos)
		assert.Nil(t, err)
	})
}
