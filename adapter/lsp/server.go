package lsp

import (
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspserv "github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"
	_ "github.com/tliron/kutil/logging/simple"
)

// Server holds the state of the Language Server.
type Server struct {
	server             *glspserv.Server
	initialized        bool
	clientCapabilities *protocol.ClientCapabilities
}

// ServerOpts holds the options to create a new Server.
type ServerOpts struct {
	Name    string
	Version string
	LogFile opt.String
}

// NewServer creates a new Server instance.
func NewServer(opts ServerOpts) *Server {
	handler := protocol.Handler{}
	debug := !opts.LogFile.IsNull()
	server := &Server{
		server: glspserv.NewServer(&handler, opts.Name, debug),
	}

	if debug {
		logging.Configure(10, opts.LogFile.Value)
	}

	handler.Initialize = func(context *glsp.Context, params *protocol.InitializeParams) (interface{}, error) {
		server.clientCapabilities = &params.Capabilities

		// To see the logs with coc.nvim, run :CocCommand workspace.showOutput
		// https://github.com/neoclide/coc.nvim/wiki/Debug-language-server#using-output-channel
		if params.Trace != nil {
			protocol.SetTraceValue(*params.Trace)
		}

		capabilities := handler.CreateServerCapabilities()
		capabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull
		capabilities.DocumentLinkProvider = &protocol.DocumentLinkOptions{
			ResolveProvider: boolPtr(true),
		}
		capabilities.CompletionProvider = &protocol.CompletionOptions{
			ResolveProvider:   boolPtr(true),
			TriggerCharacters: []string{"#"},
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
		server.initialized = true
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

	// handler.TextDocumentCompletion = textDocumentCompletion

	return server
}

// Run starts the Language Server in stdio mode.
func (s *Server) Run() error {
	return errors.Wrap(s.server.RunStdio(), "lsp")
}

func boolPtr(v bool) *bool {
	b := v
	return &b
}
