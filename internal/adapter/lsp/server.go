package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspserv "github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"
	_ "github.com/tliron/kutil/logging/simple"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/opt"
	strutil "github.com/zk-org/zk/internal/util/strings"
)

// Server holds the state of the Language Server.
type Server struct {
	server                 *glspserv.Server
	notebooks              *core.NotebookStore
	documents              *documentStore
	noteContentParser      core.NoteContentParser
	templateLoader         core.TemplateLoader
	fs                     core.FileStorage
	logger                 util.Logger
	useAdditionalTextEdits opt.Bool
}

// ServerOpts holds the options to create a new Server.
type ServerOpts struct {
	Name           string
	Version        string
	LogFile        opt.String
	Logger         *util.ProxyLogger
	Notebooks      *core.NotebookStore
	TemplateLoader core.TemplateLoader
	FS             core.FileStorage
}

// NewServer creates a new Server instance.
func NewServer(opts ServerOpts) *Server {
	fs := opts.FS
	debug := !opts.LogFile.IsNull()
	if debug {
		logging.Configure(10, opts.LogFile.Value)
	}

	handler := protocol.Handler{}
	glspServer := glspserv.NewServer(&handler, opts.Name, debug)

	// Redirect zk's logger to GLSP's to avoid breaking the JSON-RPC protocol
	// with unwanted output.
	if opts.Logger != nil {
		opts.Logger.Logger = newGlspLogger(glspServer.Log)
	}

	server := &Server{
		server:                 glspServer,
		notebooks:              opts.Notebooks,
		documents:              newDocumentStore(fs, opts.Logger),
		templateLoader:         opts.TemplateLoader,
		fs:                     fs,
		logger:                 opts.Logger,
		useAdditionalTextEdits: opt.NullBool,
	}

	var clientCapabilities protocol.ClientCapabilities

	handler.Initialize = func(context *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
		clientCapabilities = params.Capabilities

		// To see the logs with coc.nvim, run :CocCommand workspace.showOutput
		// https://github.com/neoclide/coc.nvim/wiki/Debug-language-server#using-output-channel
		if params.Trace != nil {
			protocol.SetTraceValue(*params.Trace)
		}

		if params.ClientInfo != nil {
			if params.ClientInfo.Name == "Visual Studio Code" {
				// Visual Studio Code doesn't seem to support inl
				// VSCode doesn't support deleting the trigger characters with
				// the main TextEdit. We'll use additional text edits instead.
				server.useAdditionalTextEdits = opt.True
			}
		}

		capabilities := handler.CreateServerCapabilities()
		capabilities.HoverProvider = true
		capabilities.DefinitionProvider = true
		capabilities.CodeActionProvider = true

		change := protocol.TextDocumentSyncKindIncremental
		capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{
			OpenClose: boolPtr(true),
			Change:    &change,
			Save:      boolPtr(true),
		}
		capabilities.DocumentLinkProvider = &protocol.DocumentLinkOptions{
			ResolveProvider: boolPtr(true),
		}

		triggerChars := []string{"(", "[", "#", ":"}

		capabilities.ExecuteCommandProvider = &protocol.ExecuteCommandOptions{
			Commands: []string{
				cmdIndex,
				cmdNew,
				cmdList,
				cmdTagList,
			},
		}
		capabilities.CompletionProvider = &protocol.CompletionOptions{
			TriggerCharacters: triggerChars,
			ResolveProvider:   boolPtr(true),
		}

		capabilities.ReferencesProvider = &protocol.ReferenceOptions{}

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

	handler.TextDocumentDidOpen = func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		doc, err := server.documents.DidOpen(*params, context.Notify)
		if err != nil {
			return err
		}
		if doc != nil {
			server.refreshDiagnosticsOfDocument(doc, context.Notify, false)
		}
		return nil
	}

	handler.TextDocumentDidChange = func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil
		}

		doc.ApplyChanges(params.ContentChanges)
		server.refreshDiagnosticsOfDocument(doc, context.Notify, true)
		return nil
	}

	handler.TextDocumentDidClose = func(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
		server.documents.Close(params.TextDocument.URI)
		return nil
	}

	handler.TextDocumentDidSave = func(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil
		}

		notebook, err := server.notebookOf(doc)
		if err != nil {
			server.logger.Err(err)
			return nil
		}

		_, err = notebook.Index(core.NoteIndexOpts{})
		server.logger.Err(err)
		return nil
	}

	handler.TextDocumentCompletion = func(context *glsp.Context, params *protocol.CompletionParams) (interface{}, error) {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil, nil
		}

		notebook, err := server.notebookOf(doc)
		if err != nil {
			return nil, err
		}

		if params.Context != nil && params.Context.TriggerKind == protocol.CompletionTriggerKindInvoked {
			return server.buildInvokedCompletionList(notebook, doc, params.Position)
		} else {
			return server.buildTriggerCompletionList(notebook, doc, params.Position)
		}
	}

	handler.CompletionItemResolve = func(context *glsp.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
		if path, ok := params.Data.(string); ok {
			content, err := os.ReadFile(path)
			if err != nil {
				return params, err
			}
			params.Documentation = protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: string(content),
			}
		}

		return params, nil
	}

	handler.TextDocumentHover = func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil, nil
		}

		link, err := doc.DocumentLinkAt(params.Position)
		if link == nil || err != nil {
			return nil, err
		}

		notebook, err := server.notebookOf(doc)
		if err != nil {
			return nil, err
		}

		target, err := server.noteForLink(*link, notebook)
		if err != nil || target == nil {
			return nil, err
		}

		path, err := uriToPath(target.URI)
		if err != nil {
			server.logger.Printf("unable to parse URI: %v", err)
			return nil, err
		}
		path = fs.Canonical(path)

		contents, err := os.ReadFile(path)
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
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil, nil
		}

		links, err := doc.DocumentLinks()
		if err != nil {
			return nil, err
		}

		notebook, err := server.notebookOf(doc)
		if err != nil {
			return nil, err
		}

		documentLinks := []protocol.DocumentLink{}
		for _, link := range links {
			var target string
			if strutil.IsURL(link.Href) {
				// External link
				target = link.Href
			} else {
				// Internal note link
				targetNote, err := server.noteForLink(link, notebook)
				if targetNote != nil && err == nil {
					target = targetNote.URI
				}
			}

			if target != "" {
				documentLinks = append(documentLinks, protocol.DocumentLink{
					Range:  link.Range,
					Target: &target,
				})
			}
		}

		return documentLinks, err
	}

	handler.TextDocumentDefinition = func(context *glsp.Context, params *protocol.DefinitionParams) (interface{}, error) {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil, nil
		}

		link, err := doc.DocumentLinkAt(params.Position)
		if link == nil || err != nil {
			return nil, err
		}

		notebook, err := server.notebookOf(doc)
		if err != nil {
			return nil, err
		}

		target, err := server.noteForLink(*link, notebook)
		if link == nil || target == nil || err != nil {
			return nil, err
		}

		// FIXME: Waiting for https://github.com/tliron/glsp/pull/3 to be
		// merged before using LocationLink.
		if false && isTrue(clientCapabilities.TextDocument.Definition.LinkSupport) {
			return protocol.LocationLink{
				OriginSelectionRange: &link.Range,
				TargetURI:            target.URI,
			}, nil
		} else {
			return protocol.Location{
				URI: target.URI,
			}, nil
		}
	}

	handler.WorkspaceExecuteCommand = func(context *glsp.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {

		openNotebook := func() (*core.Notebook, error) {
			args := params.Arguments
			if len(args) == 0 {
				return nil, fmt.Errorf("%s expects a notebook path as first argument", params.Command)
			}
			path, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("%s expects a notebook path as first argument, got: %v", params.Command, args[0])
			}

			return server.notebooks.Open(path)
		}

		switch params.Command {
		case cmdIndex:
			nb, err := openNotebook()
			if err != nil {
				return nil, err
			}
			return executeCommandIndex(nb, params.Arguments)

		case cmdNew:
			nb, err := openNotebook()
			if err != nil {
				return nil, err
			}
			return executeCommandNew(nb, server.documents, context, params.Arguments)

		case cmdLink:
			nb, err := openNotebook()
			if err != nil {
				return nil, err
			}
			return executeCommandLink(nb, server.documents, context, params.Arguments)

		case cmdList:
			nb, err := openNotebook()
			if err != nil {
				return nil, err
			}
			return executeCommandList(server.logger, nb, params.Arguments)

		case cmdTagList:
			nb, err := openNotebook()
			if err != nil {
				return nil, err
			}
			return executeCommandTagList(server.logger, nb, params.Arguments)

		default:
			return nil, fmt.Errorf("unknown zk LSP command: %s", params.Command)
		}
	}

	handler.TextDocumentCodeAction = func(context *glsp.Context, params *protocol.CodeActionParams) (interface{}, error) {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil, nil
		}
		wd := filepath.Dir(doc.Path)

		actions := []protocol.CodeAction{}

		missingBacklinkActions := server.getMissingBacklinkCodeActions(doc, params.TextDocument.URI, params.Range)
		actions = append(actions, missingBacklinkActions...)

		// Only add "New note" actions if range is not empty.
		if !isRangeEmpty(params.Range) {
			addAction := func(dir string, actionTitle string) error {
				opts := cmdNewOpts{
					Title: doc.ContentAtRange(params.Range),
					Dir:   dir,
					InsertLinkAtLocation: &protocol.Location{
						URI:   params.TextDocument.URI,
						Range: params.Range,
					},
				}

				var jsonOpts map[string]interface{}
				err := unmarshalJSON(opts, &jsonOpts)
				if err != nil {
					return err
				}

				actions = append(actions, protocol.CodeAction{
					Title: actionTitle,
					Kind:  stringPtr(protocol.CodeActionKindRefactor),
					Command: &protocol.Command{
						Title:     actionTitle,
						Command:   cmdNew,
						Arguments: []interface{}{wd, jsonOpts},
					},
				})

				return nil
			}

			addAction(wd, "New note in current directory")
			addAction("", "New note in top directory")
		}

		return actions, nil
	}

	handler.TextDocumentReferences = func(context *glsp.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
		doc, ok := server.documents.Get(params.TextDocument.URI)
		if !ok {
			return nil, nil
		}

		notebook, err := server.notebookOf(doc)
		if err != nil {
			return nil, err
		}

		link, err := doc.DocumentLinkAt(params.Position)
		if err != nil {
			return nil, err
		}
		if link == nil {
			link, err = doc.LinkFromRoot(notebook)
			if err != nil {
				return nil, err
			}
		}

		target, err := server.noteForLink(*link, notebook)
		if link == nil || target == nil || err != nil {
			return nil, err
		}

		opts := core.NoteFindOpts{
			LinkTo: &core.LinkFilter{Hrefs: []string{target.Path}},
		}

		notes, err := notebook.FindNotes(opts)
		if err != nil {
			return nil, err
		}

		var locations []protocol.Location

		for _, note := range notes {
			pos := strings.Index(note.RawContent, target.Path[0:len(target.Path)-3])
			var line uint32 = 0
			if pos < 0 {
				line = 0
			} else {
				linePos := strings.Count(note.RawContent[0:pos], "\n")
				line = uint32(linePos)
			}

			locations = append(locations, protocol.Location{
				URI: pathToURI(filepath.Join(notebook.Path, note.Path)),
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      line,
						Character: 0,
					},
					End: protocol.Position{
						Line:      line,
						Character: 0,
					},
				},
			})
		}

		return locations, nil
	}

	return server
}

// Run starts the Language Server in stdio mode.
func (s *Server) Run() error {
	return errors.Wrap(s.server.RunStdio(), "lsp")
}

func (s *Server) notebookOf(doc *document) (*core.Notebook, error) {
	return s.notebooks.Open(doc.Path)
}

// noteForLink returns the Note object for the note targeted by the given link.
//
// Match by order of precedence:
//  1. Prefix of relative path
//  2. Find any occurrence of the href in a note path (substring)
//  3. Match the href as a term in the note titles
func (s *Server) noteForLink(link documentLink, notebook *core.Notebook) (*Note, error) {
	note, err := s.noteForHref(link.Href, link.RelativeToDir, notebook)
	if note == nil && err == nil && link.IsWikiLink {
		// Try to find a partial href match.
		note, err = notebook.FindByHref(link.Href, true)
	}
	if note == nil || err != nil {
		return nil, err
	}

	joined_path := filepath.Join(notebook.Path, note.Path)
	return &Note{*note, pathToURI(joined_path)}, nil
}

// noteForHref returns the Note object for the note targeted by the given HREF
// relative to relativeToDir.
func (s *Server) noteForHref(href string, relativeToDir string, notebook *core.Notebook) (*core.MinimalNote, error) {
	if strutil.IsURL(href) {
		return nil, nil
	}

	path := href
	if relativeToDir != "" {
		path = filepath.Clean(filepath.Join(relativeToDir, path))
	}
	path, err := filepath.Rel(notebook.Path, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve href: %s", href)
	}
	note, err := notebook.FindByHref(path, false)
	if err != nil {
		s.logger.Printf("findByHref(%s): %s", href, err.Error())
	}
	return note, err
}

type Note struct {
	core.MinimalNote
	URI protocol.DocumentUri
}

func (s *Server) refreshDiagnosticsOfDocument(doc *document, notify glsp.NotifyFunc, delay bool) {
	if doc.NeedsRefreshDiagnostics { // Already refreshing
		return
	}

	notebook, err := s.notebookOf(doc)
	if err != nil {
		s.logger.Err(err)
		return
	}

	diagConfig := notebook.Config.LSP.Diagnostics
	if !diagConfig.IsEnabled() {
		return
	}

	doc.NeedsRefreshDiagnostics = true
	go func() {
		if delay {
			time.Sleep(1 * time.Second)
		}
		doc.NeedsRefreshDiagnostics = false

		diagnostics := []protocol.Diagnostic{}
		links, err := doc.DocumentLinks()
		if err != nil {
			s.logger.Err(err)
			return
		}

		for _, link := range links {
			if strutil.IsURL(link.Href) {
				continue
			}
			target, err := s.noteForLink(link, notebook)
			if err != nil {
				s.logger.Err(err)
				continue
			}

			var severity protocol.DiagnosticSeverity
			var message string
			if target == nil {
				if diagConfig.DeadLink == core.LSPDiagnosticNone {
					continue
				}
				severity = protocol.DiagnosticSeverity(diagConfig.DeadLink)
				message = "not found"
			} else {
				if diagConfig.WikiTitle == core.LSPDiagnosticNone {
					continue
				}
				severity = protocol.DiagnosticSeverity(diagConfig.WikiTitle)
				message = target.Title
			}

			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range:    link.Range,
				Severity: &severity,
				Source:   stringPtr("zk"),
				Message:  message,
			})
		}

		if diagConfig.MissingBacklink.Level != core.LSPDiagnosticNone {
			backlinks := s.getMissingBacklinkDiagnostics(doc, notebook, diagConfig.MissingBacklink)
			diagnostics = append(diagnostics, backlinks...)
		}

		go notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: diagnostics,
		})
	}()
}

// buildInvokedCompletionList builds the completion item response for a
// completion started automatically when typing an identifier, or manually.
func (s *Server) buildInvokedCompletionList(notebook *core.Notebook, doc *document, position protocol.Position) ([]protocol.CompletionItem, error) {
	currentWord := doc.WordAt(position)
	// currentWord will include parentheses (( but not brackets [[.
	// We check both if it's at len(currentWord) word and len(currentWord)-2 to account for auto-pair
	if strings.HasPrefix(doc.LookBehind(position, len(currentWord)+2), "[[") ||
		strings.HasPrefix(doc.LookBehind(position, len(currentWord)), "((") ||
		(len(currentWord) >= 2 && strings.HasPrefix(doc.LookBehind(position, len(currentWord)-2), "((")) {
		return s.buildLinkCompletionList(notebook, doc, position)
	}

	if doc.IsTagPosition(position, notebook.Parser) {
		return s.buildTagCompletionList(notebook, doc.WordAt(position))
	}

	return nil, nil
}

// buildTriggerCompletionList builds the completion item response for a
// completion started with a trigger character.
func (s *Server) buildTriggerCompletionList(notebook *core.Notebook, doc *document, position protocol.Position) ([]protocol.CompletionItem, error) {
	// We don't use the context because clients might not send it. Instead,
	// we'll look for trigger patterns in the document.
	switch doc.LookBehind(position, 3) {
	case "]((":
		return s.buildLinkCompletionList(notebook, doc, position)
	}

	switch doc.LookBehind(position, 2) {
	case "[[":
		return s.buildLinkCompletionList(notebook, doc, position)
	}

	switch doc.LookBehind(position, 1) {
	case "#":
		if notebook.Config.Format.Markdown.Hashtags {
			return s.buildTagCompletionList(notebook, "#")
		}
	case ":":
		if notebook.Config.Format.Markdown.ColonTags {
			return s.buildTagCompletionList(notebook, ":")
		}
	}

	return nil, nil
}

func (s *Server) buildTagCompletionList(notebook *core.Notebook, prefix string) ([]protocol.CompletionItem, error) {
	tags, err := notebook.FindCollections(core.CollectionKindTag, nil)
	if err != nil {
		return nil, err
	}

	var items []protocol.CompletionItem
	for _, tag := range tags {
		items = append(items, protocol.CompletionItem{
			Label:      tag.Name,
			InsertText: s.buildInsertForTag(tag.Name, prefix, notebook.Config),
			Detail:     stringPtr(fmt.Sprintf("%d %s", tag.NoteCount, strutil.Pluralize("note", tag.NoteCount))),
		})
	}

	return items, nil
}

func (s *Server) buildInsertForTag(name string, prefix string, config core.Config) *string {
	switch prefix {
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

func (s *Server) buildLinkCompletionList(notebook *core.Notebook, doc *document, position protocol.Position) ([]protocol.CompletionItem, error) {
	linkFormatter, err := newLinkFormatter(notebook, doc, position)
	if err != nil {
		return nil, err
	}

	templates, err := newCompletionTemplates(s.templateLoader, notebook.Config.LSP.Completion.Note)
	if err != nil {
		return nil, err
	}

	notes, err := notebook.FindMinimalNotes(core.NoteFindOpts{})
	if err != nil {
		return nil, err
	}

	var items []protocol.CompletionItem
	for _, note := range notes {
		item, err := s.newCompletionItem(notebook, note, doc, position, linkFormatter, templates)
		if err != nil {
			s.logger.Err(err)
			continue
		}

		items = append(items, item)
	}

	return items, nil
}

func newLinkFormatter(notebook *core.Notebook, doc *document, position protocol.Position) (core.LinkFormatter, error) {
	if doc.LookBehind(position, 3) == "]((" {
		return core.NewMarkdownLinkFormatter(notebook.Config.Format.Markdown, true)
	} else {
		return notebook.NewLinkFormatter()
	}
}

func (s *Server) newCompletionItem(notebook *core.Notebook, note core.MinimalNote, doc *document, pos protocol.Position, linkFormatter core.LinkFormatter, templates completionTemplates) (protocol.CompletionItem, error) {
	kind := protocol.CompletionItemKindReference
	item := protocol.CompletionItem{
		Kind: &kind,
		Data: filepath.Join(notebook.Path, note.Path),
	}

	templateContext, err := newCompletionItemRenderContext(note, notebook.Path, doc.Path)
	if err != nil {
		return item, err
	}

	if templates.Label != nil {
		item.Label, err = templates.Label.Render(templateContext)
		if err != nil {
			return item, err
		}
	} else {
		item.Label = note.Title
	}
	// Fallback on the note path to never have empty labels.
	if item.Label == "" {
		item.Label = note.Path
	}

	if templates.FilterText != nil {
		filterText, err := templates.FilterText.Render(templateContext)
		if err != nil {
			return item, err
		}
		item.FilterText = &filterText
	}
	if item.FilterText == nil || *item.FilterText == "" {
		// Add the path to the filter text to be able to complete by it.
		item.FilterText = stringPtr(item.Label + " " + note.Path)
	}

	if templates.Detail != nil {
		detail, err := templates.Detail.Render(templateContext)
		if err != nil {
			return item, err
		}
		item.Detail = &detail
	}

	item.TextEdit, err = s.newTextEditForLink(notebook, note, doc, pos, linkFormatter)
	if err != nil {
		err = errors.Wrapf(err, "failed to build TextEdit for note at %s", note.Path)
		return item, err
	}

	if s.useAdditionalTextEditsWithNotebook(notebook) {
		addTextEdits := []protocol.TextEdit{}

		startOffset := -2
		if (doc.LookBehind(pos, 2) != "[[") && (doc.LookBehind(pos, 2) != "((") {
			currentWord := doc.WordAt(pos)
			startOffset = -2 - len(currentWord)
		}

		// Some LSP clients (e.g. VSCode) don't support deleting the trigger
		// characters with the main TextEdit. So let's add an additional
		// TextEdit for that.
		addTextEdits = append(addTextEdits, protocol.TextEdit{
			NewText: "",
			Range:   rangeFromPosition(pos, startOffset, 0),
		})

		item.AdditionalTextEdits = addTextEdits
	}

	return item, nil
}

func (s *Server) newTextEditForLink(notebook *core.Notebook, note core.MinimalNote, doc *document, pos protocol.Position, linkFormatter core.LinkFormatter) (interface{}, error) {
	path := core.NotebookPath{
		Path:       note.Path,
		BasePath:   notebook.Path,
		WorkingDir: filepath.Dir(doc.Path),
	}
	context, err := core.NewLinkFormatterContext(path, note.Title, note.Metadata)
	if err != nil {
		return nil, err
	}
	link, err := linkFormatter(context)
	if err != nil {
		return nil, err
	}

	// Overwrite [[ trigger directly if the additional text edits are disabled.
	startOffset := 0
	if !s.useAdditionalTextEditsWithNotebook(notebook) {
		if (doc.LookBehind(pos, 2) == "[[") || (doc.LookBehind(pos, 2) == "((") {
			startOffset = -2
		} else {
			currentWord := doc.WordAt(pos)
			startOffset = -2 - len(currentWord)
		}
	}

	// Some LSP clients (e.g. VSCode) auto-pair brackets, so we need to
	// remove the closing ]] or )) after the completion.
	endOffset := 0
	suffix := doc.LookForward(pos, 2)
	if suffix == "]]" || suffix == "))" {
		endOffset = 2
	}

	return protocol.TextEdit{
		NewText: link,
		Range:   rangeFromPosition(pos, startOffset, endOffset),
	}, nil
}

func (s *Server) useAdditionalTextEditsWithNotebook(nb *core.Notebook) bool {
	return nb.Config.LSP.Completion.UseAdditionalTextEdits.
		Or(s.useAdditionalTextEdits).
		OrBool(false).
		Unwrap()
}

func positionInRange(content string, rng protocol.Range, pos protocol.Position) bool {
	start, end := rng.IndexesIn(content)
	i := pos.IndexIn(content)
	return i >= start && i <= end
}

func rangeFromPosition(pos protocol.Position, startOffset, endOffset int) protocol.Range {
	offsetPos := func(offset int) protocol.Position {
		newPos := pos
		if offset < 0 {
			newPos.Character -= uint32(-offset)
		} else {
			newPos.Character += uint32(offset)
		}
		return newPos
	}

	return protocol.Range{
		Start: offsetPos(startOffset),
		End:   offsetPos(endOffset),
	}
}

func isRangeEmpty(pos protocol.Range) bool {
	return pos.Start == pos.End
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

func unmarshalJSON(obj interface{}, v interface{}) error {
	js, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return json.Unmarshal(js, v)
}

func toBool(obj interface{}) bool {
	s := strings.ToLower(fmt.Sprint(obj))
	return s == "true" || s == "1"
}

// MissingBacklink contains information of a note that links to the current note.
type MissingBacklink struct {
	SourcePath  string
	SourceTitle string
}

// findMissingBacklinks finds notes that link to the current note but are not linked back.
func (s *Server) findMissingBacklinks(currentNoteID core.NoteID, notebook *core.Notebook, doc *document) ([]MissingBacklink, error) {
	currentNote, err := notebook.FindMinimalNotes(core.NoteFindOpts{
		IncludeIDs: []core.NoteID{currentNoteID},
	})
	if err != nil || len(currentNote) == 0 {
		return nil, err
	}
	currentNotePath := currentNote[0].Path

	notesThatLinkToUs, err := notebook.FindMinimalNotes(core.NoteFindOpts{
		LinkTo: &core.LinkFilter{
			Hrefs: []string{currentNotePath},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(notesThatLinkToUs) == 0 {
		return nil, nil
	}

	currentDocLinks, err := doc.DocumentLinks()
	if err != nil {
		return nil, err
	}

	backlinkedNoteIDs := make(map[core.NoteID]bool)
	for _, link := range currentDocLinks {
		if strutil.IsURL(link.Href) {
			continue
		}
		// Resolve the link to find the target note
		target, err := s.noteForLink(link, notebook)
		if err != nil || target == nil {
			continue
		}
		// Extract note ID from the URI - we need to find the note by path
		targetPath, err := uriToPath(target.URI)
		if err != nil {
			continue
		}
		targetRelPath, err := notebook.RelPath(targetPath)
		if err != nil {
			continue
		}
		targetNote, err := notebook.FindByHref(targetRelPath, false)
		if err != nil || targetNote == nil {
			continue
		}
		backlinkedNoteIDs[targetNote.ID] = true
	}

	var missingBacklinks []MissingBacklink
	for _, linkingNote := range notesThatLinkToUs {
		if !backlinkedNoteIDs[linkingNote.ID] {
			missingBacklinks = append(missingBacklinks, MissingBacklink{
				SourcePath:  linkingNote.Path,
				SourceTitle: linkingNote.Title,
			})
		}
	}

	return missingBacklinks, nil
}

// getMissingBacklinkDiagnostics returns diagnostics for missing backlinks in the current document.
func (s *Server) getMissingBacklinkDiagnostics(doc *document, notebook *core.Notebook, config core.MissingBacklinkConfig) []protocol.Diagnostic {
	relPath, err := notebook.RelPath(doc.Path)
	if err != nil {
		s.logger.Err(err)
		return nil
	}

	currentNote, err := notebook.FindByHref(relPath, false)
	if err != nil {
		s.logger.Err(err)
		return nil
	}

	if currentNote == nil {
		return nil
	}

	missingBacklinks, err := s.findMissingBacklinks(currentNote.ID, notebook, doc)
	if err != nil {
		s.logger.Err(err)
		return nil
	}

	lines := strings.Split(doc.Content, "\n")

	var targetLine uint32
	switch config.Position {
	case core.LSPDiagnosticPositionTop:
		targetLine = 0
	case core.LSPDiagnosticPositionBottom:
		targetLine = uint32(len(lines) - 1)
	case core.LSPDiagnosticPositionLastSection:
		targetLine = findLastSectionLine(lines)
	default:
		targetLine = uint32(len(lines) - 1) // Fallback to bottom
	}

	var diagnostics []protocol.Diagnostic
	for _, backlink := range missingBacklinks {
		diagSeverity := protocol.DiagnosticSeverity(config.Level)

		noteTitle := backlink.SourceTitle
		if noteTitle == "" {
			noteTitle = backlink.SourcePath
		}

		// Convert path to be relative to current note's directory
		currentNoteDir := filepath.Dir(doc.Path)
		sourceNotePath := filepath.Join(notebook.Path, backlink.SourcePath)
		relativePath, err := filepath.Rel(currentNoteDir, sourceNotePath)
		if err != nil {
			// Fallback to original path if relative path calculation fails
			relativePath = backlink.SourcePath
		}

		message := fmt.Sprintf("Missing backlink to [%s](%s)", noteTitle, relativePath)

		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: targetLine, Character: 0},
				End:   protocol.Position{Line: targetLine, Character: 0},
			},
			Severity: &diagSeverity,
			Source:   stringPtr("zk"),
			Message:  message,
		})
	}

	return diagnostics
}

// findLastSectionLine finds the line number of the last heading in the document.
// Falls back to line 0 (top) if no headings are found.
func findLastSectionLine(lines []string) uint32 {
	headingRegex := regexp.MustCompile(`^#{1,6}\s+`)
	var lastHeadingLine uint32 = 0 // Fallback to top
	for i, line := range lines {
		if headingRegex.MatchString(line) {
			lastHeadingLine = uint32(i)
		}
	}
	return lastHeadingLine
}

// getMissingBacklinkCodeActions returns code actions for adding missing backlinks.
func (s *Server) getMissingBacklinkCodeActions(doc *document, docURI protocol.DocumentUri, requestRange protocol.Range) []protocol.CodeAction {
	notebook, err := s.notebookOf(doc)
	if err != nil {
		return nil
	}

	diagConfig := notebook.Config.LSP.Diagnostics
	if diagConfig.MissingBacklink.Level == core.LSPDiagnosticNone {
		return nil
	}

	relPath, err := notebook.RelPath(doc.Path)
	if err != nil {
		return nil
	}

	currentNote, err := notebook.FindByHref(relPath, false)
	if err != nil || currentNote == nil {
		return nil
	}

	missingBacklinks, err := s.findMissingBacklinks(currentNote.ID, notebook, doc)
	if err != nil || len(missingBacklinks) == 0 {
		return nil
	}

	var actions []protocol.CodeAction
	var formattedLinks []string

	lines := strings.Split(doc.Content, "\n")
	currentLine := min(requestRange.End.Line, uint32(len(lines)-1))

	insertPosition := protocol.Position{
		Line:      currentLine,
		Character: uint32(len(lines[currentLine])),
	}

	// Format all links once and generate individual actions
	for _, backlink := range missingBacklinks {
		// Full note metadata is required for NewLinkFormatterContext
		sourceNote, err := notebook.FindByHref(backlink.SourcePath, false)
		if err != nil || sourceNote == nil {
			continue
		}

		linkFormatter, err := notebook.NewLinkFormatter()
		if err != nil {
			continue
		}

		noteDir := filepath.Dir(doc.Path)
		linkPath := core.NotebookPath{
			Path:       backlink.SourcePath,
			BasePath:   notebook.Path,
			WorkingDir: noteDir,
		}
		linkContext, err := core.NewLinkFormatterContext(linkPath, backlink.SourceTitle, sourceNote.Metadata)
		if err != nil {
			continue
		}

		link, err := linkFormatter(linkContext)
		if err != nil {
			continue
		}

		// Store formatted link for reuse
		formattedLinks = append(formattedLinks, link)

		title := backlink.SourceTitle
		if title == "" {
			title = filepath.Base(backlink.SourcePath)
		}

		// Create individual action
		actions = append(actions, protocol.CodeAction{
			Title: fmt.Sprintf("Add backlink to %s", title),
			Kind:  stringPtr(protocol.CodeActionKindQuickFix),
			Edit: &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					docURI: {{
						Range: protocol.Range{
							Start: insertPosition,
							End:   insertPosition,
						},
						NewText: "\n" + link,
					}},
				},
			},
		})
	}

	// Add "add all missing backlinks" action if there are multiple backlinks
	if len(formattedLinks) > 1 {
		// Join all formatted links with newlines
		allLinksText := "\n" + strings.Join(formattedLinks, "\n")

		actions = append(actions, protocol.CodeAction{
			Title: fmt.Sprintf("Add all %d missing backlinks", len(formattedLinks)),
			Kind:  stringPtr(protocol.CodeActionKindQuickFix),
			Edit: &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentUri][]protocol.TextEdit{
					docURI: {{
						Range: protocol.Range{
							Start: insertPosition,
							End:   insertPosition,
						},
						NewText: allLinksText,
					}},
				},
			},
		})
	}

	return actions
}
