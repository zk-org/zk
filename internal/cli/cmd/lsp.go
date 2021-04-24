package cmd

import (
	"github.com/mickael-menu/zk/internal/adapter/lsp"
	"github.com/mickael-menu/zk/internal/cli"
	"github.com/mickael-menu/zk/internal/util/opt"
)

// LSP starts a server implementing the Language Server Protocol.
type LSP struct {
	Log string `hidden type:path placeholder:PATH help:"Absolute path to the log file"`
}

func (cmd *LSP) Run(container *cli.Container) error {
	server := lsp.NewServer(lsp.ServerOpts{
		Name:      "zk",
		Version:   container.Version,
		Logger:    container.Logger,
		LogFile:   opt.NewNotEmptyString(cmd.Log),
		Notebooks: container.Notebooks,
		FS:        container.FS,
	})

	return server.Run()
}
