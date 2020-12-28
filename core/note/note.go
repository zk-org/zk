package note

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/mickael-menu/zk/util/rand"
)

// Renderer renders templates.
type Renderer interface {
	// Render renders a handlebars string template with the given context.
	Render(template string, context interface{}) (string, error)
	// RenderFile renders a handlebars template file with the given context.
	RenderFile(path string, context interface{}) (string, error)
}

// CreateOpts holds the options to create a new note.
type CreateOpts struct {
	// Parent directory for the new note.
	Dir zk.Dir
	// Title of the note.
	Title opt.String
	// Initial content of the note, which will be injected in the template.
	Content opt.String
	// Custom template to use for the note, overriding the one declared in the config.
	Template opt.String
	// Extra template variables to expand.
	Extra map[string]string
}

// Create generates a new note in the given slip box from the given options.
func Create(zk *zk.Zk, opts CreateOpts, renderer Renderer) (string, error) {
	wrap := errors.Wrapper("note creation failed")

	exists, err := paths.Exists(opts.Dir.Path)
	if err != nil {
		return "", wrap(err)
	}
	if !exists {
		return "", wrap(fmt.Errorf("directory not found at %v", opts.Dir.Path))
	}

	extra := zk.Extra(opts.Dir)
	for k, v := range opts.Extra {
		extra[k] = v
	}

	context := renderContext{
		// FIXME Customize default title in config
		Title:   opts.Title.OrDefault("Untitled"),
		Content: opts.Content.Unwrap(),
		Extra:   extra,
	}

	file, err := genFilepath(zk, opts.Dir, renderer, &context)
	if err != nil {
		return "", wrap(err)
	}

	template := opts.Template.OrDefault(
		zk.Template(opts.Dir).OrDefault(""),
	)
	if template != "" {
		content, err := renderer.RenderFile(template, context)
		if err != nil {
			return "", wrap(err)
		}
		err = paths.WriteString(path, content)
		if err != nil {
			return "", wrap(err)
		}
	}

	return file, nil
}

// renderContext holds the placeholder values which will be expanded in the templates.
type renderContext struct {
	Title        string
	Content      string
	Filename     string
	FilenameStem string `handlebars:"filename-stem"`
	RandomID     string `handlebars:"random-id"`
	Extra        map[string]string
}

func genFilepath(zk *zk.Zk, dir zk.Dir, renderer Renderer, context *renderContext) (string, error) {
	template := zk.FilenameTemplate(dir)
	isRandom := strings.Contains(template, "random-id")

	i := 0
	for {
		context.RandomID = rand.GenID(zk.RandIDOpts(dir))

		filename, err := renderer.Render(template, context)
		if err != nil {
			return "", err
		}

		// FIXME Customize extension in config
		path := filepath.Join(dir.Path, filename+".md")
		exists, err := paths.Exists(path)
		if err != nil {
			return "", err
		}

		if !exists {
			context.Filename = filepath.Base(path)
			context.FilenameStem = paths.FilenameStem(path)
			return path, nil

		} else if !isRandom || i > 50 { // Attempts 50 tries if the filename template contains a random ID before failing.
			return "", fmt.Errorf("%v: file already exists", path)
		}

		i++
	}
}
