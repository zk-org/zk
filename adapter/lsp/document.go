package lsp

import (
	"regexp"

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
	if d.lines == nil {
		d.lines = strutil.SplitLines(d.Content)
	}
	if index < 0 || index > len(d.lines) {
		return "", false
	}
	return d.lines[index], true
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
