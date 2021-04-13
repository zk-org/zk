package cli

import (
	"io"
	"os"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/adapter/editor"
	"github.com/mickael-menu/zk/internal/adapter/fs"
	"github.com/mickael-menu/zk/internal/adapter/fzf"
	"github.com/mickael-menu/zk/internal/adapter/handlebars"
	"github.com/mickael-menu/zk/internal/adapter/markdown"
	"github.com/mickael-menu/zk/internal/adapter/sqlite"
	"github.com/mickael-menu/zk/internal/adapter/term"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	osutil "github.com/mickael-menu/zk/internal/util/os"
	"github.com/mickael-menu/zk/internal/util/pager"
	"github.com/mickael-menu/zk/internal/util/paths"
	"github.com/mickael-menu/zk/internal/util/rand"
)

type Container struct {
	Version            string
	Config             core.Config
	Logger             *util.ProxyLogger
	Terminal           *term.Terminal
	FS                 core.FileStorage
	WorkingDir         string
	Notebooks          *core.NotebookStore
	currentNotebook    *core.Notebook
	currentNotebookErr error
}

func NewContainer(version string) (*Container, error) {
	wrap := errors.Wrapper("initialization")

	term := term.New()
	styler := term
	logger := util.NewProxyLogger(util.NewStdLogger("zk: ", 0))
	fs, err := fs.NewFileStorage("", logger)
	config := core.NewDefaultConfig()

	handlebars.Init(term.SupportsUTF8(), logger)

	// Load global user config
	configPath, err := locateGlobalConfig()
	if err != nil {
		return nil, wrap(err)
	}
	if configPath != "" {
		config, err = core.OpenConfig(configPath, config, fs)
		if err != nil {
			return nil, wrap(err)
		}
	}

	return &Container{
		Version:  version,
		Config:   config,
		Logger:   logger,
		Terminal: term,
		FS:       fs,
		Notebooks: core.NewNotebookStore(config, core.NotebookStorePorts{
			FS: fs,
			NotebookFactory: func(path string, config core.Config) (*core.Notebook, error) {
				dbPath := filepath.Join(path, ".zk/notebook.db")
				db, err := sqlite.Open(dbPath)
				if err != nil {
					return nil, err
				}

				notebook := core.NewNotebook(path, config, core.NotebookPorts{
					NoteIndex: sqlite.NewNoteIndex(db, logger),
					NoteParser: markdown.NewParser(markdown.ParserOpts{
						HashtagEnabled:      config.Format.Markdown.Hashtags,
						MultiWordTagEnabled: config.Format.Markdown.MultiwordTags,
						ColontagEnabled:     config.Format.Markdown.ColonTags,
					}),
					TemplateLoaderFactory: func(language string) (core.TemplateLoader, error) {
						return handlebars.NewLoader(handlebars.LoaderOpts{
							LookupPaths: []string{
								filepath.Join(globalConfigDir(), "templates"),
								filepath.Join(path, ".zk/templates"),
							},
							Lang:   config.Note.Lang,
							Styler: styler,
							Logger: logger,
						}), nil
					},
					IDGeneratorFactory: func(opts core.IDOptions) func() string {
						return rand.NewIDGenerator(opts)
					},
					FS:     fs,
					Logger: logger,
					OSEnv: func() map[string]string {
						return osutil.Env()
					},
				})

				return notebook, nil
			},
		}),
	}, nil
}

// locateGlobalConfig looks for the global zk config file following the
// XDG Base Directory specification
// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func locateGlobalConfig() (string, error) {
	configPath := filepath.Join(globalConfigDir(), "config.toml")
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

// globalConfigDir returns the parent directory of the global configuration file.
func globalConfigDir() string {
	path, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		home, ok := os.LookupEnv("HOME")
		if !ok {
			home = "~/"
		}
		path = filepath.Join(home, ".config")
	}
	return filepath.Join(path, "zk")
}

// SetCurrentNotebook sets the first notebook found in the given search paths
// as the current default one.
func (c *Container) SetCurrentNotebook(searchPaths []string) {
	if len(searchPaths) == 0 {
		return
	}

	for _, path := range searchPaths {
		c.currentNotebook, c.currentNotebookErr = c.Notebooks.Open(path)
		if c.currentNotebookErr == nil {
			c.WorkingDir = path
			c.Config = c.currentNotebook.Config
			// FIXME: multiple notebooks
			os.Setenv("ZK_NOTEBOOK_DIR", c.currentNotebook.Path)
			return
		}
	}
}

// CurrentNotebook returns the current default notebook.
func (c *Container) CurrentNotebook() (*core.Notebook, error) {
	return c.currentNotebook, c.currentNotebookErr
}

func (c *Container) NewNoteFilter(opts fzf.NoteFilterOpts) *fzf.NoteFilter {
	return fzf.NewNoteFilter(opts, c.Terminal)
}

func (c *Container) NewNoteEditor(notebook *core.Notebook) (*editor.Editor, error) {
	return editor.NewEditor(notebook.Config.Tool.Editor)
}

// Paginate creates an auto-closing io.Writer which will be automatically
// paginated if noPager is false, using the user's pager.
//
// You can write to the pager only in the run callback.
func (c *Container) Paginate(noPager bool, run func(out io.Writer) error) error {
	pager, err := c.pager(noPager || c.Config.Tool.Pager.IsEmpty())
	if err != nil {
		return err
	}
	err = run(pager)
	pager.Close()
	return err
}

func (c *Container) pager(noPager bool) (*pager.Pager, error) {
	if noPager || !c.Terminal.IsInteractive() {
		return pager.PassthroughPager, nil
	} else {
		return pager.New(c.Config.Tool.Pager, c.Logger)
	}
}
