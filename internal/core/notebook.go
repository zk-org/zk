package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
)

// Notebook handles queries and commands performed on an opened notebook.
type Notebook struct {
	Path                  string
	config                Config
	index                 NoteIndex
	fs                    FileStorage
	templateLoaderFactory TemplateLoaderFactory
	idGeneratorFactory    IDGeneratorFactory
	// Returns the OS environment variables.
	osEnv func() map[string]string
}

// Index indexes the content of the notebook to be searchable.
// If force is true, existing notes will be reindexed.
func (n *Notebook) Index(force bool) (NoteIndexingStats, error) {
	return NoteIndexingStats{}, nil
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

	dir, err := n.requireDirAt(opts.Directory.OrString(n.Path).Unwrap())
	if err != nil {
		return "", wrap(err)
	}

	config, err := n.config.GroupConfigNamed(opts.Group.OrString(dir.Group).Unwrap())
	if err != nil {
		return "", wrap(err)
	}

	extra := config.Extra
	for k, v := range opts.Extra {
		extra[k] = v
	}

	templates, err := n.newTemplateLoader(config.Note.Lang)
	if err != nil {
		return "", wrap(err)
	}

	cmd := newNoteCmd{
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
	path, err := cmd.execute()
	return path, wrap(err)
}

func (n *Notebook) newTemplateLoader(lang string) (TemplateLoader, error) {
	lookupPaths := []string{filepath.Join(n.Path, ".zk/templates")}
	return n.templateLoaderFactory(lang, lookupPaths)
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

// dir represents a directory inside a notebook.
type dir struct {
	// Name of the directory, which is the path relative to the notebook's root.
	Name string
	// Absolute path to the directory.
	Path string
	// Name of the config group this directory belongs to, if any.
	Group string
}

// rootDir returns the root directory for this notebook.
func (n *Notebook) rootDir() dir {
	return dir{
		Name:  "",
		Path:  n.Path,
		Group: "",
	}
}

// dirAt returns a dir representation of the notebook directory at the given path.
//
// If config overrides are provided, the dir.Config will be modified using them.
func (n *Notebook) dirAt(path string) (dir, error) {
	name, err := n.RelPath(path)
	if err != nil {
		return dir{}, err
	}

	group, err := n.config.GroupNameForPath(name)
	if err != nil {
		return dir{}, err
	}

	return dir{
		Name:  name,
		Path:  path,
		Group: group,
	}, nil
}

// requireDirAt is the same as dirAt, but checks that the directory exists
// before returning the dir.
func (n *Notebook) requireDirAt(path string) (dir, error) {
	dir, err := n.dirAt(path)
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
