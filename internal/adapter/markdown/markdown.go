package markdown

import (
	"bufio"

	"net/url"
	"strings"

	"github.com/mvdan/xurls"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/zk-org/zk/internal/adapter/markdown/extensions"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/opt"
	strutil "github.com/zk-org/zk/internal/util/strings"
	yaml "github.com/zk-org/zk/internal/util/yaml"
)

// Parser parses the content of Markdown notes.
type Parser struct {
	md     goldmark.Markdown
	logger util.Logger
}

type ParserOpts struct {
	// Indicates whether #hashtags are parsed.
	HashtagEnabled bool
	// Indicates whether Bear's multi-word tags are parsed. Hashtags must be enabled as well.
	MultiWordTagEnabled bool
	// Indicates whether :colon:tags: are parsed.
	ColontagEnabled bool
}

type FrontMatterExtended = yaml.FrontmatterExtended

// NewParser creates a new Markdown Parser.
func NewParser(options ParserOpts, logger util.Logger) *Parser {
	return &Parser{
		md: goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
				extension.Footnote,
				extension.NewLinkify(
					extension.WithLinkifyAllowedProtocols([][]byte{
						[]byte("http:"),
						[]byte("https:"),
					}),
					extension.WithLinkifyURLRegexp(
						xurls.Strict,
					),
				),
				extensions.WikiLinkExt,
				&extensions.TagExt{
					HashtagEnabled:      options.HashtagEnabled,
					MultiWordTagEnabled: options.MultiWordTagEnabled,
					ColontagEnabled:     options.ColontagEnabled,
				},
			),
		),
		logger: logger,
	}
}

func tagNames(parsedTags []core.Tag) []string {
	if parsedTags == nil {
		return []string{}
	}
	check := make(map[string]struct{})
	res := make([]string, 0)

	for _, tag := range parsedTags {
		tagName := tag.Name
		if _, ok := check[tagName]; ok {
			continue
		}
		check[tagName] = struct{}{}
		res = append(res, tagName)
	}
	return res
}

// ParseNoteContent implements core.NoteContentParser.
func (p *Parser) ParseNoteContent(content string) (*core.NoteContent, error) {
	bytes := []byte(content)

	context := parser.NewContext()
	root := p.md.Parser().Parse(
		text.NewReader(bytes),
		parser.WithContext(context),
	)

	links, err := p.parseLinks(root, bytes)
	if err != nil {
		return nil, err
	}
	frontmatter, err := yaml.ParseFrontmatter(bytes)
	if err != nil {
		return nil, err
	}

	title, bodyStart, err := parseTitle(frontmatter, root, bytes)
	if err != nil {
		return nil, err
	}

	body := parseBody(bodyStart, bytes)

	tags, err := parseTags(root)
	if err != nil {
		return nil, err
	}

	combinedTags := append(frontmatter.Tags, tags...)

	return &core.NoteContent{
		Title:        title,
		Body:         body,
		Lead:         parseLead(body),
		Links:        links,
		Tags:         tagNames(combinedTags),
		Metadata:     frontmatter.Values,
		ExtendedTags: combinedTags,
	}, nil
}

// parseTitle extracts the note title with its node.
func parseTitle(frontmatter *FrontMatterExtended, root ast.Node, source []byte) (title opt.String, bodyStart int, err error) {
	titleString := frontmatter.GetTitleString()
	if !titleString.IsNull() {
		title = titleString
		bodyStart = frontmatter.End
		return
	}

	var titleNode *ast.Heading
	err = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*ast.Heading); ok && entering &&
			(titleNode == nil || heading.Level < titleNode.Level) {

			titleNode = heading
			if heading.Level == 1 {
				return ast.WalkStop, nil
			}
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return
	}

	if titleNode != nil {
		title = opt.NewNotEmptyString(string(titleNode.Text(source)))

		if lines := titleNode.Lines(); lines.Len() > 0 {
			bodyStart = lines.At(lines.Len() - 1).Stop
		}
	}
	return
}

// parseBody extracts the whole content after the title.
func parseBody(startIndex int, source []byte) opt.String {
	return opt.NewNotEmptyString(
		strings.TrimSpace(
			string(source[startIndex:]),
		),
	)
}

// parseLead extracts the body content until the first blank line.
func parseLead(body opt.String) opt.String {
	var lead strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(body.String()))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			break
		}
		lead.WriteString(scanner.Text() + "\n")
	}

	return opt.NewNotEmptyString(strings.TrimSpace(lead.String()))
}

func resolveTags(tags []string, pos int) []core.Tag {
	// if the tags slice has length larger than one, we assume these are colon separated tags,
	// that the parser has returned only the start position of the first of them, and we make position adjustments.
	result := make([]core.Tag, 0)
	currentPosition := 0
	for _, t := range tags {
		result = append(result, core.Tag{Name: t, Pos: pos + currentPosition})
		currentPosition += len(t) + 1
	}
	return result
}

func parseTags(root ast.Node) ([]core.Tag, error) {
	// Parse #hashtags and :colon:tags:
	tags := []core.Tag{}
	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if tagsNode, ok := n.(*extensions.Tags); ok && entering {
			tagNodePosition := tagsNode.Pos()
			tagNodeContent := tagsNode.Tags
			tags = append(tags, resolveTags(tagNodeContent, tagNodePosition)...)
		}
		return ast.WalkContinue, nil
	})
	return tags, err
}

// parseLinks extracts outbound links from the note.
func (p *Parser) parseLinks(root ast.Node, source []byte) ([]core.Link, error) {
	links := make([]core.Link, 0)

	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch link := n.(type) {
			case *ast.Link:
				href, err := url.PathUnescape(string(link.Destination))
				p.logger.Err(err)
				if href != "" {
					snippet, snStart, snEnd := extractLines(n, source)
					links = append(links, core.Link{
						Title:        string(link.Text(source)),
						Href:         href,
						Type:         core.LinkTypeMarkdown,
						Rels:         core.LinkRels(strings.Fields(string(link.Title))...),
						IsExternal:   strutil.IsURL(href),
						Snippet:      snippet,
						SnippetStart: snStart,
						SnippetEnd:   snEnd,
					})
				}

			case *ast.AutoLink:
				if href := string(link.URL(source)); href != "" && link.AutoLinkType == ast.AutoLinkURL {
					snippet, snStart, snEnd := extractLines(n, source)
					links = append(links, core.Link{
						Title:        string(link.Label(source)),
						Href:         href,
						Type:         core.LinkTypeImplicit,
						Rels:         []core.LinkRelation{},
						IsExternal:   true,
						Snippet:      snippet,
						SnippetStart: snStart,
						SnippetEnd:   snEnd,
					})
				}

			case *extensions.WikiLink:
				href := string(link.Destination)
				if href != "" {
					snippet, snStart, snEnd := extractLines(n, source)
					links = append(links, core.Link{
						Title:        string(link.Text(source)),
						Href:         href,
						Type:         core.LinkTypeWikiLink,
						Rels:         core.LinkRels(strings.Fields(string(link.Title))...),
						IsExternal:   strutil.IsURL(href),
						Snippet:      snippet,
						SnippetStart: snStart,
						SnippetEnd:   snEnd,
					})
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return links, err
}

func extractLines(n ast.Node, source []byte) (content string, start, end int) {
	if n == nil {
		return
	}
	switch n.Type() {
	case ast.TypeInline:
		return extractLines(n.Parent(), source)

	case ast.TypeBlock:
		segs := n.Lines()
		if segs.Len() == 0 {
			return
		}
		start = segs.At(0).Start
		end = segs.At(segs.Len() - 1).Stop
		content = string(source[start:end])
	}

	return
}
