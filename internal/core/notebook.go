package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/paths"
	"github.com/schollz/progressbar/v3"
)

// Notebook handles queries and commands performed on an opened notebook.
type Notebook struct {
	Path   string
	Config Config

	index                 NoteIndex
	parser                NoteParser
	templateLoaderFactory TemplateLoaderFactory
	idGeneratorFactory    IDGeneratorFactory
	fs                    FileStorage
	logger                util.Logger
	osEnv                 func() map[string]string
}

// NewNotebook creates a new Notebook instance.
func NewNotebook(
	path string,
	config Config,
	ports NotebookPorts,
) *Notebook {
	return &Notebook{
		Path:                  path,
		Config:                config,
		index:                 ports.NoteIndex,
		parser:                ports.NoteParser,
		templateLoaderFactory: ports.TemplateLoaderFactory,
		idGeneratorFactory:    ports.IDGeneratorFactory,
		fs:                    ports.FS,
		logger:                ports.Logger,
		osEnv:                 ports.OSEnv,
	}
}

type NotebookPorts struct {
	NoteIndex             NoteIndex
	NoteParser            NoteParser
	TemplateLoaderFactory TemplateLoaderFactory
	IDGeneratorFactory    IDGeneratorFactory
	FS                    FileStorage
	Logger                util.Logger
	OSEnv                 func() map[string]string
}

// NotebookFactory creates a new Notebook instance at the given root path.
type NotebookFactory func(path string, config Config) (*Notebook, error)

// Index indexes the content of the notebook to be searchable.
// If force is true, existing notes will be reindexed.
func (n *Notebook) Index(force bool) (stats NoteIndexingStats, err error) {
	// FIXME: Move out of Core
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSpinnerType(14),
	)

	err = n.index.Commit(func(index NoteIndex) error {
		task := indexTask{
			notebook: n,
			force:    force,
			index:    index,
			parser:   n.parser,
			logger:   n.logger,
		}
		stats, err = task.execute(func(change paths.DiffChange) {
			bar.Add(1)
			bar.Describe(change.String())
		})
		return err
	})

	bar.Clear()
	err = errors.Wrap(err, "indexing")
	return
}

// NewNoteOpts holds the options used to create a new note in a Notebook.
type NewNoteOpts struct {
	// Title of the new note.
	Title opt.String
	// Initial content of the note.
	Content string
	// Directory in which to create the note, relative to the root of the notebook.
	Directory opt.String
	// Group this note belongs to.
	Group opt.String
	// Path to a custom template used to render the note.
	Template opt.String
	// Extra variables passed to the templates.
	Extra map[string]string
	// Creation date provided to the templates.
	Date time.Time
}

// ErrNoteExists is an error returned when a note already exists with the
// generated filename.
type ErrNoteExists struct {
	Name string
	Path string
}

func (e ErrNoteExists) Error() string {
	return fmt.Sprintf("%s: note already exists", e.Path)
}

// NewNote generates a new note in the notebook and returns its path.
//
// Returns ErrNoteExists if no free filename can be generated for this note.
func (n *Notebook) NewNote(opts NewNoteOpts) (string, error) {
	wrap := errors.Wrapper("new note")

	dir, err := n.RequireDirAt(opts.Directory.OrString(n.Path).Unwrap())
	if err != nil {
		return "", wrap(err)
	}

	config, err := n.Config.GroupConfigNamed(opts.Group.OrString(dir.Group).Unwrap())
	if err != nil {
		return "", wrap(err)
	}

	extra := config.Extra
	for k, v := range opts.Extra {
		extra[k] = v
	}

	templates, err := n.templateLoaderFactory(config.Note.Lang)
	if err != nil {
		return "", wrap(err)
	}

	task := newNoteTask{
		dir:              dir,
		title:            opts.Title.OrString(config.Note.DefaultTitle).Unwrap(),
		content:          opts.Content,
		date:             opts.Date,
		extra:            extra,
		env:              n.osEnv(),
		fs:               n.fs,
		filenameTemplate: config.Note.FilenameTemplate + "." + config.Note.Extension,
		bodyTemplatePath: opts.Template.Or(config.Note.BodyTemplatePath),
		templates:        templates,
		genID:            n.idGeneratorFactory(config.Note.IDOptions),
	}
	path, err := task.execute()
	return path, wrap(err)
}

// FindNotes retrieves the notes matching the given filtering options.
func (n *Notebook) FindNotes(opts NoteFindOpts) ([]ContextualNote, error) {
	return n.index.Find(opts)
}

// FindMinimalNotes retrieves lightweight metadata for the notes matching
// the given filtering options.
func (n *Notebook) FindMinimalNotes(opts NoteFindOpts) ([]MinimalNote, error) {
	return n.index.FindMinimal(opts)
}

// FindByHref retrieves the first note matching the given link href.
func (n *Notebook) FindByHref(href string) (*MinimalNote, error) {
	notes, err := n.FindMinimalNotes(NoteFindOpts{
		IncludePaths: []string{href},
		Limit:        1,
		// To find the best match possible, we sort by path length.
		// See https://github.com/mickael-menu/zk/issues/23
		Sorters: []NoteSorter{{Field: NoteSortPathLength, Ascending: true}},
	})

	switch {
	case err != nil:
		return nil, err
	case len(notes) == 0:
		return nil, nil
	default:
		return &notes[0], nil
	}
}

// FindCollections retrieves all the collections of the given kind.
func (n *Notebook) FindCollections(kind CollectionKind) ([]Collection, error) {
	return n.index.FindCollections(kind)
}

// RelPath returns the path relative to the notebook root to the given path.
func (n *Notebook) RelPath(originalPath string) (string, error) {
	wrap := errors.Wrapperf("%v: not a valid notebook path", originalPath)

	path, err := n.fs.Abs(originalPath)
	if err != nil {
		return path, wrap(err)
	}

	path, err = filepath.Rel(n.Path, path)
	if err != nil {
		return path, wrap(err)
	}
	if strings.HasPrefix(path, "..") {
		return path, fmt.Errorf("%s: path is outside the notebook", originalPath)
	}
	if path == "." {
		path = ""
	}
	return path, nil
}

// Dir represents a directory inside a notebook.
type Dir struct {
	// Name of the directory, which is the path relative to the notebook's root.
	Name string
	// Absolute path to the directory.
	Path string
	// Name of the config group this directory belongs to, if any.
	Group string
}

// RootDir returns the root directory for this notebook.
func (n *Notebook) RootDir() Dir {
	return Dir{
		Name:  "",
		Path:  n.Path,
		Group: "",
	}
}

// DirAt returns a Dir representation of the notebook directory at the given path.
func (n *Notebook) DirAt(path string) (Dir, error) {
	path, err := n.fs.Abs(path)
	if err != nil {
		return Dir{}, err
	}

	name, err := n.RelPath(path)
	if err != nil {
		return Dir{}, err
	}

	group, err := n.Config.GroupNameForPath(name)
	if err != nil {
		return Dir{}, err
	}

	return Dir{
		Name:  name,
		Path:  path,
		Group: group,
	}, nil
}

// RequireDirAt is the same as DirAt, but checks that the directory exists
// before returning the Dir.
func (n *Notebook) RequireDirAt(path string) (Dir, error) {
	dir, err := n.DirAt(path)
	if err != nil {
		return dir, err
	}
	exists, err := n.fs.DirExists(dir.Path)
	if err != nil {
		return dir, err
	}
	if !exists {
		return dir, fmt.Errorf("%v: directory not found", path)
	}
	return dir, nil
}

// NewNoteFormatter returns a NoteFormatter used to format notes with the given template.
func (n *Notebook) NewNoteFormatter(templateString string) (NoteFormatter, error) {
	templates, err := n.templateLoaderFactory(n.Config.Note.Lang)
	if err != nil {
		return nil, err
	}
	template, err := templates.LoadTemplate(templateString)
	if err != nil {
		return nil, err
	}

	return newNoteFormatter(n.Path, template, n.fs)
}
