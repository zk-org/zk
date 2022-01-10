package lsp

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
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
	if !ok {
		return ""
	}

	charIdx := int(pos.Character)
	if length > charIdx {
		return line[0:charIdx]
	}
	return line[(charIdx - length):charIdx]
}

// LookForward returns the n characters after the given position, on the same line.
func (d *document) LookForward(pos protocol.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	lineLength := len(line)
	charIdx := int(pos.Character)
	if lineLength <= charIdx+length {
		return line[charIdx:]
	}
	return line[charIdx:(charIdx + length)]
}

var wikiLinkRegex = regexp.MustCompile(`\[?\[\[(.+?)(?: *\| *(.+?))?\]\]`)
var markdownLinkRegex = regexp.MustCompile(`\[([^\]]+?[^\\])\]\((.+?[^\\])\)`)

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

			links = append(links, documentLink{
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
				HasTitle:   hasTitle,
				IsWikiLink: isWikiLink,
			})
		}

		for _, match := range markdownLinkRegex.FindAllStringSubmatchIndex(line, -1) {
			// Ignore embedded image, e.g. ![title](href.png)
			if match[0] > 0 && line[match[0]-1] == '!' {
				continue
			}

			href := line[match[4]:match[5]]
			// Valid Markdown links are percent-encoded.
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
	// https://github.com/mickael-menu/zk/issues/144#issuecomment-1006108485
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
	Href  string
	Range protocol.Range
	// HasTitle indicates whether this link has a title information. For
	// example [[filename]] doesn't but [[filename|title]] does.
	HasTitle bool
	// IsWikiLink indicates whether this link is a [[WikiLink]] instead of a
	// regular Markdown link.
	IsWikiLink bool
}
