package extensions

import (
	"strings"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// WikiLinkExt is an extension parsing wiki links and Neuron's Folgezettel.
//
// For example, [[wiki link]], [[[legacy downlink]]], #[[uplink]], [[downlink]]#.
var WikiLinkExt = &wikiLink{}

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

	var (
		href  string
		label string
		rel   core.LinkRelation
	)

	var (
		opened          = false // Found at least [[
		closed          = false // Found at least ]]
		escaping        = false // Found a backslash, next character will be literal
		parsingLabel    = false // Found a | in a Wikilink, now we parse the link's label
		openerCharCount = 0     // Number of [ encountered
		closerCharCount = 0     // Number of ] encountered
		endPos          = 0     // Last position of the link in the line
	)

	appendChar := func(c byte) {
		if parsingLabel {
			label += string(c)
		} else {
			href += string(c)
		}
	}

	for i, char := range line {
		endPos = i

		if closed {
			// Supports trailing hash syntax for Neuron's Folgezettel, e.g. [[id]]#
			if char == '#' {
				rel = core.LinkRelationDown
			}
			break
		}

		if !opened {
			switch char {
			// Supports leading hash syntax for Neuron's Folgezettel, e.g. #[[id]]
			case '#':
				rel = core.LinkRelationUp
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

		if !escaping {
			switch char {

			case '|': // [[href | label]]
				parsingLabel = true
				continue

			case '\\':
				escaping = true
				continue

			case ']':
				closerCharCount += 1
				if closerCharCount == openerCharCount {
					closed = true
					// Neuron's legacy [[[Folgezettel]]].
					if closerCharCount == 3 {
						rel = core.LinkRelationDown
					}
				}
				continue
			}
		}
		escaping = false

		// Found incomplete number of closing brackets to close the link.
		// We add them to the HREF and reset the count.
		if closerCharCount > 0 {
			for i := 0; i < closerCharCount; i++ {
				appendChar(']')
			}
			closerCharCount = 0
		}
		appendChar(char)
	}

	if !closed || len(href) == 0 {
		return nil
	}

	block.Advance(endPos)

	href = strings.TrimSpace(href)
	label = strings.TrimSpace(label)
	if len(label) == 0 {
		label = href
	}

	link := ast.NewLink()
	link.Destination = []byte(href)
	// Title will be parsed as the link's rel by the Markdown parser.
	link.Title = []byte(rel)
	link.AppendChild(link, ast.NewString([]byte(label)))

	return link
}
