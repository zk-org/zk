package core

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/paths"
)

// Metadata used to generate a link.
type LinkFormatterContext struct {
	// Filename of the note
	Filename string `json:"filename"`
	// File path to the note, relative to the notebook root.link-encode-path
	Path string `json:"path"`
	// Absolute file path to the note.
	AbsPath string `json:"absPath" handlebars:"abs-path"`
	// File path to the note, relative to the current directory.
	RelPath string `json:"relPath" handlebars:"rel-path"`
	// Title of the note.
	Title string `json:"title"`
	// Metadata extracted from the YAML frontmatter.
	Metadata map[string]interface{} `json:"metadata"`
}

func NewLinkFormatterContext(path NotebookPath, title string, metadata map[string]interface{}) (LinkFormatterContext, error) {
	relPath, err := path.PathRelToWorkingDir()
	if err != nil {
		return LinkFormatterContext{}, err
	}
	return LinkFormatterContext{
		Filename: path.Filename(),
		Path:     path.Path,
		AbsPath:  path.AbsPath(),
		RelPath:  relPath,
		Title:    title,
		Metadata: metadata,
	}, nil
}

// LinkFormatter formats internal links according to user configuration.
type LinkFormatter func(context LinkFormatterContext) (string, error)

// NewLinkFormatter generates a new LinkFormatter from the user Markdown
// configuration.
func NewLinkFormatter(config MarkdownConfig, templateLoader TemplateLoader) (LinkFormatter, error) {
	switch config.LinkFormat {
	case "markdown", "":
		return NewMarkdownLinkFormatter(config, false)
	case "wiki":
		return NewWikiLinkFormatter(config)
	default:
		return NewCustomLinkFormatter(config, templateLoader)
	}
}

func NewMarkdownLinkFormatter(config MarkdownConfig, onlyHref bool) (LinkFormatter, error) {
	return func(context LinkFormatterContext) (string, error) {
		path := formatPath(context.RelPath, config)
		if !config.LinkEncodePath {
			path = strings.ReplaceAll(path, `\`, `\\`)
			path = strings.ReplaceAll(path, `)`, `\)`)
		}
		if onlyHref {
			return fmt.Sprintf("(%s)", path), nil
		} else {
			title := context.Title
			title = strings.ReplaceAll(title, `\`, `\\`)
			title = strings.ReplaceAll(title, `]`, `\]`)
			return fmt.Sprintf("[%s](%s)", title, path), nil
		}
	}, nil
}

func NewWikiLinkFormatter(config MarkdownConfig) (LinkFormatter, error) {
	return func(context LinkFormatterContext) (string, error) {
		path := formatPath(context.Path, config)
		if !config.LinkEncodePath {
			path = strings.ReplaceAll(path, `\`, `\\`)
			path = strings.ReplaceAll(path, `]]`, `\]]`)
		}
		return "[[" + path + "]]", nil
	}, nil
}

func NewCustomLinkFormatter(config MarkdownConfig, templateLoader TemplateLoader) (LinkFormatter, error) {
	wrap := errors.Wrapperf("failed to render custom link with format: %s", config.LinkFormat)
	template, err := templateLoader.LoadTemplate(config.LinkFormat)
	if err != nil {
		return nil, wrap(err)
	}

	return func(context LinkFormatterContext) (string, error) {
		context.Filename = formatPath(context.Filename, config)
		context.Path = formatPath(context.Path, config)
		context.RelPath = formatPath(context.RelPath, config)
		context.AbsPath = formatPath(context.AbsPath, config)
		return template.Render(context)
	}, nil
}

func formatPath(path string, config MarkdownConfig) string {
	if config.LinkDropExtension {
		path = paths.DropExt(path)
	}
	if config.LinkEncodePath {
		path = strings.ReplaceAll(url.PathEscape(path), "%2F", "/")
	}
	return path
}
