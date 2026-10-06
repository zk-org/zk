package lsp

import (
	"path/filepath"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/paths"
)

// completionTemplates holds templates to render the various elements of an LSP
// completion item.
type completionTemplates struct {
	Label      core.Template
	FilterText core.Template
	Detail     core.Template
}

func newCompletionTemplates(loader core.TemplateLoader, templates core.LSPCompletionTemplates) (result completionTemplates, err error) {
	if templates.Label != nil {
		result.Label, err = loader.LoadTemplate(*templates.Label)
	}
	if templates.FilterText != nil {
		result.FilterText, err = loader.LoadTemplate(*templates.FilterText)
	}
	if templates.Detail != nil {
		result.Detail, err = loader.LoadTemplate(*templates.Detail)
	}

	return
}

type completionItemRenderContext struct {
	ID           int64
	Filename     string
	FilenameStem string `handlebars:"filename-stem"`
	Path         string
	AbsPath      string `handlebars:"abs-path"`
	RelPath      string `handlebars:"rel-path"`
	Title        string
	TitleOrPath  string `handlebars:"title-or-path"`
	Metadata     map[string]any
}

func newCompletionItemRenderContext(note core.MinimalNote, notebookDir string, currentDir string) (completionItemRenderContext, error) {
	absPath := filepath.Join(notebookDir, note.Path)
	relPath, err := filepath.Rel(currentDir, absPath)
	if err != nil {
		return completionItemRenderContext{}, err
	}

	context := completionItemRenderContext{
		ID:           int64(note.ID),
		Filename:     filepath.Base(note.Path),
		FilenameStem: paths.FilenameStem(note.Path),
		Path:         note.Path,
		AbsPath:      absPath,
		RelPath:      relPath,
		Title:        note.Title,
		TitleOrPath:  note.Title,
		Metadata:     note.Metadata,
	}
	if context.TitleOrPath == "" {
		context.TitleOrPath = note.Path
	}
	return context, nil
}
