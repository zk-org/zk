package extensions

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// WikiLink is an extension parsing wiki links and neuron's Folgezettel.
//
// For example, [[wiki link]], [[[legacy downlink]]], #[[uplink]], [[downlink]]#.
var WikiLink = &wikiLink{}

type wikiLink struct{}

func (w *wikiLink) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&wlParser{}, 199),
		),
	)
}

type wlParser struct{}

func (p *wlParser) Trigger() []byte {
	return []byte{'[', '#'}
}

func (p *wlParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	openerCharCount := 0
	opened := false
	closed := false
	content := []byte{}
	closerCharCount := 0
	endPos := 0

	// Folgezettel direction: -1 down, 0 unknown, 1 up
	direction := 0

	for i, char := range line {
		endPos = i

		if closed {
			// Supports trailing hash syntax for neuron's Folgezettel, e.g. [[id]]#
			if char == '#' {
				direction = -1
			}
			break
		}

		if !opened {
			switch char {
			// Supports leading hash syntax for neuron's Folgezettel, e.g. #[[id]]
			case '#':
				direction = 1
				continue
			case '[':
				openerCharCount += 1
				continue
			}

			if openerCharCount < 2 || openerCharCount > 3 {
				return nil
			}
		}
		opened = true

		if char == ']' {
			closerCharCount += 1
			if closerCharCount == openerCharCount {
				closed = true
				// neuron's legacy [[[Folgezettel]]].
				if closerCharCount == 3 {
					direction = -1
				}
			}
		} else {
			if closerCharCount > 0 {
				content = append(content, strings.Repeat("]", closerCharCount)...)
				closerCharCount = 0
			}
			content = append(content, char)
		}
	}

	if !closed || len(content) == 0 {
		return nil
	}

	block.Advance(endPos)

	link := ast.NewLink()
	link.Destination = content

	// Title will be parsed as the link's rels.
	switch direction {
	case -1:
		link.Title = []byte("down")
	case 1:
		link.Title = []byte("up")
	}

	return link
}
