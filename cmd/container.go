package cmd

import (
	"log"
	"os"

	"github.com/mickael-menu/zk/adapter/handlebars"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/date"
)

type Container struct {
	Zk       *zk.Zk
	Date     date.Provider
	Logger   util.Logger
	renderer *handlebars.HandlebarsRenderer
}

func NewContainer() *Container {
	zk, _ := zk.Open(".")
	date := date.NewFrozenNow()

	return &Container{
		Zk:     zk,
		Logger: log.New(os.Stderr, "zk: warning: ", 0),
		// zk is short-lived, so we freeze the current date to use the same
		// date for any rendering during the execution.
		Date: &date,
	}
}

func (c *Container) Renderer() *handlebars.HandlebarsRenderer {
	if c.renderer == nil {
		// FIXME take the language from the config
		handlebars.Init("en", c.Logger, c.Date)
		c.renderer = handlebars.NewRenderer()
	}
	return c.renderer
}
