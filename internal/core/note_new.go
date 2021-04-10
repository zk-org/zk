package core

import (
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/mickael-menu/zk/internal/util/paths"
)

type newNoteCmd struct {
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

func (c *newNoteCmd) execute() (string, error) {
	filenameTemplate, err := c.templates.LoadTemplate(c.filenameTemplate)
	if err != nil {
		return "", err
	}

	var contentTemplate Template = NullTemplate
	if templatePath := c.bodyTemplatePath.Unwrap(); templatePath != "" {
		contentTemplate, err = c.templates.LoadTemplateAt(templatePath)
		if err != nil {
			return "", err
		}
	}

	context := renderContext{
		Title:   c.title,
		Content: c.content,
		Dir:     c.dir.Name,
		Extra:   c.extra,
		Now:     c.date,
		Env:     c.env,
	}

	path, context, err := c.generatePath(context, filenameTemplate)
	if err != nil {
		return "", err
	}

	content, err := contentTemplate.Render(context)
	if err != nil {
		return "", err
	}

	err = c.fs.Write(path, []byte(content))
	if err != nil {
		return "", err
	}

	return path, nil
}

func (c *newNoteCmd) generatePath(context renderContext, filenameTemplate Template) (string, renderContext, error) {
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

// renderContext holds the placeholder values which will be expanded in the templates.
type renderContext struct {
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
