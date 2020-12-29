package note

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/core"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/mickael-menu/zk/util/paths"
	"github.com/mickael-menu/zk/util/rand"
)

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

// Create generates a new note from the given options.
func Create(zk *zk.Zk, opts CreateOpts, templateLoader core.TemplateLoader) (string, error) {
	wrap := errors.Wrapper("note creation failed")

	exists, err := paths.Exists(opts.Dir.Path)
	if err != nil {
		return "", wrap(err)
	}
	if !exists {
		return "", wrap(fmt.Errorf("directory not found at %v", opts.Dir.Path))
	}

	context, err := newRenderContext(zk, opts, templateLoader)
	if err != nil {
		return "", wrap(err)
	}

	templatePath := opts.Template.OrDefault(
		zk.Template(opts.Dir).OrDefault(""),
	)
	if templatePath != "" {
		template, err := templateLoader.LoadFile(templatePath)
		if err != nil {
			return "", wrap(err)
		}
		content, err := template.Render(context)
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
	ID           string
	Extra        map[string]string
}

func newRenderContext(zk *zk.Zk, opts CreateOpts, templateLoader core.TemplateLoader) (renderContext, error) {
	if opts.Extra == nil {
		opts.Extra = make(map[string]string)
	}
	for k, v := range zk.Extra(opts.Dir) {
		if _, ok := opts.Extra[k]; !ok {
			opts.Extra[k] = v
		}
	}

	template, err := templateLoader.Load(zk.FilenameTemplate(opts.Dir))
	if err != nil {
		return renderContext{}, err
	}

	genContext := newRenderContextGenerator(template, opts)
	genId := rand.NewIDGenerator(zk.IDOptions(opts.Dir))
	// Attempts to generate a new render context until the generated filepath doesn't exist.
	for {
		context, err := genContext(genId())
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

type renderContextGenerator func(id string) (renderContext, error)

func newRenderContextGenerator(
	filenameTemplate core.Template,
	opts CreateOpts,
) renderContextGenerator {
	context := renderContext{
		// FIXME Customize default title in config
		Title:   opts.Title.OrDefault("Untitled"),
		Content: opts.Content.Unwrap(),
		Extra:   opts.Extra,
	}

	i := 0
	return func(id string) (renderContext, error) {
		i++
		// Attempts 50ish tries before failing.
		if i >= 50 {
			return context, fmt.Errorf("%v: file already exists", context.Path)
		}

		context.ID = id

		filename, err := filenameTemplate.Render(context)
		if err != nil {
			return context, err
		}

		// FIXME Customize extension in config
		path := filepath.Join(opts.Dir.Path, filename+".md")
		// Same path as before? We can fail because there's no random component
		// in the filename template.
		if context.Path == path {
			return context, fmt.Errorf("%v: file already exists", path)
		}

		context.Path = path
		context.Filename = filepath.Base(path)
		context.FilenameStem = paths.FilenameStem(path)
		return context, nil
	}
}
