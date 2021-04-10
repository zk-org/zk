package lsp

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mickael-menu/zk/internal/adapter"
	"github.com/mickael-menu/zk/internal/adapter/sqlite"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/core/note"
	"github.com/mickael-menu/zk/internal/core/zk"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
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
	documents map[protocol.DocumentUri]*document
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
		documents: map[string]*document{},
	}

	var clientCapabilities protocol.ClientCapabilities

	handler.Initialize = func(context *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
		clientCapabilities = params.Capabilities

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

			capabilities.DefinitionProvider = boolPtr(true)
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

		server.documents[params.TextDocument.URI] = &document{
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

		doc.ApplyChanges(params.ContentChanges)
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
			return server.buildLinkCompletionList(params)
		}

		return nil, nil
	}

	handler.TextDocumentHover = func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := server.documents[params.TextDocument.URI]
		if !ok {
			return nil, nil
		}

		link, err := doc.DocumentLinkAt(params.Position)
		if link == nil || err != nil {
			return nil, err
		}

		zk, err := server.container.Zk()
		if err != nil {
			return nil, err
		}

		db, _, err := server.container.Database(false)
		if err != nil {
			return nil, err
		}

		index := sqlite.NewNoteIndex(db, server.container.Logger)
		target, err := server.targetForHref(link.Href, zk.Path, index)
		if err != nil || target == "" || strutil.IsURL(target) {
			return nil, err
		}

		target = strings.TrimPrefix(target, "file://")
		contents, err := ioutil.ReadFile(target)
		if err != nil {
			return nil, err
		}

		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: string(contents),
			},
		}, nil
	}

	handler.TextDocumentDocumentLink = func(context *glsp.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
		doc, ok := server.documents[params.TextDocument.URI]
		if !ok {
			return nil, nil
		}

		links, err := doc.DocumentLinks()
		if err != nil {
			return nil, err
		}

		zk, err := server.container.Zk()
		if err != nil {
			return nil, err
		}

		db, _, err := server.container.Database(false)
		if err != nil {
			return nil, err
		}

		index := sqlite.NewNoteIndex(db, server.container.Logger)

		documentLinks := []protocol.DocumentLink{}
		for _, link := range links {
			target, err := server.targetForHref(link.Href, zk.Path, index)
			if target == "" || err != nil {
				continue
			}

			documentLinks = append(documentLinks, protocol.DocumentLink{
				Range:  link.Range,
				Target: &target,
			})
		}

		return documentLinks, err
	}

	handler.TextDocumentDefinition = func(context *glsp.Context, params *protocol.DefinitionParams) (interface{}, error) {
		doc, ok := server.documents[params.TextDocument.URI]
		if !ok {
			return nil, nil
		}

		link, err := doc.DocumentLinkAt(params.Position)
		if link == nil || err != nil {
			return nil, err
		}

		zk, err := server.container.Zk()
		if err != nil {
			return nil, err
		}

		db, _, err := server.container.Database(false)
		if err != nil {
			return nil, err
		}

		index := sqlite.NewNoteIndex(db, server.container.Logger)
		target, err := server.targetForHref(link.Href, zk.Path, index)
		if link == nil || target == "" || err != nil {
			return nil, err
		}

		if isTrue(clientCapabilities.TextDocument.Definition.LinkSupport) {
			return protocol.LocationLink{
				OriginSelectionRange: &link.Range,
				TargetURI:            target,
			}, nil
		} else {
			return protocol.Location{
				URI: target,
			}, nil
		}
	}

	return server
}

// targetForHref returns the LSP documentUri for the note at the given HREF.
func (s *Server) targetForHref(href string, basePath string, index core.NoteIndex) (string, error) {
	if strutil.IsURL(href) {
		return href, nil
	} else {
		// FIXME:
		return "", nil
		// note, err := finder.FindByHref(href)
		// if err != nil {
		// 	s.server.Log.Errorf("findByHref(%s): %s", href, err.Error())
		// 	return "", err
		// }
		// if note == nil {
		// 	return "", nil
		// }
		// return "file://" + filepath.Join(basePath, note.Path), nil
	}
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
	zk, err := s.container.Zk()
	if err != nil {
		return nil, err
	}
	doc, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil, nil
	}

	db, _, err := s.container.Database(false)
	if err != nil {
		return nil, err
	}

	index := sqlite.NewNoteIndex(db, s.container.Logger)
	notes, err := index.Find(core.NoteFindOpts{})
	if err != nil {
		return nil, err
	}

	var items []protocol.CompletionItem
	for _, note := range notes {
		items = append(items, protocol.CompletionItem{
			Label:    note.Title,
			TextEdit: s.buildTextEditForLink(zk, note, doc, params.Position),
			Documentation: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: note.RawContent,
			},
		})
	}

	return items, nil
}

func (s *Server) buildTextEditForLink(zk *zk.Zk, note core.ContextualNote, document *document, pos protocol.Position) interface{} {
	isWikiLink := (document.LookBehind(pos, 2) == "[[")
	var text string

	path := filepath.Join(zk.Path, note.Path)
	documentPath := strings.TrimPrefix(document.URI, "file://")
	path, err := filepath.Rel(filepath.Dir(documentPath), path)
	if err != nil {
		path = note.Path
	}
	ext := filepath.Ext(path)
	path = strings.TrimSuffix(path, ext)
	if isWikiLink {
		text = path + "]]"
	} else {
		path = strings.ReplaceAll(url.PathEscape(path), "%2F", "/")
		text = note.Title + "](" + path + ")"
	}

	return protocol.TextEdit{
		Range: protocol.Range{
			Start: pos,
			End:   pos,
		},
		NewText: text,
	}
}

func positionInRange(content string, rng protocol.Range, pos protocol.Position) bool {
	start, end := rng.IndexesIn(content)
	i := pos.IndexIn(content)
	return i >= start && i <= end
}

func boolPtr(v bool) *bool {
	b := v
	return &b
}

func isTrue(v *bool) bool {
	return v != nil && *v == true
}

func isFalse(v *bool) bool {
	return v == nil || *v == false
}

func stringPtr(v string) *string {
	s := v
	return &s
}
