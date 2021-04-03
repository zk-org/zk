package lsp

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mickael-menu/zk/core/note"
	strutil "github.com/mickael-menu/zk/util/strings"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/kutil/logging"
)

// document represents an opened file.
type document struct {
	URI     protocol.DocumentUri
	Content string
	Log     logging.Logger
	lines   []string
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *document) ApplyChanges(changes []interface{}) {
	for _, change := range changes {
		switch c := change.(type) {
		case protocol.TextDocumentContentChangeEvent:
			startIndex, endIndex := c.Range.IndexesIn(d.Content)
			d.Content = d.Content[:startIndex] + c.Text + d.Content[endIndex:]
		case protocol.TextDocumentContentChangeEventWhole:
			d.Content = c.Text
		}
	}

	d.lines = nil
}

var nonEmptyString = regexp.MustCompile(`\S+`)

// WordAt returns the word found at the given location.
// Credit https://github.com/aca/neuron-language-server/blob/450a7cff71c14e291ee85ff8a0614fa9d4dd5145/utils.go#L13
func (d *document) WordAt(pos protocol.Position) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	charIdx := int(pos.Character)
	wordIdxs := nonEmptyString.FindAllStringIndex(line, -1)
	for _, wordIdx := range wordIdxs {
		if wordIdx[0] <= charIdx && charIdx <= wordIdx[1] {
			return line[wordIdx[0]:wordIdx[1]]
		}
	}

	return ""
}

// GetLine returns the line at the given index.
func (d *document) GetLine(index int) (string, bool) {
	lines := d.GetLines()
	if index < 0 || index > len(lines) {
		return "", false
	}
	return lines[index], true
}

// GetLines returns all the lines in the document.
func (d *document) GetLines() []string {
	if d.lines == nil {
		// We keep \r on purpose, to avoid messing up position conversions.
		d.lines = strings.Split(d.Content, "\n")
	}
	return d.lines
}

// LookBehind returns the n characters before the given position, on the same line.
func (d *document) LookBehind(pos protocol.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	charIdx := int(pos.Character)
	if length > charIdx {
		return line[0:charIdx]
	}
	return line[(charIdx - length):charIdx]
}

var wikiLinkRegex = regexp.MustCompile(`\[?\[\[(.+?)(?:\|(.+?))?\]\]`)
var markdownLinkRegex = regexp.MustCompile(`\[([^\]]+?[^\\])\]\((.+?[^\\])\)`)

// DocumentLinkAt returns the internal or external link found in the document
// at the given position.
func (d *document) DocumentLinkAt(pos protocol.Position, basePath string, finder note.Finder) (*protocol.DocumentLink, error) {
	links, err := d.DocumentLinks(basePath, finder)
	if err != nil {
		return nil, err
	}

	for _, link := range links {
		if positionInRange(d.Content, link.Range, pos) {
			return &link, nil
		}
	}

	return nil, nil
}

// DocumentLinks returns all the internal and external links found in the
// document.
func (d *document) DocumentLinks(basePath string, finder note.Finder) ([]protocol.DocumentLink, error) {
	links := []protocol.DocumentLink{}

	lines := d.GetLines()
	for lineIndex, line := range lines {

		matches := []linkMatch{}
		appendMatch := func(href string, start, end int) {
			if href == "" {
				return
			}

			matches = append(matches, linkMatch{
				Href: href,
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      protocol.UInteger(lineIndex),
						Character: protocol.UInteger(start),
					},
					End: protocol.Position{
						Line:      protocol.UInteger(lineIndex),
						Character: protocol.UInteger(end),
					},
				},
			})
		}

		for _, match := range markdownLinkRegex.FindAllStringSubmatchIndex(line, -1) {
			href := line[match[4]:match[5]]
			appendMatch(href, match[0], match[1])
		}

		for _, match := range wikiLinkRegex.FindAllStringSubmatchIndex(line, -1) {
			href := line[match[2]:match[3]]
			appendMatch(href, match[0], match[1])
		}

		for _, match := range matches {
			link := protocol.DocumentLink{
				Range: match.Range,
			}

			if strutil.IsURL(match.Href) {
				link.Target = &match.Href
			} else {
				note, err := finder.FindByHref(match.Href)
				if err != nil {
					d.Log.Errorf("findByHref(%s): %s", match.Href, err.Error())
				}
				if note == nil {
					continue
				}
				link.Target = stringPtr("file://" + filepath.Join(basePath, note.Path))
				link.Tooltip = &note.Title
			}

			links = append(links, link)
		}
	}

	return links, nil
}

type linkMatch struct {
	Href  string
	Range protocol.Range
}
