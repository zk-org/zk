package lsp

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

// Test helper to extract just the hrefs from document links
func extractHrefs(doc *document) []string {
	links, _ := doc.DocumentLinks()
	hrefs := make([]string, len(links))
	for i, link := range links {
		hrefs[i] = link.Href
	}
	return hrefs
}

func TestDocumentLinks_EscapedBackticks(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedHrefs []string
	}{
		{
			name:          "link after escaped backtick on same line",
			content:       "Some text with \\` escaped and a [[wikilink]]",
			expectedHrefs: []string{"wikilink"},
		},
		{
			name:          "markdown link after escaped backtick",
			content:       "Here is \\` and a [link](target.md)",
			expectedHrefs: []string{"target.md"},
		},
		{
			name:          "link on next line after escaped backtick",
			content:       "Line with \\` escaped backtick\n[[link-on-next-line]]",
			expectedHrefs: []string{"link-on-next-line"},
		},
		{
			name:          "multiple escaped backticks",
			content:       "Text \\` with \\` multiple escaped [[wikilink]]",
			expectedHrefs: []string{"wikilink"},
		},
		{
			name:          "real inline code should still work",
			content:       "Text with `real code` and [[wikilink]]",
			expectedHrefs: []string{"wikilink"},
		},
		{
			name:          "link inside real inline code should be ignored",
			content:       "Text with `[[code-link]]` and [[real-link]]",
			expectedHrefs: []string{"real-link"},
		},
		{
			name:          "escaped backtick inside inline code",
			content:       "Text with `code \\` still code` and [[wikilink]]",
			expectedHrefs: []string{"wikilink"},
		},
		{
			name:          "mixed escaped and real backticks",
			content:       "\\` not code `real code` [[wikilink]]",
			expectedHrefs: []string{"wikilink"},
		},
		{
			name:          "escaped backtick at end of line affects next line",
			content:       "Line ending with \\`\n[[link-that-should-be-found]]\n[[another-link]]",
			expectedHrefs: []string{"link-that-should-be-found", "another-link"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &document{
				Content: tt.content,
				Path:    "/test/note.md",
			}
			hrefs := extractHrefs(doc)
			assert.Equal(t, hrefs, tt.expectedHrefs)
		})
	}
}

func TestLinkWithinInlineCode_EscapedBackticks(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		linkStart    int
		linkEnd      int
		insideInline bool
		expected     bool
	}{
		{
			name:         "link after escaped backtick",
			line:         "\\` [[link]]",
			linkStart:    3,
			linkEnd:      11,
			insideInline: false,
			expected:     false,
		},
		{
			name:         "link after real backtick",
			line:         "` [[link]]",
			linkStart:    2,
			linkEnd:      10,
			insideInline: false,
			expected:     true,
		},
		{
			name:         "link after real inline code span",
			line:         "`code` [[link]]",
			linkStart:    7,
			linkEnd:      15,
			insideInline: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linkWithinInlineCode(tt.line, tt.linkStart, tt.linkEnd, tt.insideInline)
			assert.Equal(t, result, tt.expected)
		})
	}
}
