package cmd

import (
	"io"
	"os"
	"path/filepath"
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
	Config         zk.Config
	Date           date.Provider
	Logger         util.Logger
	Terminal       *term.Terminal
	templateLoader *handlebars.Loader
	zk             *zk.Zk
	zkErr          error
}

func NewContainer() (*Container, error) {
	wrap := errors.Wrapper("initialization")

	config := zk.NewDefaultConfig()

	// Load global user config
	configPath, err := locateGlobalConfig()
	if err != nil {
		return nil, wrap(err)
	}
	if configPath != "" {
		config, err = zk.OpenConfig(configPath, config)
		if err != nil {
			return nil, wrap(err)
		}
	}

	// Open current notebook
	zk, zkErr := zk.Open(".", config)
	if zkErr == nil {
		config = zk.Config
		os.Setenv("ZK_PATH", zk.Path)
	}

	date := date.NewFrozenNow()

	return &Container{
		Config: config,
		Logger: util.NewStdLogger("zk: ", 0),
		// zk is short-lived, so we freeze the current date to use the same
		// date for any rendering during the execution.
		Date:     &date,
		Terminal: term.New(),
		zk:       zk,
		zkErr:    zkErr,
	}, nil
}

// locateGlobalConfig looks for the global zk config file following the
// XDG Base Directory specification
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func locateGlobalConfig() (string, error) {
	configHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		home, ok := os.LookupEnv("HOME")
		if !ok {
			home = "~/"
		}
		configHome = filepath.Join(home, ".config")
	}

	configPath := filepath.Join(configHome, "zk/config.toml")
	exists, err := paths.Exists(configPath)
	switch {
	case err != nil:
		return "", err
	case exists:
		return configPath, nil
	default:
		return "", nil
	}
}

func (c *Container) Zk() (*zk.Zk, error) {
	return c.zk, c.zkErr
}

func (c *Container) TemplateLoader(lang string) *handlebars.Loader {
	if c.templateLoader == nil {
		handlebars.Init(lang, c.Terminal.SupportsUTF8(), c.Logger, c.Terminal)
		c.templateLoader = handlebars.NewLoader()
	}
	return c.templateLoader
}

func (c *Container) Parser(zk *zk.Zk) *markdown.Parser {
	return markdown.NewParser(markdown.ParserOpts{
		HashtagEnabled:      zk.Config.Format.Markdown.Hashtags,
		MultiWordTagEnabled: zk.Config.Format.Markdown.MultiwordTags,
		ColontagEnabled:     zk.Config.Format.Markdown.ColonTags,
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
			c.Parser(zk),
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
