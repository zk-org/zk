package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gmext "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/zk-org/zk/internal/adapter/markdown/extensions"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
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

func (s *documentStore) normalizePath(pathOrURI string) (string, error) {
	path, err := uriToPath(pathOrURI)
	if err != nil {
		return "", fmt.Errorf("unable to parse URI: %s: %w", pathOrURI, err)
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
func (d *document) ApplyChanges(changes []any) {
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

// documentParser is a goldmark parser configured for extracting links.
var documentParser = goldmark.New(
	goldmark.WithExtensions(
		gmext.Footnote,
		extensions.WikiLinkExt,
		extensions.MarkdownLinkExt,
	),
)

// byteOffsetToPosition converts a byte offset in source to a protocol.Position (line, character).
// lineOffsets must contain the byte offset of the start of each line.
func byteOffsetToPosition(offset int, source []byte, lineOffsets []int) protocol.Position {
	line := 0
	for i := 1; i < len(lineOffsets); i++ {
		if lineOffsets[i] > offset {
			break
		}
		line = i
	}
	lineStart := lineOffsets[line]

	// Convert byte offset within line to rune index.
	var lineContent string
	if line+1 < len(lineOffsets) {
		lineContent = string(source[lineStart : lineOffsets[line+1]-1]) // -1 to exclude newline.
	} else {
		lineContent = string(source[lineStart:]) // Last line.
	}
	byteOffsetInLine := offset - lineStart
	charPos := strutil.ByteIndexToRuneIndex(lineContent, byteOffsetInLine)

	return protocol.Position{
		Line:      protocol.UInteger(line),
		Character: protocol.UInteger(charPos),
	}
}

// buildLineOffsets returns a slice where lineOffsets[i] is the byte offset of line i.
func buildLineOffsets(source []byte) []int {
	offsets := []int{0}
	for i, b := range source {
		if b == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

// DocumentLinks returns all the internal and external links found in the
// document.
func (d *document) DocumentLinks() ([]documentLink, error) {
	links := []documentLink{}
	source := []byte(d.Content)
	lineOffsets := buildLineOffsets(source)

	reader := text.NewReader(source)
	context := parser.NewContext()
	root := documentParser.Parser().Parse(reader, parser.WithContext(context))

	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch link := n.(type) {
		case *extensions.MarkdownLink:
			href := string(link.Destination)
			if href == "" {
				return ast.WalkContinue, nil
			}

			if strings.HasPrefix(href, "file:///") || strings.HasPrefix(href, "magnet:?") {
				return ast.WalkContinue, nil
			}

			// Decode the href if it's percent-encoded
			if decodedHref, err := url.PathUnescape(href); err == nil {
				href = decodedHref
			}

			links = append(links, documentLink{
				Href:          href,
				RelativeToDir: filepath.Dir(d.Path),
				Range: protocol.Range{
					Start: byteOffsetToPosition(link.StartOffset, source, lineOffsets),
					End:   byteOffsetToPosition(link.EndOffset, source, lineOffsets),
				},
				IsWikiLink: false,
			})

		case *extensions.WikiLink:
			href := string(link.Destination)
			if href == "" {
				return ast.WalkContinue, nil
			}

			links = append(links, documentLink{
				Href:          href,
				RelativeToDir: filepath.Dir(d.Path),
				Range: protocol.Range{
					Start: byteOffsetToPosition(link.StartOffset, source, lineOffsets),
					End:   byteOffsetToPosition(link.EndOffset, source, lineOffsets),
				},
				HasTitle:   len(link.Title) > 0,
				IsWikiLink: true,
			})
		}

		return ast.WalkContinue, nil
	})

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
	if string(targetWord[0]) == "#" {
		targetWord = targetWord[1:]
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
