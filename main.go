package main

import (
	"github.com/alecthomas/kong"
	"github.com/mickael-menu/zk/cmd"
)

var cli struct {
	Init cmd.Init `cmd help:"Create a slip box in the given directory"`
	New  cmd.New  `cmd help:"Add a new note to the slip box"`
}

func main() {
	// Create the dependency graph.
	container := cmd.NewContainer()

	ctx := kong.Parse(&cli, kong.Name("zk"))
	err := ctx.Run(container)
	ctx.FatalIfErrorf(err)
}
