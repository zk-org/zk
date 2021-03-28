package cmd

import (
	"github.com/mickael-menu/zk/adapter/lsp"
	"github.com/mickael-menu/zk/util/opt"
)

// LSP starts a server implementing the Language Server Protocol.
type LSP struct{}

func (cmd *LSP) Run(container *Container) error {
	server := lsp.NewServer(lsp.ServerOpts{
		Name:    "zk",
		Version: container.Version,
		LogFile: opt.NewString("/tmp/zk-lsp.log"),
	})

	return server.Run()
}
