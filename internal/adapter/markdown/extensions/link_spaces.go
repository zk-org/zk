package extensions

import (
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// MarkdownLinkWithSpacesExt is an extension parsing Markdown links with spaces in the destination.
// For example, [label](file name.md).
var MarkdownLinkWithSpacesExt = &markdownLinkWithSpaces{}

type markdownLinkWithSpaces struct{}

func (e *markdownLinkWithSpaces) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&mlwsParser{}, 199),
		),
	)
}

type mlwsParser struct{}

func (p *mlwsParser) Trigger() []byte {
	return []byte{'['}
}

var mlwsRegex = regexp.MustCompile(`^\[([^\]]+)\]\(\s*([^)"]*?)\s*(?:\s+"([^"]+)")?\s*\)`)

func (p *mlwsParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	match := mlwsRegex.FindSubmatch(line)
	if match == nil {
		return nil
	}

	label := match[1]
	destination := match[2]
	var title []byte
	if len(match) > 3 {
		title = match[3]
	}

	block.Advance(len(match[0]))

	link := ast.NewLink()
	link.Destination = destination
	if len(title) > 0 {
		link.Title = title
	}
	
	// Use ast.Text so that LinkPosition can find the text segments
	txt := ast.NewText()
	start := segment.Start + 1
	txt.Segment = text.NewSegment(start, start+len(label))
	link.AppendChild(link, txt)

	return link
}
