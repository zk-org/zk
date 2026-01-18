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

// Test that links inside footnotes are be detected as wiki-links.
// See https://github.com/zk-org/zk/issues/574
func TestDocumentLinks_FootnoteWikiLinks(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedHrefs []string
	}{
		{
			name:          "wiki link only in footnote",
			content:       "# Title note 2\n\nsome content[^1]\n\n[^1]: [[note1]]",
			expectedHrefs: []string{"note1"},
		},
		{
			name:          "wiki link with surrounding text",
			content:       "Text[^1].\n\n[^1]: See [[note1]] for details.",
			expectedHrefs: []string{"note1"},
		},
		{
			name:          "multiple wiki links in footnote",
			content:       "Text[^1].\n\n[^1]: See [[note1]] and [[note2]].",
			expectedHrefs: []string{"note1", "note2"},
		},
		{
			name:          "markdown link in footnote",
			content:       "Text[^1].\n\n[^1]: See [docs](readme.md).",
			expectedHrefs: []string{"readme.md"},
		},
		{
			name:          "wiki link with title in footnote",
			content:       "Text[^1].\n\n[^1]: [[note1|Note Title]]",
			expectedHrefs: []string{"note1"},
		},
		{
			name:          "multiple footnotes with links",
			content:       "A[^1] and B[^2].\n\n[^1]: [[note1]]\n[^2]: [[note2]]",
			expectedHrefs: []string{"note1", "note2"},
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

func TestDocumentLinks_HTMLComments(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedHrefs []string
	}{
		{
			name:          "link inside HTML comment should be ignored",
			content:       "<!-- [commented link](ignored.md) -->\n[[real-link]]",
			expectedHrefs: []string{"real-link"},
		},
		{
			name:          "wikilink inside HTML comment should be ignored",
			content:       "<!-- [[commented-wiki]] -->\n[real](real.md)",
			expectedHrefs: []string{"real.md"},
		},
		{
			name:          "link before and after HTML comment",
			content:       "[[before]]\n<!-- [[ignored]] -->\n[[after]]",
			expectedHrefs: []string{"before", "after"},
		},
		{
			name:          "multiline HTML comment with link",
			content:       "<!--\n[link](ignored.md)\n-->\n[[visible]]",
			expectedHrefs: []string{"visible"},
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

func TestDocumentLinks_CodeBlocks(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedHrefs []string
	}{
		{
			name:          "link inside fenced code block should be ignored",
			content:       "```\n[[code-link]]\n```\n[[real-link]]",
			expectedHrefs: []string{"real-link"},
		},
		{
			name:          "link inside indented code block should be ignored",
			content:       "Normal text\n\n    [[indented-code]]\n\n[[real-link]]",
			expectedHrefs: []string{"real-link"},
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

func TestDocumentLinks_ComplexContent(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedHrefs []string
	}{
		{
			name:          "link with bold text",
			content:       "[**bold link**](dest.md)",
			expectedHrefs: []string{"dest.md"},
		},
		{
			name:          "link with italic text",
			content:       "[*italic link*](dest.md)",
			expectedHrefs: []string{"dest.md"},
		},
		{
			name:          "link with inline code",
			content:       "[`code` link](dest.md)",
			expectedHrefs: []string{"dest.md"},
		},
		{
			name:          "link with parentheses in destination",
			content:       "[link](path/with(parens).md)",
			expectedHrefs: []string{"path/with(parens).md"},
		},
		{
			name:          "multiple complex links",
			content:       "[**bold**](one.md) and [*italic*](two.md)",
			expectedHrefs: []string{"one.md", "two.md"},
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
