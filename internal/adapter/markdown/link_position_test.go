package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestLinkPositionTransformer(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantStart int
		wantEnd   int
		wantDest  string
	}{
		{
			name:      "simple link",
			source:    "Hello [world](http://example.com) there",
			wantStart: 6,
			wantEnd:   33,
			wantDest:  "http://example.com",
		},
		{
			name:      "link with newline",
			source:    "[Hello\nworld](http://example.com) there",
			wantStart: 0,
			wantEnd:   33,
			wantDest:  "http://example.com",
		},
		{
			name:      "link at start",
			source:    "[link](dest) after",
			wantStart: 0,
			wantEnd:   12,
			wantDest:  "dest",
		},
		{
			name:      "link at end",
			source:    "before [link](dest)",
			wantStart: 7,
			wantEnd:   19,
			wantDest:  "dest",
		},
		{
			name:      "link with title",
			source:    `[text](http://example.com "title")`,
			wantStart: 0,
			wantEnd:   34,
			wantDest:  "http://example.com",
		},
		{
			name:      "multiple links first",
			source:    "[a](1) [b](2)",
			wantStart: 0,
			wantEnd:   6,
			wantDest:  "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := goldmark.New()

			source := []byte(tt.source)
			reader := text.NewReader(source)
			root := md.Parser().Parse(reader, parser.WithContext(parser.NewContext()))

			// Find link by destination (some cases have multiple links).
			var found bool
			ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if !entering || found {
					return ast.WalkContinue, nil
				}

				if link, ok := n.(*ast.Link); ok {
					if string(link.Destination) != tt.wantDest {
						return ast.WalkContinue, nil
					}
					found = true
					pos := GetLinkPosition(link, source)
					assert.NotNil(t, pos)

					assert.Equal(t, pos.Start, tt.wantStart)
					assert.Equal(t, pos.End, tt.wantEnd)
					assert.Equal(t, string(link.Destination), tt.wantDest)
				}
				return ast.WalkContinue, nil
			})

			assert.True(t, found)
		})
	}
}

func TestLinkPositionMultipleLinks(t *testing.T) {
	source := []byte("[first](1) and [second](2)")
	md := goldmark.New()

	reader := text.NewReader(source)
	root := md.Parser().Parse(reader, parser.WithContext(parser.NewContext()))

	positions := make(map[string]LinkPosition)

	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if link, ok := n.(*ast.Link); ok {
			pos := GetLinkPosition(link, source)
			if pos != nil {
				positions[string(link.Destination)] = *pos
			}
		}
		return ast.WalkContinue, nil
	})

	assert.Equal(t, len(positions), 2)

	pos1, ok := positions["1"]
	assert.True(t, ok)
	assert.Equal(t, pos1.Start, 0)
	assert.Equal(t, pos1.End, 10)

	pos2, ok := positions["2"]
	assert.True(t, ok)
	assert.Equal(t, pos2.Start, 15)
	assert.Equal(t, pos2.End, 26)
}
