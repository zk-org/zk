package handlebars

import (
	"fmt"
	"html"
	"path/filepath"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/internal/adapter/handlebars/helpers"
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
)

func Init(supportsUTF8 bool, logger util.Logger) {
	helpers.RegisterConcat()
	helpers.RegisterSubstring()
	helpers.RegisterDate(logger)
	helpers.RegisterJoin()
	helpers.RegisterJSON(logger)
	helpers.RegisterList(supportsUTF8)
	helpers.RegisterPrepend(logger)
	helpers.RegisterShell(logger)
}

// Template renders a parsed handlebars template.
type Template struct {
	template *raymond.Template
	styler   core.Styler
}

// Styler implements core.Template.
func (t *Template) Styler() core.Styler {
	return t.styler
}

// Render implements core.Template.
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
	styler      core.Styler
	helpers     map[string]interface{}
}

type LoaderOpts struct {
	// LookupPaths is used to resolve relative template paths.
	LookupPaths []string
	Styler      core.Styler
}

// NewLoader creates a new instance of Loader.
//
func NewLoader(opts LoaderOpts) *Loader {
	return &Loader{
		strings:     make(map[string]*Template),
		files:       make(map[string]*Template),
		lookupPaths: opts.LookupPaths,
		styler:      opts.Styler,
		helpers:     map[string]interface{}{},
	}
}

// RegisterHelper declares a new template helper to be used with this loader only.
func (l *Loader) RegisterHelper(name string, helper interface{}) {
	l.helpers[name] = helper
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
	template = l.newTemplate(vendorTempl)
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
	template = l.newTemplate(vendorTempl)
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

func (l *Loader) newTemplate(vendorTempl *raymond.Template) *Template {
	vendorTempl.RegisterHelpers(l.helpers)
	return &Template{vendorTempl, l.styler}
}
