package handlebars

import (
	"html"
	"log"
	"os"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/adapter/handlebars/helpers"
	"github.com/mickael-menu/zk/util/errors"
)

func Init() {
	// FIXME Pass the logger as dependency
	logger := log.New(os.Stderr, "zk: warning: ", 0)
	// FIXME Support custom languages from the config
	helpers.RegisterSlug(logger, "en")
}

// HandlebarsRenderer holds parsed handlebars template and renders them.
type HandlebarsRenderer struct {
	templates map[string]*raymond.Template
}

// NewRenderer creates a new instance of HandlebarsRenderer.
func NewRenderer() *HandlebarsRenderer {
	return &HandlebarsRenderer{
		templates: make(map[string]*raymond.Template),
	}
}

// Render renders a handlebars string template with the given context.
func (hr *HandlebarsRenderer) Render(template string, context interface{}) (string, error) {
	res, err := raymond.Render(template, context)
	if err != nil {
		return "", errors.Wrap(err, "render template failed")
	}

	return html.UnescapeString(res), nil
}

// RenderFile renders a handlebars template file with the given context.
func (hr *HandlebarsRenderer) RenderFile(path string, context interface{}) (string, error) {
	wrap := errors.Wrapper("render template failed")

	templ, err := hr.loadFileTemplate(path)
	if err != nil {
		return "", wrap(err)
	}

	res, err := templ.Exec(context)
	if err != nil {
		return "", wrap(err)
	}

	return html.UnescapeString(res), nil
}

// LoadFileTemplate loads the template at given path into the renderer if needed.
// Returns the parsed template.
func (hr *HandlebarsRenderer) loadFileTemplate(path string) (*raymond.Template, error) {
	wrap := errors.Wrapperf("load template file failed: %v", path)

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}

	// Already loaded?
	templ, ok := hr.templates[path]
	if ok {
		return templ, nil
	}

	// Load new template.
	templ, err = raymond.ParseFile(path)
	if err != nil {
		return nil, wrap(err)
	}
	hr.templates[path] = templ
	return templ, nil
}
