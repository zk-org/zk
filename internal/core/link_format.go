package core

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/paths"
)

// LinkFormatter formats internal links according to user configuration.
type LinkFormatter func(path string, title string) (string, error)

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
	return func(path, title string) (string, error) {
		path = formatPath(path, config)
		if !config.LinkEncodePath {
			path = strings.ReplaceAll(path, `\`, `\\`)
			path = strings.ReplaceAll(path, `)`, `\)`)
		}
		if onlyHref {
			return fmt.Sprintf("(%s)", path), nil
		} else {
			title = strings.ReplaceAll(title, `\`, `\\`)
			title = strings.ReplaceAll(title, `]`, `\]`)
			return fmt.Sprintf("[%s](%s)", title, path), nil
		}
	}, nil
}

func NewWikiLinkFormatter(config MarkdownConfig) (LinkFormatter, error) {
	return func(path, title string) (string, error) {
		path = formatPath(path, config)
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

	return func(path, title string) (string, error) {
		path = formatPath(path, config)
		return template.Render(customLinkRenderContext{Path: path, Title: title})
	}, nil
}

type customLinkRenderContext struct {
	Path  string
	Title string
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
