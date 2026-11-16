package core

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// NoteURIScheme is the URI scheme used to address notes from outside systems.
const NoteURIScheme = "zk"

// NoteURI addresses a location inside a notebook, in a format suitable for
// exchange with systems outside of zk.
//
//	zk://?v=<notebook path>&f=<path relative to the notebook>&l=<line>&c=<column>
type NoteURI struct {
	// Absolute path to the notebook (vault) directory.
	NotebookPath string
	// Path to the file, relative to the notebook root, using forward slashes.
	Path string
	// 1-based line number, 0 when unspecified.
	Line int
	// 1-based column number, 0 when unspecified.
	Column int
}

// ParseNoteURI parses a zk:// URI. Both the zk://?… and zk:?… forms are
// accepted, as some launchers rewrite one into the other.
//
// Only the URI itself is validated. It is up to the caller to check that the
// notebook and the file actually exist.
func ParseNoteURI(uri string) (NoteURI, error) {
	res := NoteURI{}

	parsed, err := url.Parse(uri)
	if err != nil {
		return res, fmt.Errorf("%s: not a valid zk:// URI: %w", uri, err)
	}
	if parsed.Scheme != NoteURIScheme {
		return res, fmt.Errorf("%s: not a zk:// URI", uri)
	}
	// An empty path and a lone "/" are both the authority-less form we expect.
	// Anything else means the URI carries the location in its path, which is
	// not a shape we support.
	if parsed.Opaque != "" || parsed.Host != "" || (parsed.Path != "" && parsed.Path != "/") {
		return res, fmt.Errorf("%s: unsupported zk:// URI, expected the form zk://?v=<notebook>&f=<path>", uri)
	}

	// Not parsed.Query(), which silently drops malformed escapes.
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return res, fmt.Errorf("%s: not a valid zk:// URI: %w", uri, err)
	}

	res.NotebookPath = query.Get("v")
	if res.NotebookPath == "" {
		return res, fmt.Errorf("%s: zk:// URI is missing the 'v' (notebook path) parameter", uri)
	}
	if !filepath.IsAbs(res.NotebookPath) && !strings.HasPrefix(res.NotebookPath, "/") {
		return res, fmt.Errorf("%s: the 'v' parameter must be an absolute notebook path", res.NotebookPath)
	}
	res.NotebookPath = filepath.Clean(filepath.FromSlash(res.NotebookPath))

	res.Path = query.Get("f")
	if res.Path == "" {
		return res, fmt.Errorf("%s: zk:// URI is missing the 'f' (file path) parameter", uri)
	}
	// The URI comes from an outside system, so make sure it can't point
	// anywhere but inside the notebook. filepath.IsLocal rejects absolute
	// paths, ".." escapes and, on Windows, reserved device names.
	res.Path = path.Clean(res.Path)
	if !filepath.IsLocal(filepath.FromSlash(res.Path)) {
		return res, fmt.Errorf("%s: the 'f' parameter must be a relative path inside the notebook", query.Get("f"))
	}

	res.Line, err = parseURIPosition(query.Get("l"), "l", "line")
	if err != nil {
		return res, err
	}
	res.Column, err = parseURIPosition(query.Get("c"), "c", "column")
	if err != nil {
		return res, err
	}

	return res, nil
}

// String formats the URI in its zk://?v=…&f=…&l=…&c=… form.
//
// A column is only emitted alongside a line, as no editor can act on a column
// alone.
func (u NoteURI) String() string {
	var res strings.Builder
	res.WriteString(NoteURIScheme)
	res.WriteString("://?v=")
	res.WriteString(escapeURIParam(filepath.ToSlash(u.NotebookPath)))
	if u.Path != "" {
		res.WriteString("&f=")
		res.WriteString(escapeURIParam(filepath.ToSlash(u.Path)))
	}
	if u.Line > 0 {
		res.WriteString("&l=")
		res.WriteString(strconv.Itoa(u.Line))
		if u.Column > 0 {
			res.WriteString("&c=")
			res.WriteString(strconv.Itoa(u.Column))
		}
	}
	return res.String()
}

// escapeURIParam percent-encodes a query parameter value, but keeps path
// separators as-is to make the URI readable when pasted around.
func escapeURIParam(value string) string {
	value = url.QueryEscape(value)
	value = strings.ReplaceAll(value, "%2F", "/")
	// QueryEscape encodes a space as "+", which reads as a literal plus. A
	// literal plus is already encoded as "%2B", so this is unambiguous.
	return strings.ReplaceAll(value, "+", "%20")
}

// parseURIPosition parses a line or column parameter, which must be a positive
// number when present. A producer counting from zero fails loudly here instead
// of being off by one on every link.
func parseURIPosition(value string, key string, name string) (int, error) {
	if value == "" {
		return 0, nil
	}
	pos, err := strconv.Atoi(value)
	if err != nil || pos < 1 {
		return 0, fmt.Errorf("%s: the '%s' parameter must be a positive %s number", value, key, name)
	}
	return pos, nil
}
