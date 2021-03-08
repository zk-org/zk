package cmd

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/mickael-menu/zk/adapter/fzf"
	"github.com/mickael-menu/zk/adapter/handlebars"
	"github.com/mickael-menu/zk/adapter/markdown"
	"github.com/mickael-menu/zk/adapter/sqlite"
	"github.com/mickael-menu/zk/adapter/term"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/date"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/pager"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/schollz/progressbar/v3"
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
		if c.zkErr == nil {
			os.Setenv("ZK_PATH", c.zk.Path)
		}
	})
	return c.zk, c.zkErr
}

func (c *Container) TemplateLoader(lang string) *handlebars.Loader {
	if c.templateLoader == nil {
		handlebars.Init(lang, c.Terminal.SupportsUTF8(), c.Logger, c.Terminal)
		c.templateLoader = handlebars.NewLoader()
	}
	return c.templateLoader
}

func (c *Container) Parser() *markdown.Parser {
	return markdown.NewParser(markdown.ParserOpts{
		HashtagEnabled:      true,
		MultiWordTagEnabled: false,
		ColontagEnabled:     true,
	})
}

func (c *Container) NoteFinder(tx sqlite.Transaction, opts fzf.NoteFinderOpts) *fzf.NoteFinder {
	notes := sqlite.NewNoteDAO(tx, c.Logger)
	return fzf.NewNoteFinder(opts, notes, c.Terminal)
}

func (c *Container) NoteIndexer(tx sqlite.Transaction) *sqlite.NoteIndexer {
	notes := sqlite.NewNoteDAO(tx, c.Logger)
	collections := sqlite.NewCollectionDAO(tx, c.Logger)
	return sqlite.NewNoteIndexer(notes, collections, c.Logger)
}

// Database returns the DB instance for the given notebook, after executing any
// pending migration and indexing the notes if needed.
func (c *Container) Database(zk *zk.Zk, forceIndexing bool) (*sqlite.DB, note.IndexingStats, error) {
	var stats note.IndexingStats

	db, err := sqlite.Open(zk.DBPath())
	if err != nil {
		return nil, stats, err
	}
	needsReindexing, err := db.Migrate()
	if err != nil {
		return nil, stats, errors.Wrap(err, "failed to migrate the database")
	}

	stats, err = c.index(zk, db, forceIndexing || needsReindexing)
	if err != nil {
		return nil, stats, err
	}

	return db, stats, err
}

func (c *Container) index(zk *zk.Zk, db *sqlite.DB, force bool) (note.IndexingStats, error) {
	var bar = progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSpinnerType(14),
	)

	var err error
	var stats note.IndexingStats
	err = db.WithTransaction(func(tx sqlite.Transaction) error {
		stats, err = note.Index(
			zk,
			force,
			c.Parser(),
			c.NoteIndexer(tx),
			c.Logger,
			func(change paths.DiffChange) {
				bar.Add(1)
				bar.Describe(change.String())
			},
		)
		return err
	})
	bar.Clear()

	return stats, err
}

// Paginate creates an auto-closing io.Writer which will be automatically
// paginated if noPager is false, using the user's pager.
//
// You can write to the pager only in the run callback.
func (c *Container) Paginate(noPager bool, config zk.Config, run func(out io.Writer) error) error {
	pager, err := c.pager(noPager || config.Tool.Pager.IsEmpty(), config)
	if err != nil {
		return err
	}
	err = run(pager)
	pager.Close()
	return err
}

func (c *Container) pager(noPager bool, config zk.Config) (*pager.Pager, error) {
	if noPager || !c.Terminal.IsInteractive() {
		return pager.PassthroughPager, nil
	} else {
		return pager.New(config.Tool.Pager, c.Logger)
	}
}
