package handlebars

import (
	"html"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/adapter/handlebars/helpers"
	"github.com/mickael-menu/zk/core/templ"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/errors"
)

func Init(lang string, logger util.Logger) {
	helpers.RegisterDate(logger)
	helpers.RegisterPrepend(logger)
	helpers.RegisterShell(logger)
	helpers.RegisterSlug(logger, lang)
}

// Template renders a parsed handlebars template.
type Template struct {
	template *raymond.Template
}

// Render renders the template with the given context.
func (t *Template) Render(context interface{}) (string, error) {
	res, err := t.template.Exec(context)
	if err != nil {
		return "", errors.Wrap(err, "render template failed")
	}
	return html.UnescapeString(res), nil
}

// Loader loads and holds parsed handlebars templates.
type Loader struct {
	strings map[string]*Template
	files   map[string]*Template
}

// NewLoader creates a new instance of Loader.
func NewLoader() *Loader {
	return &Loader{
		strings: make(map[string]*Template),
		files:   make(map[string]*Template),
	}
}

// Load retrieves or parses a handlebars string template.
func (l *Loader) Load(content string) (templ.Renderer, error) {
	wrap := errors.Wrapperf("load template failed")

	// Already loaded?
	template, ok := l.strings[content]
	if ok {
		return template, nil
	}

	// Load new template.
	vendorTempl, err := raymond.Parse(content)
	if err != nil {
		return nil, wrap(err)
	}
	template = &Template{vendorTempl}
	l.strings[content] = template
	return template, nil
}

// LoadFile retrieves or parses a handlebars file template.
func (l *Loader) LoadFile(path string) (templ.Renderer, error) {
	wrap := errors.Wrapper("load template file failed")

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, wrap(err)
	}

	// Already loaded?
	template, ok := l.files[path]
	if ok {
		return template, nil
	}

	// Load new template.
	vendorTempl, err := raymond.ParseFile(path)
	if err != nil {
		return nil, wrap(err)
	}
	template = &Template{vendorTempl}
	l.files[path] = template
	return template, nil
}
