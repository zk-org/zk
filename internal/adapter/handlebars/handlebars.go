package handlebars

import (
	"fmt"
	"html"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/internal/adapter/handlebars/helpers"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/core/style"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
)

func Init(lang string, supportsUTF8 bool, logger util.Logger, styler style.Styler) {
	helpers.RegisterConcat()
	helpers.RegisterDate(logger)
	helpers.RegisterJoin()
	helpers.RegisterList(supportsUTF8)
	helpers.RegisterPrepend(logger)
	helpers.RegisterShell(logger)
	helpers.RegisterSlug(lang, logger)
	helpers.RegisterStyle(styler, logger)
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
	strings     map[string]*Template
	files       map[string]*Template
	lookupPaths []string
}

// NewLoader creates a new instance of Loader.
//
// lookupPaths is used to resolve relative template paths.
func NewLoader(lookupPaths []string) *Loader {
	return &Loader{
		strings:     make(map[string]*Template),
		files:       make(map[string]*Template),
		lookupPaths: lookupPaths,
	}
}

// LoadTemplate implements core.TemplateLoader.
func (l *Loader) LoadTemplate(content string) (core.Template, error) {
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

// LoadTemplateAt implements core.TemplateLoader.
func (l *Loader) LoadTemplateAt(path string) (core.Template, error) {
	wrap := errors.Wrapper("load template file failed")

	path, ok := l.locateTemplate(path)
	if !ok {
		return nil, wrap(fmt.Errorf("cannot find template at %s", path))
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

// locateTemplate returns the absolute path for the given template path, by
// looking for it in the templates directories registered in this Config.
func (l *Loader) locateTemplate(path string) (string, bool) {
	if path == "" {
		return "", false
	}

	exists := func(path string) bool {
		exists, err := paths.Exists(path)
		return exists && err == nil
	}

	if filepath.IsAbs(path) {
		return path, exists(path)
	}

	for _, dir := range l.lookupPaths {
		if candidate := filepath.Join(dir, path); exists(candidate) {
			return candidate, true
		}
	}

	return path, false
}
