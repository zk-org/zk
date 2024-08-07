package cli

import (
	"io"
	"os"
	"path/filepath"

	"github.com/zk-org/zk/internal/adapter/editor"
	"github.com/zk-org/zk/internal/adapter/fs"
	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/adapter/handlebars"
	hbhelpers "github.com/zk-org/zk/internal/adapter/handlebars/helpers"
	"github.com/zk-org/zk/internal/adapter/markdown"
	"github.com/zk-org/zk/internal/adapter/sqlite"
	"github.com/zk-org/zk/internal/adapter/term"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	osutil "github.com/zk-org/zk/internal/util/os"
	"github.com/zk-org/zk/internal/util/pager"
	"github.com/zk-org/zk/internal/util/paths"
	"github.com/zk-org/zk/internal/util/rand"
)

type Dirs struct {
	NotebookDir string
	WorkingDir  string
}

type Container struct {
	Version            string
	Config             core.Config
	Logger             *util.ProxyLogger
	Styler             *core.ProxyStyler
	Terminal           *term.Terminal
	FS                 *fs.FileStorage
	TemplateLoader     core.TemplateLoader
	WorkingDir         string
	Notebooks          *core.NotebookStore
	currentNotebook    *core.Notebook
	currentNotebookErr error
}

func NewContainer(version string) (*Container, error) {
	wrap := errors.Wrapper("initialization")

	term := term.New()
	styler := core.NewProxyStyler(term)
	logger := util.NewProxyLogger(util.NewStdLogger("zk: ", 0))
	fs, err := fs.NewFileStorage("", logger)
	config := core.NewDefaultConfig()

	handlebars.Init(term.SupportsUTF8(), logger)
	// Template loader used for embedded templates (e.g. default config, fzf
	// line, etc.).
	templateLoader := handlebars.NewLoader(handlebars.LoaderOpts{
		LookupPaths: []string{},
		Styler:      styler,
	})
	templateLoader.RegisterHelper("style", hbhelpers.NewStyleHelper(styler, logger))

	// Load global user config
	configPath, err := locateGlobalConfig()
	if err != nil {
		return nil, wrap(err)
	}
	if configPath != "" {
		config, err = core.OpenConfig(configPath, config, fs, true)
		if err != nil {
			return nil, wrap(err)
		}
	}

	// Set the default notebook if not already set
	// might be overrided if --notebook-dir flag is present
	os.Setenv("ZK_NOTEBOOK_DIR", config.Notebook.Dir.Unwrap())

	// Set the default shell if not already set
	if osutil.GetOptEnv("ZK_SHELL").IsNull() && !config.Tool.Shell.IsEmpty() {
		os.Setenv("ZK_SHELL", config.Tool.Shell.Unwrap())
	}

	return &Container{
		Version:        version,
		Config:         config,
		Logger:         logger,
		Styler:         styler,
		Terminal:       term,
		FS:             fs,
		TemplateLoader: templateLoader,
		Notebooks: core.NewNotebookStore(config, core.NotebookStorePorts{
			FS:             fs,
			TemplateLoader: templateLoader,
			NotebookFactory: func(path string, config core.Config) (*core.Notebook, error) {
				dbPath := filepath.Join(path, ".zk/notebook.db")
				db, err := sqlite.Open(dbPath)
				if err != nil {
					return nil, err
				}

				notebook := core.NewNotebook(path, config, core.NotebookPorts{
					NoteIndex: sqlite.NewNoteIndex(path, db, logger),
					NoteContentParser: markdown.NewParser(
						markdown.ParserOpts{
							HashtagEnabled:      config.Format.Markdown.Hashtags,
							MultiWordTagEnabled: config.Format.Markdown.MultiwordTags,
							ColontagEnabled:     config.Format.Markdown.ColonTags,
						},
						logger,
					),
					TemplateLoaderFactory: func(language string) (core.TemplateLoader, error) {
						loader := handlebars.NewLoader(handlebars.LoaderOpts{
							LookupPaths: []string{
								filepath.Join(globalConfigDir(), "templates"),
								filepath.Join(path, ".zk/templates"),
							},
							Styler: styler,
						})

						loader.RegisterHelper("style", hbhelpers.NewStyleHelper(styler, logger))
						loader.RegisterHelper("slug", hbhelpers.NewSlugHelper(language, logger))

						linkFormatter, err := core.NewLinkFormatter(config.Format.Markdown, loader)
						if err != nil {
							return nil, err
						}
						loader.RegisterHelper("format-link", hbhelpers.NewLinkHelper(linkFormatter, logger))

						return loader, nil
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
	if _, ok := os.LookupEnv("RUNNING_TESH"); ok {
		return "", nil
	}

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
func (c *Container) SetCurrentNotebook(searchDirs []Dirs) error {
	if len(searchDirs) == 0 {
		return nil
	}

	for _, dirs := range searchDirs {
		notebookDir := c.FS.Canonical(dirs.NotebookDir)
		workingDir := c.FS.Canonical(dirs.WorkingDir)

		c.currentNotebook, c.currentNotebookErr = c.Notebooks.Open(notebookDir)
		if c.currentNotebookErr == nil {
			c.setWorkingDir(workingDir)
			c.Config = c.currentNotebook.Config
			// FIXME: Is there something to do to support multiple notebooks here?
			os.Setenv("ZK_NOTEBOOK_DIR", c.currentNotebook.Path)
		}
		// Report the error only if it's not the "notebook not found" one.
		var errNotFound core.ErrNotebookNotFound
		if !errors.As(c.currentNotebookErr, &errNotFound) {
			return c.currentNotebookErr
		}
	}
	return nil
}

// SetWorkingDir resets the current working directory.
func (c *Container) setWorkingDir(path string) {
	path = c.FS.Canonical(path)
	c.WorkingDir = path
	c.FS.SetWorkingDir(path)
}

// CurrentNotebook returns the current default notebook.
func (c *Container) CurrentNotebook() (*core.Notebook, error) {
	return c.currentNotebook, c.currentNotebookErr
}

func (c *Container) NewNoteFilter(opts fzf.NoteFilterOpts) *fzf.NoteFilter {
	opts.PreviewCmd = c.Config.Tool.FzfPreview
	opts.LineTemplate = c.Config.Tool.FzfLine
	opts.FzfOptions = c.Config.Tool.FzfOptions
	opts.NewBinding = c.Config.Tool.FzfBindNew
	return fzf.NewNoteFilter(opts, c.FS, c.Terminal, c.TemplateLoader)
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
