package lsp

import (
	"regexp"
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/kutil/logging"
)

type document struct {
	URI     protocol.DocumentUri
	Content string
	Log     logging.Logger
}

func (d document) ApplyChanges(changes []interface{}) document {
	for _, change := range changes {
		switch c := change.(type) {
		case protocol.TextDocumentContentChangeEvent:
			startIndex, endIndex := c.Range.IndexesIn(d.Content)
			d.Content = d.Content[:startIndex] + c.Text + d.Content[endIndex:]
		case protocol.TextDocumentContentChangeEventWhole:
			d.Content = c.Text
		}
	}

	return d
}

var nonEmptyString = regexp.MustCompile(`\S+`)

// Credit https://github.com/aca/neuron-language-server/blob/450a7cff71c14e291ee85ff8a0614fa9d4dd5145/utils.go#L13
func (d document) WordAt(line int, char int) string {
	lines := strings.Split(d.Content, "\n")
	if line < 0 || line > len(lines) {
		return ""
	}

	curLine := lines[line]
	wordIdxs := nonEmptyString.FindAllStringIndex(curLine, -1)
	for _, wordIdx := range wordIdxs {
		if wordIdx[0] <= char && char <= wordIdx[1] {
			return curLine[wordIdx[0]:wordIdx[1]]
		}
	}

	return ""
}
