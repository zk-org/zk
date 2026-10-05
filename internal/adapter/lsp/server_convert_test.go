package lsp

import (
	"os"
	"path/filepath"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/adapter/fs"
	"github.com/zk-org/zk/internal/adapter/markdown"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
)

func TestGetConvertToFrontmatterCodeAction(t *testing.T) {
	t.Run("converts the title heading", func(t *testing.T) {
		server, doc := newConvertToFrontmatterServer(t, "# This is a title\n\nBody")

		actions := server.getConvertToFrontmatterCodeAction(doc, doc.URI, protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 0},
		})

		if len(actions) != 1 {
			t.Fatalf("len(actions) = %d, want 1", len(actions))
		}

		action := actions[0]
		if action.Title != "Convert to frontmatter" {
			t.Errorf("Title = %q, want %q", action.Title, "Convert to frontmatter")
		}

		edits := action.Edit.Changes[doc.URI]
		if len(edits) != 1 {
			t.Fatalf("len(edits) = %d, want 1", len(edits))
		}

		edit := edits[0]
		if edit.NewText != "---\ntitle: This is a title\n---\n\n" {
			t.Errorf("NewText = %q", edit.NewText)
		}
		if edit.Range.Start.Line != 0 || edit.Range.End.Line != 2 {
			t.Errorf("Range lines = %d..%d, want 0..2", edit.Range.Start.Line, edit.Range.End.Line)
		}
	})

	t.Run("no action outside the title line", func(t *testing.T) {
		server, doc := newConvertToFrontmatterServer(t, "# This is a title\n\nBody")

		actions := server.getConvertToFrontmatterCodeAction(doc, doc.URI, protocol.Range{
			Start: protocol.Position{Line: 2, Character: 0},
			End:   protocol.Position{Line: 2, Character: 0},
		})

		if len(actions) != 0 {
			t.Fatalf("len(actions) = %d, want 0", len(actions))
		}
	})

	t.Run("lowest level heading is the title", func(t *testing.T) {
		server, doc := newConvertToFrontmatterServer(t, "## Small Heading\n# Bigger Heading\n")

		actions := server.getConvertToFrontmatterCodeAction(doc, doc.URI, protocol.Range{
			Start: protocol.Position{Line: 1, Character: 0},
			End:   protocol.Position{Line: 1, Character: 0},
		})

		if len(actions) != 1 {
			t.Fatalf("len(actions) = %d, want 1", len(actions))
		}
		if got := actions[0].Edit.Changes[doc.URI][0].NewText; got != "---\ntitle: Bigger Heading\n---\n\n" {
			t.Errorf("NewText = %q", got)
		}
	})

	t.Run("setext heading is the title", func(t *testing.T) {
		server, doc := newConvertToFrontmatterServer(t, "Title\n---\n\nBody")

		actions := server.getConvertToFrontmatterCodeAction(doc, doc.URI, protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 0},
		})

		if len(actions) != 1 {
			t.Fatalf("len(actions) = %d, want 1", len(actions))
		}
		if got := actions[0].Edit.Changes[doc.URI][0].NewText; got != "---\ntitle: Title\n---\n\n" {
			t.Errorf("NewText = %q", got)
		}
	})

	t.Run("no action when frontmatter exists", func(t *testing.T) {
		server, doc := newConvertToFrontmatterServer(t, "---\ntitle: Already\n---\n# Other\n")

		actions := server.getConvertToFrontmatterCodeAction(doc, doc.URI, protocol.Range{
			Start: protocol.Position{Line: 3, Character: 0},
			End:   protocol.Position{Line: 3, Character: 0},
		})

		if len(actions) != 0 {
			t.Fatalf("len(actions) = %d, want 0", len(actions))
		}
	})

	t.Run("no action without a title heading", func(t *testing.T) {
		server, doc := newConvertToFrontmatterServer(t, "Body only\n")

		actions := server.getConvertToFrontmatterCodeAction(doc, doc.URI, protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 0},
		})

		if len(actions) != 0 {
			t.Fatalf("len(actions) = %d, want 0", len(actions))
		}
	})
}

func newConvertToFrontmatterServer(t *testing.T, content string) (*Server, *document) {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".zk"), 0o755); err != nil {
		t.Fatal(err)
	}

	notePath := filepath.Join(root, "note.md")
	if err := os.WriteFile(notePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	fstore, err := fs.NewFileStorage(root, &util.NullLogger)
	if err != nil {
		t.Fatal(err)
	}

	parser := markdown.NewParser(markdown.ParserOpts{HashtagEnabled: true}, &util.NullLogger)

	store := core.NewNotebookStore(core.NewDefaultConfig(), core.NotebookStorePorts{
		NotebookFactory: func(path string, config core.Config) (*core.Notebook, error) {
			return core.NewNotebook(path, config, core.NotebookPorts{
				NoteContentParser: parser,
				TemplateLoaderFactory: func(lang string) (core.TemplateLoader, error) {
					return &core.NullTemplateLoader, nil
				},
				FS:     fstore,
				Logger: &util.NullLogger,
			}), nil
		},
		TemplateLoader: &core.NullTemplateLoader,
		FS:             fstore,
	})

	doc := &document{
		URI:     protocol.DocumentUri(pathToURI(notePath)),
		Path:    notePath,
		Content: content,
	}

	return &Server{notebooks: store, logger: &util.NullLogger}, doc
}
