package cmd

import (
	"github.com/mickael-menu/zk/adapter"
	"github.com/mickael-menu/zk/adapter/lsp"
	"github.com/mickael-menu/zk/util/opt"
)

// LSP starts a server implementing the Language Server Protocol.
type LSP struct {
	Log string `type:path placeholder:PATH help:"Absolute path to the log file"`
}

func (cmd *LSP) Run(container *adapter.Container) error {
	server := lsp.NewServer(lsp.ServerOpts{
		Name:      "zk",
		Version:   container.Version,
		LogFile:   opt.NewNotEmptyString(cmd.Log),
		Container: container,
	})

	return server.Run()
}
