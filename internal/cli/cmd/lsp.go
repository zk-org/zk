package cmd

import (
	"github.com/zk-org/zk/internal/adapter/lsp"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/util/ptr"
)

// LSP starts a server implementing the Language Server Protocol.
type LSP struct {
	Log string
}

// LSPSpec describes the `zk lsp` command. It is hidden, as it is meant for an
// editor to start, not a person.
func LSPSpec(container *cli.Container) *spec.Command {
	cmd := LSP{}

	return &spec.Command{
		Name:   "lsp",
		Hidden: true,
		Flags: []spec.Flag{
			{Name: "log", Hidden: true, Placeholder: "PATH", Target: &cmd.Log,
				Help: "Absolute path to the log file", CompleteFiles: true},
		},
		Run: func() error {
			// The log file is given to an editor's configuration, where `~`
			// and environment variables are the norm.
			if cmd.Log != "" {
				var err error
				if cmd.Log, err = expandPath(cmd.Log); err != nil {
					return err
				}
			}
			return cmd.Run(container)
		},
	}
}

func (cmd *LSP) Run(container *cli.Container) error {
	server := lsp.NewServer(lsp.ServerOpts{
		Name:           "zk",
		Version:        container.Version,
		Logger:         container.Logger,
		LogFile:        ptr.NotEmptyString(cmd.Log),
		Notebooks:      container.Notebooks,
		TemplateLoader: container.TemplateLoader,
		FS:             container.FS,
	})

	return server.Run()
}
