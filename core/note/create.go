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

	context, err := newRenderContext(zk, opts, renderer)
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
		err = paths.WriteString(context.Path, content)
		if err != nil {
			return "", wrap(err)
		}

		fmt.Printf("<<<\n%v\n<<<\n", content)
	}

	return context.Path, nil
}

// renderContext holds the placeholder values which will be expanded in the templates.
type renderContext struct {
	Title        string
	Content      string
	Path         string
	Filename     string
	FilenameStem string `handlebars:"filename-stem"`
	RandomID     string `handlebars:"random-id"`
	Extra        map[string]string
}

func newRenderContext(zk *zk.Zk, opts CreateOpts, renderer Renderer) (renderContext, error) {
	if opts.Extra == nil {
		opts.Extra = make(map[string]string)
	}
	for k, v := range zk.Extra(opts.Dir) {
		if _, ok := opts.Extra[k]; !ok {
			opts.Extra[k] = v
		}
	}

	template := zk.FilenameTemplate(opts.Dir)
	idGenerator := rand.NewIDGenerator(zk.RandIDOpts(opts.Dir))
	contextGenerator := newRenderContextGenerator(template, opts, renderer)
	for {
		context, err := contextGenerator(idGenerator())
		if err != nil {
			return context, err
		}
		exists, err := paths.Exists(context.Path)
		if err != nil {
			return context, err
		}
		if !exists {
			return context, nil
		}
	}
}

type renderContextGenerator func(randomID string) (renderContext, error)

func newRenderContextGenerator(
	filenameTemplate string,
	opts CreateOpts,
	renderer Renderer,
) renderContextGenerator {
	context := renderContext{
		// FIXME Customize default title in config
		Title:   opts.Title.OrDefault("Untitled"),
		Content: opts.Content.Unwrap(),
		Extra:   opts.Extra,
	}

	isRandom := strings.Contains(filenameTemplate, "random-id")

	i := 0
	return func(randomID string) (renderContext, error) {
		// Attempts 50ish tries if the filename template contains a random ID before failing.
		if i > 0 && !isRandom || i >= 50 {
			return context, fmt.Errorf("%v: file already exists", context.Path)
		}
		i++

		context.RandomID = randomID

		filename, err := renderer.Render(filenameTemplate, context)
		if err != nil {
			return context, err
		}

		// FIXME Customize extension in config
		path := filepath.Join(opts.Dir.Path, filename+".md")
		context.Path = path
		context.Filename = filepath.Base(path)
		context.FilenameStem = paths.FilenameStem(path)
		return context, nil
	}
}
