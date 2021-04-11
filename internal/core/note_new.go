package core

import (
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/paths"
)

type newNoteTask struct {
	dir              Dir
	title            string
	content          string
	date             time.Time
	extra            map[string]string
	env              map[string]string
	fs               FileStorage
	filenameTemplate string
	bodyTemplatePath opt.String
	templates        TemplateLoader
	genID            IDGenerator
}

func (t *newNoteTask) execute() (string, error) {
	filenameTemplate, err := t.templates.LoadTemplate(t.filenameTemplate)
	if err != nil {
		return "", err
	}

	var contentTemplate Template = NullTemplate
	if templatePath := t.bodyTemplatePath.Unwrap(); templatePath != "" {
		contentTemplate, err = t.templates.LoadTemplateAt(templatePath)
		if err != nil {
			return "", err
		}
	}

	context := newNoteTemplateContext{
		Title:   t.title,
		Content: t.content,
		Dir:     t.dir.Name,
		Extra:   t.extra,
		Now:     t.date,
		Env:     t.env,
	}

	path, context, err := t.generatePath(context, filenameTemplate)
	if err != nil {
		return "", err
	}

	content, err := contentTemplate.Render(context)
	if err != nil {
		return "", err
	}

	err = t.fs.Write(path, []byte(content))
	if err != nil {
		return "", err
	}

	return path, nil
}

func (c *newNoteTask) generatePath(context newNoteTemplateContext, filenameTemplate Template) (string, newNoteTemplateContext, error) {
	var err error
	var filename string
	var path string

	for i := 0; i < 50; i++ {
		context.ID = c.genID()

		filename, err = filenameTemplate.Render(context)
		if err != nil {
			return "", context, err
		}

		path = filepath.Join(c.dir.Path, filename)
		exists, err := c.fs.FileExists(path)
		if err != nil {
			return "", context, err
		} else if !exists {
			context.Filename = filepath.Base(path)
			context.FilenameStem = paths.FilenameStem(path)
			return path, context, nil
		}
	}

	return "", context, ErrNoteExists{
		Name: filepath.Join(c.dir.Name, filename),
		Path: path,
	}
}

// newNoteTemplateContext holds the placeholder values which will be expanded in the templates.
type newNoteTemplateContext struct {
	ID           string `handlebars:"id"`
	Title        string
	Content      string
	Dir          string
	Filename     string
	FilenameStem string `handlebars:"filename-stem"`
	Extra        map[string]string
	Now          time.Time
	Env          map[string]string
}
