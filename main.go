package main

import (
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/mickael-menu/zk/adapter/handlebars"
	"github.com/mickael-menu/zk/cmd"
	"github.com/mickael-menu/zk/util/date"
)

var cli struct {
	Init cmd.Init `cmd help:"Create a slip box in the given directory"`
	New  cmd.New  `cmd help:"Add a new note to the slip box"`
}

func main() {
	logger := log.New(os.Stderr, "zk: warning: ", 0)
	// zk is short-lived, so we freeze the current date to use the same date
	// for any rendering during the execution.
	date := date.NewFrozenNow()
	// FIXME take the language from the config
	handlebars.Init("en", logger, &date)

	ctx := kong.Parse(&cli,
		kong.Name("zk"),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
