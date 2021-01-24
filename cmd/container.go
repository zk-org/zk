package cmd

import (
	"io"
	"sync"

	"github.com/mickael-menu/zk/adapter/fzf"
	"github.com/mickael-menu/zk/adapter/handlebars"
	"github.com/mickael-menu/zk/adapter/markdown"
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/adapter/term"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/date"
	"github.com/mickael-menu/zk/util/pager"
)

type Container struct {
	Date           date.Provider
	Logger         util.Logger
	Terminal       *term.Terminal
	templateLoader *handlebars.Loader

	zkOnce sync.Once
	zk     *zk.Zk
	zkErr  error
}

func NewContainer() *Container {
	date := date.NewFrozenNow()

	return &Container{
		Logger: util.NewStdLogger("zk: ", 0),
		// zk is short-lived, so we freeze the current date to use the same
		// date for any rendering during the execution.
		Date:     &date,
		Terminal: term.New(),
	}
}

func (c *Container) OpenZk() (*zk.Zk, error) {
	c.zkOnce.Do(func() {
		c.zk, c.zkErr = zk.Open(".")
	})
	return c.zk, c.zkErr
}

func (c *Container) TemplateLoader(lang string) *handlebars.Loader {
	if c.templateLoader == nil {
		handlebars.Init(lang, c.Logger, c.Terminal)
		c.templateLoader = handlebars.NewLoader()
	}
	return c.templateLoader
}

func (c *Container) Parser() *markdown.Parser {
	return markdown.NewParser()
}

func (c *Container) NoteFinder(tx sqlite.Transaction) note.Finder {
	notes := sqlite.NewNoteDAO(tx, c.Logger)
	return fzf.NewNoteFinder(notes, c.Terminal)
}

// Database returns the DB instance for the given slip box, after executing any
// pending migration.
func (c *Container) Database(path string) (*sqlite.DB, error) {
	db, err := sqlite.Open(path)
	if err != nil {
		return nil, err
	}
	err = db.Migrate()
	return db, err
}

// Paginate creates an auto-closing io.Writer which will be automatically
// paginated if noPager is false, using the user's pager.
//
// You can write to the pager only in the run callback.
func (c *Container) Paginate(noPager bool, config zk.Config, run func(out io.Writer) error) error {
	pager, err := c.pager(noPager || config.NoPager, config)
	if err != nil {
		return err
	}
	err = run(pager)
	pager.Close()
	return err
}

func (c *Container) pager(noPager bool, config zk.Config) (*pager.Pager, error) {
	if noPager {
		return pager.PassthroughPager, nil
	} else {
		return pager.New(config.Pager, c.Logger)
	}
}
