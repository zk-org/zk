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
	var formatter LinkFormatter
	var err error
	switch config.LinkFormat {
	case "markdown", "":
		formatter, err = newMarkdownLinkFormatter(config)
	case "wiki":
		formatter, err = newWikiLinkFormatter(config)
	default:
		formatter, err = newCustomLinkFormatter(config, templateLoader)
	}

	if err != nil {
		return nil, err
	}

	return func(path, title string) (string, error) {
		if config.LinkDropExtension {
			path = paths.DropExt(path)
		}
		if config.LinkEncodePath {
			path = strings.ReplaceAll(url.PathEscape(path), "%2F", "/")
		}

		return formatter(path, title)
	}, nil
}

func newMarkdownLinkFormatter(config MarkdownConfig) (LinkFormatter, error) {
	return func(path, title string) (string, error) {
		if !config.LinkEncodePath {
			path = strings.ReplaceAll(path, `\`, `\\`)
			path = strings.ReplaceAll(path, `)`, `\)`)
		}
		title = strings.ReplaceAll(title, `\`, `\\`)
		title = strings.ReplaceAll(title, `]`, `\]`)
		return fmt.Sprintf("[%s](%s)", title, path), nil
	}, nil
}

func newWikiLinkFormatter(config MarkdownConfig) (LinkFormatter, error) {
	return func(path, title string) (string, error) {
		if !config.LinkEncodePath {
			path = strings.ReplaceAll(path, `\`, `\\`)
			path = strings.ReplaceAll(path, `]]`, `\]]`)
		}
		return "[[" + path + "]]", nil
	}, nil
}

func newCustomLinkFormatter(config MarkdownConfig, templateLoader TemplateLoader) (LinkFormatter, error) {
	wrap := errors.Wrapperf("failed to render custom link with format: %s", config.LinkFormat)
	template, err := templateLoader.LoadTemplate(config.LinkFormat)
	if err != nil {
		return nil, wrap(err)
	}

	return func(path, title string) (string, error) {
		return template.Render(customLinkRenderContext{Path: path, Title: title})
	}, nil
}

type customLinkRenderContext struct {
	Path  string
	Title string
}
