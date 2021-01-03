package main

import (
	"github.com/alecthomas/kong"
	"github.com/mickael-menu/zk/cmd"
)

var Version = "dev"
var Build = "dev"

var cli struct {
	Index   cmd.Index        `cmd help:"Index the notes in the given directory to be searchable"`
	Init    cmd.Init         `cmd help:"Create a slip box in the given directory"`
	List    cmd.List         `cmd help:"List notes matching given criteria"`
	New     cmd.New          `cmd help:"Create a new note in the given slip box directory"`
	Version kong.VersionFlag `help:"Print zk version"`
}

func main() {
	// Create the dependency graph.
	container := cmd.NewContainer()

	ctx := kong.Parse(&cli,
		kong.Name("zk"),
		kong.Vars{
			"version": Version,
		},
	)
	err := ctx.Run(container)
	ctx.FatalIfErrorf(err)
}
