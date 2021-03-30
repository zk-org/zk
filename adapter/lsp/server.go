package lsp

import (
	"fmt"
	"strings"

	"github.com/mickael-menu/zk/adapter"
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	strutil "github.com/mickael-menu/zk/util/strings"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspserv "github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"
	_ "github.com/tliron/kutil/logging/simple"
)

// Server holds the state of the Language Server.
type Server struct {
	server    *glspserv.Server
	container *adapter.Container
	documents map[protocol.DocumentUri]document
}

// ServerOpts holds the options to create a new Server.
type ServerOpts struct {
	Name      string
	Version   string
	LogFile   opt.String
	Container *adapter.Container
}

// NewServer creates a new Server instance.
func NewServer(opts ServerOpts) *Server {
	debug := !opts.LogFile.IsNull()
	if debug {
		logging.Configure(10, opts.LogFile.Value)
	}

	workspace := newWorkspace()
	handler := protocol.Handler{}
	server := &Server{
		server:    glspserv.NewServer(&handler, opts.Name, debug),
		container: opts.Container,
		documents: map[string]document{},
	}

	handler.Initialize = func(context *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
		// clientCapabilities = &params.Capabilities

		if len(params.WorkspaceFolders) > 0 {
			for _, f := range params.WorkspaceFolders {
				workspace.addFolder(f.URI)
			}
		} else if params.RootURI != nil {
			workspace.addFolder(*params.RootURI)
		} else if params.RootPath != nil {
			workspace.addFolder(*params.RootPath)
		}

		server.container.OpenNotebook(workspace.folders)

		// To see the logs with coc.nvim, run :CocCommand workspace.showOutput
		// https://github.com/neoclide/coc.nvim/wiki/Debug-language-server#using-output-channel
		if params.Trace != nil {
			protocol.SetTraceValue(*params.Trace)
		}

		capabilities := handler.CreateServerCapabilities()
		capabilities.HoverProvider = true

		zk, err := server.container.Zk()
		if err == nil {
			capabilities.TextDocumentSync = protocol.TextDocumentSyncKindIncremental
			capabilities.DocumentLinkProvider = &protocol.DocumentLinkOptions{
				ResolveProvider: boolPtr(true),
			}

			triggerChars := []string{"["}

			// Setup tag completion trigger characters
			if zk.Config.Format.Markdown.Hashtags {
				triggerChars = append(triggerChars, "#")
			}
			if zk.Config.Format.Markdown.ColonTags {
				triggerChars = append(triggerChars, ":")
			}

			capabilities.CompletionProvider = &protocol.CompletionOptions{
				TriggerCharacters: triggerChars,
			}
		}

		return protocol.InitializeResult{
			Capabilities: capabilities,
			ServerInfo: &protocol.InitializeResultServerInfo{
				Name:    opts.Name,
				Version: &opts.Version,
			},
		}, nil
	}

	handler.Initialized = func(context *glsp.Context, params *protocol.InitializedParams) error {
		return nil
	}

	handler.Shutdown = func(context *glsp.Context) error {
		protocol.SetTraceValue(protocol.TraceValueOff)
		return nil
	}

	handler.SetTrace = func(context *glsp.Context, params *protocol.SetTraceParams) error {
		protocol.SetTraceValue(params.Value)
		return nil
	}

	handler.WorkspaceDidChangeWorkspaceFolders = func(context *glsp.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
		for _, f := range params.Event.Added {
			workspace.addFolder(f.URI)
		}
		for _, f := range params.Event.Removed {
			workspace.removeFolder(f.URI)
		}
		return nil
	}

	handler.TextDocumentDidOpen = func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		langID := params.TextDocument.LanguageID
		if langID != "markdown" && langID != "vimwiki" {
			return nil
		}

		server.documents[params.TextDocument.URI] = document{
			URI:     params.TextDocument.URI,
			Content: params.TextDocument.Text,
			Log:     server.server.Log,
		}

		return nil
	}

	handler.TextDocumentDidChange = func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		doc, ok := server.documents[params.TextDocument.URI]
		if !ok {
			return nil
		}

		server.documents[params.TextDocument.URI] = doc.ApplyChanges(params.ContentChanges)
		return nil
	}

	handler.TextDocumentDidClose = func(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
		delete(server.documents, params.TextDocument.URI)
		return nil
	}

	handler.TextDocumentDidSave = func(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
		return nil
	}

	handler.TextDocumentCompletion = func(context *glsp.Context, params *protocol.CompletionParams) (interface{}, error) {
		triggerChar := params.Context.TriggerCharacter

		if params.Context.TriggerKind != protocol.CompletionTriggerKindTriggerCharacter || triggerChar == nil {
			return nil, nil
		}

		switch *triggerChar {
		case "#", ":":
			return server.buildTagCompletionList(*triggerChar)

		case "[":
			doc, ok := server.documents[params.TextDocument.URI]
			if !ok {
				return nil, nil
			}
			server.server.Log.Info("word at " + doc.WordAt(int(params.Position.Line), int(params.Position.Character)))
			return server.buildLinkCompletionList(params)
		}

		return nil, nil
	}

	handler.TextDocumentHover = func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		return nil, nil
	}

	return server
}

// Run starts the Language Server in stdio mode.
func (s *Server) Run() error {
	return errors.Wrap(s.server.RunStdio(), "lsp")
}

func (s *Server) buildTagCompletionList(triggerChar string) ([]protocol.CompletionItem, error) {
	zk, err := s.container.Zk()
	if err != nil {
		return nil, err
	}
	db, _, err := s.container.Database(false)
	if err != nil {
		return nil, err
	}

	var tags []note.Collection
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		tags, err = sqlite.NewCollectionDAO(tx, s.container.Logger).FindAll(note.CollectionKindTag)
		return err
	})
	if err != nil {
		return nil, err
	}

	var items []protocol.CompletionItem
	for _, tag := range tags {
		items = append(items, protocol.CompletionItem{
			Label:      tag.Name,
			InsertText: s.buildInsertForTag(tag.Name, triggerChar, zk.Config),
			Detail:     stringPtr(fmt.Sprintf("%d %s", tag.NoteCount, strutil.Pluralize("note", tag.NoteCount))),
		})
	}

	return items, nil
}

func (s *Server) buildInsertForTag(name string, triggerChar string, config zk.Config) *string {
	switch triggerChar {
	case ":":
		name += ":"
	case "#":
		if strings.Contains(name, " ") {
			if config.Format.Markdown.MultiwordTags {
				name += "#"
			} else {
				name = strings.ReplaceAll(name, " ", "\\ ")
			}
		}
	}
	return &name
}

func (s *Server) buildLinkCompletionList(params *protocol.CompletionParams) ([]protocol.CompletionItem, error) {
	db, _, err := s.container.Database(false)
	if err != nil {
		return nil, err
	}

	var notes []note.Match
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		finder := sqlite.NewNoteDAO(tx, s.container.Logger)
		notes, err = finder.Find(note.FinderOpts{})
		return err
	})
	if err != nil {
		return nil, err
	}

	var items []protocol.CompletionItem
	for _, note := range notes {
		items = append(items, protocol.CompletionItem{
			Label: note.Title,
			Documentation: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: note.RawContent,
			},
		})
	}

	return items, nil
}

func boolPtr(v bool) *bool {
	b := v
	return &b
}

func stringPtr(v string) *string {
	s := v
	return &s
}