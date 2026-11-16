//go:build !zk_kong

package main

import (
	"github.com/zk-org/zk/internal/cli/frontend/cobracli"
	"github.com/zk-org/zk/internal/cli/spec"
)

// runFrontend drives zk's commands with Cobra, which is what zk is built
// with. Cobra knows the whole command tree, so it can also generate the shell
// completion scripts `zk completion` prints.
//
// Build with `-tags zk_kong` to use the Kong front-end instead.
func runFrontend(app *spec.App, args []string, before spec.Hook) error {
	return cobracli.Run(app, args, before)
}
