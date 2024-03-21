package lsp

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/errors"
	strutil "github.com/zk-org/zk/internal/util/strings"
)

// documentStore holds opened documents.
type documentStore struct {
	documents map[string]*document
	fs        core.FileStorage
	logger    util.Logger
}

func newDocumentStore(fs core.FileStorage, logger util.Logger) *documentStore {
	return &documentStore{
		documents: map[string]*document{},
		fs:        fs,
		logger:    logger,
	}
}

func (s *documentStore) DidOpen(params protocol.DidOpenTextDocumentParams, notify glsp.NotifyFunc) (*document, error) {
	langID := params.TextDocument.LanguageID
	if langID != "markdown" && langID != "vimwiki" && langID != "pandoc" {
		return nil, nil
	}

	uri := params.TextDocument.URI
	path, err := s.normalizePath(uri)
	if err != nil {
		return nil, err
	}
	doc := &document{
		URI:     uri,
		Path:    path,
		Content: params.TextDocument.Text,
	}
	s.documents[path] = doc
	return doc, nil
}

func (s *documentStore) Close(uri protocol.DocumentUri) {
	delete(s.documents, uri)
}

func (s *documentStore) Get(pathOrURI string) (*document, bool) {
	path, err := s.normalizePath(pathOrURI)
	if err != nil {
		s.logger.Err(err)
		return nil, false
	}
	d, ok := s.documents[path]
	return d, ok
}

func (s *documentStore) normalizePath(pathOrUri string) (string, error) {
	path, err := uriToPath(pathOrUri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse URI: %s", pathOrUri)
	}
	return s.fs.Canonical(path), nil
}

// document represents an opened file.
type document struct {
	URI                     protocol.DocumentUri
	Path                    string
	NeedsRefreshDiagnostics bool
	Content                 string
	lines                   []string
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

// WordAt returns the word found at the given location.
func (d *document) WordAt(pos protocol.Position) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}
	return strutil.WordAt(line, int(pos.Character))
}

// ContentAtRange returns the document text at given range.
func (d *document) ContentAtRange(rng protocol.Range) string {
	startIndex, endIndex := rng.IndexesIn(d.Content)
	return d.Content[startIndex:endIndex]
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
	utf16Bytes := utf16.Encode([]rune(line))
	if !ok {
		return ""
	}

	charIdx := int(pos.Character)
	if length > charIdx {
		return string(utf16.Decode(utf16Bytes[0:charIdx]))
	}
	return string(utf16.Decode(utf16Bytes[(charIdx - length):charIdx]))
}

// LookForward returns the n characters after the given position, on the same line.
func (d *document) LookForward(pos protocol.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	utf16Bytes := utf16.Encode([]rune(line))
	if !ok {
		return ""
	}

	lineLength := len(utf16Bytes)
	charIdx := int(pos.Character)
	if lineLength <= charIdx+length {
		return string(utf16.Decode(utf16Bytes[charIdx:]))
	}
	return string(utf16.Decode(utf16Bytes[charIdx:(charIdx + length)]))
}

var wikiLinkRegex = regexp.MustCompile(`\[?\[\[(.+?)(?: *\| *(.+?))?\]\]`)
var markdownLinkRegex = regexp.MustCompile(`\[([^\]]+?[^\\])\]\((.+?[^\\])\)`)
var fileURIregex = regexp.MustCompile(`file:///`)

// LinkFromRoot returns a Link to this document from the root of the given
// notebook.
func (d *document) LinkFromRoot(nb *core.Notebook) (*documentLink, error) {
	href, err := nb.RelPath(d.Path)
	if err != nil {
		return nil, err
	}
	return &documentLink{
		Href:          href,
		RelativeToDir: nb.Path,
	}, nil
}

// DocumentLinkAt returns the internal or external link found in the document
// at the given position.
func (d *document) DocumentLinkAt(pos protocol.Position) (*documentLink, error) {
	links, err := d.DocumentLinks()
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
func (d *document) DocumentLinks() ([]documentLink, error) {
	links := []documentLink{}

	lines := d.GetLines()
	for lineIndex, line := range lines {

		appendLink := func(href string, start, end int, hasTitle bool, isWikiLink bool) {
			if href == "" {
				return
			}

			// Go regexes work with bytes, but the LSP client expects character indexes.
			start = strutil.ByteIndexToRuneIndex(line, start)
			end = strutil.ByteIndexToRuneIndex(line, end)

			links = append(links, documentLink{
				Href:          href,
				RelativeToDir: filepath.Dir(d.Path),
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
				HasTitle:   hasTitle,
				IsWikiLink: isWikiLink,
			})
		}

		// extract link paths from [title](path) patterns
		// note: match[0:1] is the entire match, match[2:3] is the contents of
		// brackets, match[4:5] is contents of parentheses
		for _, match := range markdownLinkRegex.FindAllStringSubmatchIndex(line, -1) {

			// Ignore embedded images ![title](file.png)
			if match[0] > 0 && line[match[0]-1] == '!' {
				continue
			}

			// ignore tripple dash file URIs [title](file:///foo.go)
			if match[5]-match[4] >= 8 {
				linkURL := line[match[4]:match[5]]
				fileURIresult := linkURL[:8]
				if fileURIregex.MatchString(fileURIresult) {
					continue
				}
			}

			href := line[match[4]:match[5]]

			// Decode the href if it's percent-encoded
			if decodedHref, err := url.PathUnescape(href); err == nil {
				href = decodedHref
			}

			appendLink(href, match[0], match[1], false, false)
		}

		for _, match := range wikiLinkRegex.FindAllStringSubmatchIndex(line, -1) {
			href := line[match[2]:match[3]]
			hasTitle := match[4] != -1
			appendLink(href, match[0], match[1], hasTitle, true)
		}
	}

	return links, nil
}

// IsTagPosition returns whether the given caret position is inside a tag (YAML frontmatter, #hashtag, etc.).
func (d *document) IsTagPosition(position protocol.Position, noteContentParser core.NoteContentParser) bool {
	lines := strutil.CopyList(d.GetLines())
	lineIdx := int(position.Line)
	charIdx := int(position.Character)
	line := lines[lineIdx]
	// https://github.com/zk-org/zk/issues/144#issuecomment-1006108485
	line = line[:charIdx] + "ZK_PLACEHOLDER" + line[charIdx:]
	lines[lineIdx] = line
	targetWord := strutil.WordAt(line, charIdx)
	if targetWord == "" {
		return false
	}

	content := strings.Join(lines, "\n")
	note, err := noteContentParser.ParseNoteContent(content)
	if err != nil {
		return false
	}
	return strutil.Contains(note.Tags, targetWord)
}

type documentLink struct {
	Href          string
	RelativeToDir string
	Range         protocol.Range
	// HasTitle indicates whether this link has a title information. For
	// example [[filename]] doesn't but [[filename|title]] does.
	HasTitle bool
	// IsWikiLink indicates whether this link is a [[WikiLink]] instead of a
	// regular Markdown link.
	IsWikiLink bool
}
