package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/opt"
	"github.com/zk-org/zk/internal/util/paths"
)

// Notebook handles queries and commands performed on an opened notebook.
type Notebook struct {
	Path   string
	Config Config
	Parser NoteContentParser

	index                 NoteIndex
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
		Parser:                ports.NoteContentParser,
		index:                 ports.NoteIndex,
		templateLoaderFactory: ports.TemplateLoaderFactory,
		idGeneratorFactory:    ports.IDGeneratorFactory,
		fs:                    ports.FS,
		logger:                ports.Logger,
		osEnv:                 ports.OSEnv,
	}
}

type NotebookPorts struct {
	NoteIndex             NoteIndex
	NoteContentParser     NoteContentParser
	TemplateLoaderFactory TemplateLoaderFactory
	IDGeneratorFactory    IDGeneratorFactory
	FS                    FileStorage
	Logger                util.Logger
	OSEnv                 func() map[string]string
}

// NotebookFactory creates a new Notebook instance at the given root path.
type NotebookFactory func(path string, config Config) (*Notebook, error)

// Index indexes the content of the notebook to be searchable.
func (n *Notebook) Index(opts NoteIndexOpts) (stats NoteIndexingStats, err error) {
	return n.IndexWithCallback(opts, func(change paths.DiffChange) {})
}

// IndexWithCallback indexes the content of the notebook to be searchable.
func (n *Notebook) IndexWithCallback(opts NoteIndexOpts, callback func(change paths.DiffChange)) (stats NoteIndexingStats, err error) {
	err = n.index.Commit(func(index NoteIndex) error {
		task := indexTask{
			path:    n.Path,
			config:  n.Config,
			force:   opts.Force,
			verbose: opts.Verbose,
			index:   index,
			parser:  n,
			logger:  n.logger,
		}
		stats, err = task.execute(callback)
		return err
	})

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
	// Don't save the generated note on the file system.
	DryRun bool
	// Use a provided id over generating one
	ID string
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

// NewNote generates a new note in the notebook, index and returns it.
//
// Returns ErrNoteExists if no free filename can be generated for this note.
func (n *Notebook) NewNote(opts NewNoteOpts) (*Note, error) {
	wrap := errors.Wrapper("new note")

	dir, err := n.RequireDirAt(opts.Directory.OrString(n.Path).Unwrap())
	if err != nil {
		return nil, wrap(err)
	}

	config, err := n.Config.GroupConfigNamed(opts.Group.OrString(dir.Group).Unwrap())
	if err != nil {
		return nil, wrap(err)
	}

	extra := config.Extra
	for k, v := range opts.Extra {
		extra[k] = v
	}

	templates, err := n.templateLoaderFactory(config.Note.Lang)
	if err != nil {
		return nil, wrap(err)
	}

	var idGenerator IDGenerator
	if opts.ID != "" {
		idGenerator = func() string {
			return opts.ID
		}
	} else {
		idGenerator = n.idGeneratorFactory(config.Note.IDOptions)
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
		genID:            idGenerator,
		dryRun:           opts.DryRun,
	}
	path, content, err := task.execute()
	if err != nil {
		return nil, wrap(err)
	}

	note, err := n.ParseNoteWithContent(path, []byte(content))
	if note == nil || err != nil {
		return nil, wrap(err)
	}

	if !opts.DryRun {
		id, err := n.index.Add(*note)
		if err != nil {
			return nil, wrap(err)
		}
		note.ID = id
	}

	return note, nil
}

// FindNotes retrieves the notes matching the given filtering options.
func (n *Notebook) FindNotes(opts NoteFindOpts) ([]ContextualNote, error) {
	return n.index.Find(opts)
}

// FindNote retrieves the first note matching the given filtering options.
func (n *Notebook) FindNote(opts NoteFindOpts) (*Note, error) {
	opts.Limit = 1
	notes, err := n.FindNotes(opts)
	switch {
	case err != nil:
		return nil, err
	case len(notes) == 0:
		return nil, nil
	default:
		return &notes[0].Note, nil
	}
}

// FindMinimalNotes retrieves lightweight metadata for the notes matching
// the given filtering options.
func (n *Notebook) FindMinimalNotes(opts NoteFindOpts) ([]MinimalNote, error) {
	return n.index.FindMinimal(opts)
}

// FindMinimalNote retrieves lightweight metadata for the first note matching
// the given filtering options.
func (n *Notebook) FindMinimalNote(opts NoteFindOpts) (*MinimalNote, error) {
	opts.Limit = 1
	notes, err := n.FindMinimalNotes(opts)
	switch {
	case err != nil:
		return nil, err
	case len(notes) == 0:
		return nil, nil
	default:
		return &notes[0], nil
	}
}

// FindByHref retrieves the first note matching the given link href.
// If allowPartialHref is true, the href can match any unique sub portion of a note path.
func (n *Notebook) FindByHref(href string, allowPartialHref bool) (*MinimalNote, error) {
	return n.FindMinimalNote(NoteFindOpts{
		IncludeHrefs:      []string{href},
		AllowPartialHrefs: allowPartialHref,
	})
}

// FindLinksBetweenNotes retrieves the links between the given notes.
func (n *Notebook) FindLinksBetweenNotes(ids []NoteID) ([]ResolvedLink, error) {
	return n.index.FindLinksBetweenNotes(ids)
}

// FindCollections retrieves all the collections of the given kind.
func (n *Notebook) FindCollections(kind CollectionKind, sorters []CollectionSorter) ([]Collection, error) {
	return n.index.FindCollections(kind, sorters)
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
		return path, fmt.Errorf("%s: path is outside the notebook at %s", originalPath, n.Path)
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

	linkFormatter, err := NewLinkFormatter(n.Config.Format.Markdown, templates)
	if err != nil {
		return nil, err
	}

	return newNoteFormatter(n.Path, template, linkFormatter, n.osEnv(), n.fs)
}

// NewCollectionFormatter returns a CollectionFormatter used to format notes with the given template.
func (n *Notebook) NewCollectionFormatter(templateString string) (CollectionFormatter, error) {
	templates, err := n.templateLoaderFactory(n.Config.Note.Lang)
	if err != nil {
		return nil, err
	}
	template, err := templates.LoadTemplate(templateString)
	if err != nil {
		return nil, err
	}

	return newCollectionFormatter(template)
}

// NewLinkFormatter returns a LinkFormatter used to generate internal links between notes.
func (n *Notebook) NewLinkFormatter() (LinkFormatter, error) {
	templates, err := n.templateLoaderFactory(n.Config.Note.Lang)
	if err != nil {
		return nil, err
	}

	return NewLinkFormatter(n.Config.Format.Markdown, templates)
}
