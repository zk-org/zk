package cmd

import (
	"github.com/zk-org/zk/internal/adapter/lsp"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/opt"
)

// LSP starts a server implementing the Language Server Protocol.
type LSP struct {
	Log string `hidden type:path placeholder:PATH help:"Absolute path to the log file"`
}

func (cmd *LSP) Run(container *cli.Container) error {
	server := lsp.NewServer(lsp.ServerOpts{
		Name:           "zk",
		Version:        container.Version,
		Logger:         container.Logger,
		LogFile:        opt.NewNotEmptyString(cmd.Log),
		Notebooks:      container.Notebooks,
		TemplateLoader: container.TemplateLoader,
		FS:             container.FS,
	})

	return server.Run()
}
