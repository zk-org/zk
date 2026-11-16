package core

import (
	"testing"

	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestNoteURIString(t *testing.T) {
	tests := []struct {
		name     string
		uri      NoteURI
		expected string
	}{
		{
			name:     "notebook and path",
			uri:      NoteURI{NotebookPath: "/home/u/notes", Path: "a.md"},
			expected: "zk://?v=/home/u/notes&f=a.md",
		},
		{
			name:     "path separators are left readable",
			uri:      NoteURI{NotebookPath: "/home/u/notes", Path: "ref/sub/a.md"},
			expected: "zk://?v=/home/u/notes&f=ref/sub/a.md",
		},
		{
			name:     "line",
			uri:      NoteURI{NotebookPath: "/notes", Path: "a.md", Line: 12},
			expected: "zk://?v=/notes&f=a.md&l=12",
		},
		{
			name:     "line and column",
			uri:      NoteURI{NotebookPath: "/notes", Path: "a.md", Line: 12, Column: 4},
			expected: "zk://?v=/notes&f=a.md&l=12&c=4",
		},
		{
			name:     "a column alone is dropped",
			uri:      NoteURI{NotebookPath: "/notes", Path: "a.md", Column: 4},
			expected: "zk://?v=/notes&f=a.md",
		},
		{
			name:     "spaces are encoded as %20, not +",
			uri:      NoteURI{NotebookPath: "/my notes", Path: "a b.md"},
			expected: "zk://?v=/my%20notes&f=a%20b.md",
		},
		{
			name:     "a literal plus stays encoded",
			uri:      NoteURI{NotebookPath: "/notes", Path: "a+b.md"},
			expected: "zk://?v=/notes&f=a%2Bb.md",
		},
		{
			name:     "query separators are encoded",
			uri:      NoteURI{NotebookPath: "/notes", Path: "a&b=c#d.md"},
			expected: "zk://?v=/notes&f=a%26b%3Dc%23d.md",
		},
		{
			name:     "unicode is encoded",
			uri:      NoteURI{NotebookPath: "/notes", Path: "日本.md"},
			expected: "zk://?v=/notes&f=%E6%97%A5%E6%9C%AC.md",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.uri.String(), test.expected)
		})
	}
}

func TestParseNoteURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected NoteURI
	}{
		{
			name:     "notebook and path",
			uri:      "zk://?v=/home/u/notes&f=a.md",
			expected: NoteURI{NotebookPath: "/home/u/notes", Path: "a.md"},
		},
		{
			name:     "line and column",
			uri:      "zk://?v=/notes&f=ref/a.md&l=12&c=4",
			expected: NoteURI{NotebookPath: "/notes", Path: "ref/a.md", Line: 12, Column: 4},
		},
		{
			// Some launchers drop the empty authority.
			name:     "without the double slash",
			uri:      "zk:?v=/notes&f=a.md&l=12",
			expected: NoteURI{NotebookPath: "/notes", Path: "a.md", Line: 12},
		},
		{
			// Others normalize the empty authority to a root path.
			name:     "with a root path",
			uri:      "zk:///?v=/notes&f=a.md",
			expected: NoteURI{NotebookPath: "/notes", Path: "a.md"},
		},
		{
			name:     "parameter order is irrelevant",
			uri:      "zk://?c=4&l=12&f=a.md&v=/notes",
			expected: NoteURI{NotebookPath: "/notes", Path: "a.md", Line: 12, Column: 4},
		},
		{
			name:     "unknown parameters are ignored",
			uri:      "zk://?v=/notes&f=a.md&future=1",
			expected: NoteURI{NotebookPath: "/notes", Path: "a.md"},
		},
		{
			name:     "percent-encoded spaces",
			uri:      "zk://?v=/my%20notes&f=a%20b.md",
			expected: NoteURI{NotebookPath: "/my notes", Path: "a b.md"},
		},
		{
			name:     "a plus decodes to a space",
			uri:      "zk://?v=/notes&f=a+b.md",
			expected: NoteURI{NotebookPath: "/notes", Path: "a b.md"},
		},
		{
			name:     "unicode",
			uri:      "zk://?v=/notes&f=%E6%97%A5%E6%9C%AC.md",
			expected: NoteURI{NotebookPath: "/notes", Path: "日本.md"},
		},
		{
			name:     "paths are cleaned",
			uri:      "zk://?v=/notes/&f=ref/./sub/../a.md",
			expected: NoteURI{NotebookPath: "/notes", Path: "ref/a.md"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseNoteURI(test.uri)
			assert.Nil(t, err)
			assert.Equal(t, actual, test.expected)
		})
	}
}

func TestParseNoteURIRoundTrip(t *testing.T) {
	uris := []string{
		"zk://?v=/home/u/notes&f=a.md",
		"zk://?v=/home/u/notes&f=ref/sub/a.md&l=12",
		"zk://?v=/home/u/notes&f=ref/sub/a.md&l=12&c=4",
		"zk://?v=/my%20notes&f=a%20b.md&l=1&c=1",
		"zk://?v=/notes&f=%E6%97%A5%E6%9C%AC.md",
	}

	for _, uri := range uris {
		t.Run(uri, func(t *testing.T) {
			parsed, err := ParseNoteURI(uri)
			assert.Nil(t, err)
			assert.Equal(t, parsed.String(), uri)
		})
	}
}

func TestParseNoteURIRejects(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		err  string
	}{
		{
			name: "empty",
			uri:  "",
			err:  "not a zk:// URI",
		},
		{
			name: "another scheme",
			uri:  "https://example.com",
			err:  "not a zk:// URI",
		},
		{
			name: "a location in the path",
			uri:  "zk://notebook/a-note",
			err:  "unsupported zk:// URI, expected the form zk://?v=<notebook>&f=<path>",
		},
		{
			name: "an opaque location",
			uri:  "zk:notebook/a-note",
			err:  "unsupported zk:// URI, expected the form zk://?v=<notebook>&f=<path>",
		},
		{
			name: "missing notebook",
			uri:  "zk://?f=a.md",
			err:  "missing the 'v' (notebook path) parameter",
		},
		{
			name: "relative notebook",
			uri:  "zk://?v=notes&f=a.md",
			err:  "the 'v' parameter must be an absolute notebook path",
		},
		{
			name: "missing file",
			uri:  "zk://?v=/notes",
			err:  "missing the 'f' (file path) parameter",
		},
		{
			name: "escaping file path",
			uri:  "zk://?v=/notes&f=../../etc/passwd",
			err:  "the 'f' parameter must be a relative path inside the notebook",
		},
		{
			name: "escaping sibling path",
			uri:  "zk://?v=/notes&f=ref/../../sibling/a.md",
			err:  "the 'f' parameter must be a relative path inside the notebook",
		},
		{
			name: "absolute file path",
			uri:  "zk://?v=/notes&f=/etc/passwd",
			err:  "the 'f' parameter must be a relative path inside the notebook",
		},
		{
			name: "zero line",
			uri:  "zk://?v=/notes&f=a.md&l=0",
			err:  "the 'l' parameter must be a positive line number",
		},
		{
			name: "negative line",
			uri:  "zk://?v=/notes&f=a.md&l=-1",
			err:  "the 'l' parameter must be a positive line number",
		},
		{
			name: "non-numeric line",
			uri:  "zk://?v=/notes&f=a.md&l=abc",
			err:  "the 'l' parameter must be a positive line number",
		},
		{
			name: "zero column",
			uri:  "zk://?v=/notes&f=a.md&l=12&c=0",
			err:  "the 'c' parameter must be a positive column number",
		},
		{
			name: "malformed escape",
			uri:  "zk://?v=/notes&f=a%zz.md",
			err:  "not a valid zk:// URI",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseNoteURI(test.uri)
			assert.Err(t, err, test.err)
		})
	}
}
