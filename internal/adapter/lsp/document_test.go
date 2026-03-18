package lsp

import (
	"regexp"
	"testing"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/core"
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
		{
			name:          "link with newline in text",
			content:       "[link\ntext](dest.md)",
			expectedHrefs: []string{"dest.md"},
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

func TestDocument_LookBehind(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		pos      protocol.Position
		length   int
		expected string
	}{
		{
			name:    "empty",
			content: "",
			pos: protocol.Position{
				Line:      0,
				Character: 0,
			},
			length:   0,
			expected: "",
		},
		{
			name:     "normal case",
			content:  "hello world",
			pos:      protocol.Position{Line: 0, Character: 5},
			length:   2,
			expected: "lo",
		},
		{
			name:     "normal case",
			content:  "hello world",
			pos:      protocol.Position{Line: 0, Character: 5},
			length:   2,
			expected: "lo",
		},
		{
			name:     "out of bound line index",
			content:  "lorem\nipsum\ndor\nsit\n",
			pos:      protocol.Position{Line: 6, Character: 0},
			length:   0,
			expected: "",
		},
		{
			name:     "out of bound length",
			content:  "hello world",
			pos:      protocol.Position{Line: 0, Character: 5},
			length:   10,
			expected: "hello",
		},
		{
			name:     "out of bound character",
			content:  "#abc #z",
			pos:      protocol.Position{Line: 0, Character: 8},
			length:   8,
			expected: "#abc #z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &document{
				Content: tt.content,
				Path:    "/test/note.md",
			}
			actual := doc.LookBehind(tt.pos, tt.length)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestDocument_LookForward(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		pos      protocol.Position
		length   int
		expected string
	}{
		{
			name:    "empty",
			content: "",
			pos: protocol.Position{
				Line:      0,
				Character: 0,
			},
			length:   0,
			expected: "",
		},
		{
			name:     "normal case",
			content:  "hello world",
			pos:      protocol.Position{Line: 0, Character: 5},
			length:   2,
			expected: " w",
		},
		{
			name:     "out of bound line index",
			content:  "lorem\nipsum\ndor\nsit\n",
			pos:      protocol.Position{Line: 6, Character: 0},
			length:   0,
			expected: "",
		},
		{
			name:     "out of bound length",
			content:  "hello world",
			pos:      protocol.Position{Line: 0, Character: 5},
			length:   10,
			expected: " world",
		},
		{
			name:     "out of bound character",
			content:  "#abc #z",
			pos:      protocol.Position{Line: 0, Character: 100},
			length:   8,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &document{
				Content: tt.content,
				Path:    "/test/note.md",
			}
			actual := doc.LookForward(tt.pos, tt.length)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

type mockNoteContentParser struct{}

func (m *mockNoteContentParser) ParseNoteContent(content string) (*core.NoteContent, error) {
	tags := []string{}
	// This mock parser only finds hashtags.
	// It's important to match the behavior where ZK_PLACEHOLDER is part of the tag.
	re := regexp.MustCompile(`#[^ \t\n\f\r,;\[\]\"\']+`)
	matches := re.FindAllString(content, -1)
	for _, match := range matches {
		tags = append(tags, match[1:]) // strip #
	}
	return &core.NoteContent{Tags: tags}, nil
}

func TestDocument_IsTagPosition(t *testing.T) {
	parser := &mockNoteContentParser{}
	tests := []struct {
		name     string
		content  string
		pos      protocol.Position
		expected bool
	}{
		{
			name:    "at the start of a hashtag (#)",
			content: "hello #tag world",
			pos: protocol.Position{
				Line:      0,
				Character: 6, // at #
			},
			expected: false, // ZK_PLACEHOLDER is inserted BEFORE #, so it's "ZK_PLACEHOLDER#tag" which is not a tag
		},
		{
			name:    "just after #",
			content: "hello #tag world",
			pos: protocol.Position{
				Line:      0,
				Character: 7, // after #
			},
			expected: true,
		},
		{
			name:    "inside a hashtag",
			content: "hello #tag world",
			pos: protocol.Position{
				Line:      0,
				Character: 8, // at a
			},
			expected: true,
		},
		{
			name:    "at the end of a hashtag",
			content: "hello #tag world",
			pos: protocol.Position{
				Line:      0,
				Character: 10, // after g
			},
			expected: true,
		},
		{
			name:    "not in a tag (at 'h')",
			content: "hello #tag world",
			pos: protocol.Position{
				Line:      0,
				Character: 0, // at h
			},
			expected: false,
		},
		{
			name:    "between two tags (at space)",
			content: "#tag1 #tag2",
			pos: protocol.Position{
				Line:      0,
				Character: 5, // space between
			},
			expected: true, // WordAt returns "#tag1" at index 5
		},
		{
			name:    "in multiple lines",
			content: "line1\n#tag line2",
			pos: protocol.Position{
				Line:      1,
				Character: 2, // at 't' in #tag
			},
			expected: true,
		},
		{
			name:    "out of bound line",
			content: "line1\n#tag line2",
			pos: protocol.Position{
				Line:      10,
				Character: 2,
			},
			expected: false,
		},
		{
			name:    "out of bound character",
			content: "line1\n#tag line2",
			pos: protocol.Position{
				Line:      1,
				Character: 100,
			},
			expected: false,
		},
		{
			name:    "at end of line (true for completion)",
			content: "#zk",
			pos: protocol.Position{
				Line:      0,
				Character: 3,
			},
			expected: true,
		},
		{
			name:    "utf-16 test",
			content: "#雨果",
			pos: protocol.Position{
				Line:      0,
				Character: 2,
			},
			expected: true,
		},
		{
			name:    "utf-16 test out of bounds",
			content: "#雨果",
			pos: protocol.Position{
				Line:      0,
				Character: 4,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &document{
				Content: tt.content,
				Path:    "/test/note.md",
			}
			actual := doc.IsTagPosition(tt.pos, parser)
			assert.Equal(t, actual, tt.expected)
		})
	}
}

func TestDocument_WordAt(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		pos      protocol.Position
		expected string
	}{
		{
			name:    "empty",
			content: "",
			pos: protocol.Position{
				Line:      0,
				Character: 0,
			},
			expected: "",
		},
		{
			name:     "normal case",
			content:  "Hello World",
			pos:      protocol.Position{Line: 0, Character: 8},
			expected: "World",
		},
		{
			name:     "should handle utf-16 correctly",
			content:  "ビール 焼酎 ジン ワイン ウィスキー",
			pos:      protocol.Position{Line: 0, Character: 8},
			expected: "ジン",
		},
		{
			name:     "when Character is out-of-bound, it treats as end of line",
			content:  "ビール 焼酎 ジン ワイン ウィスキー",
			pos:      protocol.Position{Line: 0, Character: 100},
			expected: "ウィスキー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &document{
				Content: tt.content,
				Path:    "/test/note.md",
			}
			actual := doc.WordAt(tt.pos)
			assert.Equal(t, actual, tt.expected)
		})
	}
}
