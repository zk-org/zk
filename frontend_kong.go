//go:build zk_kong

package main

import (
	"github.com/zk-org/zk/internal/cli/frontend/kongcli"
	"github.com/zk-org/zk/internal/cli/spec"
)

// runFrontend drives zk's commands with Kong, which zk was built with before
// it moved to Cobra. It accepts the same commands and flags, but generates no
// shell completion scripts.
//
// Build without `-tags zk_kong` to use the Cobra front-end.
func runFrontend(app *spec.App, args []string, before spec.Hook) error {
	return kongcli.Run(app, args, before)
}
