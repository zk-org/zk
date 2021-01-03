package cmd

import (
	"github.com/mickael-menu/zk/adapter/handlebars"
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/date"
)

type Container struct {
	Date           date.Provider
	Logger         util.Logger
	templateLoader *handlebars.Loader
}

func NewContainer() *Container {
	date := date.NewFrozenNow()

	return &Container{
		Logger: util.NewStdLogger("zk: ", 0),
		// zk is short-lived, so we freeze the current date to use the same
		// date for any rendering during the execution.
		Date: &date,
	}
}

func (c *Container) TemplateLoader() *handlebars.Loader {
	if c.templateLoader == nil {
		// FIXME take the language from the config
		handlebars.Init("en", c.Logger, c.Date)
		c.templateLoader = handlebars.NewLoader()
	}
	return c.templateLoader
}

// Database returns the DB instance for the given slip box, after executing any
// pending migration.
func (c *Container) Database(zk *zk.Zk) (*sqlite.DB, error) {
	db, err := sqlite.Open(zk.DBPath())
	if err != nil {
		return nil, err
	}
	err = db.Migrate()
	return db, err
}
