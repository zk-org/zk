package markdown

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/mickael-menu/zk/adapter/markdown/extensions"
	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util/opt"
	strutil "github.com/mickael-menu/zk/util/strings"
	"github.com/mvdan/xurls"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser parses the content of Markdown notes.
type Parser struct {
	md goldmark.Markdown
}

type ParserOpts struct {
	// Indicates whether #hashtags are parsed.
	HashtagEnabled bool
	// Indicates whether Bear's multi-word tags are parsed. Hashtags must be enabled as well.
	MultiWordTagEnabled bool
	// Indicates whether :colon:tags: are parsed.
	ColontagEnabled bool
}

// NewParser creates a new Markdown Parser.
func NewParser(options ParserOpts) *Parser {
	return &Parser{
		md: goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
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
	}
}

// Parse implements note.Parse.
func (p *Parser) Parse(source string) (*note.Content, error) {
	bytes := []byte(source)

	context := parser.NewContext()
	root := p.md.Parser().Parse(
		text.NewReader(bytes),
		parser.WithContext(context),
	)

	links, err := parseLinks(root, bytes)
	if err != nil {
		return nil, err
	}

	frontmatter, err := parseFrontmatter(context, bytes)
	if err != nil {
		return nil, err
	}

	title, bodyStart, err := parseTitle(frontmatter, root, bytes)
	if err != nil {
		return nil, err
	}
	body := parseBody(bodyStart, bytes)

	tags, err := parseTags(frontmatter, root, bytes)
	if err != nil {
		return nil, err
	}

	return &note.Content{
		Title: title,
		Body:  body,
		Lead:  parseLead(body),
		Links: links,
		Tags:  tags,
	}, nil
}

// parseTitle extracts the note title with its node.
func parseTitle(frontmatter frontmatter, root ast.Node, source []byte) (title opt.String, bodyStart int, err error) {
	if title = frontmatter.getString("title", "Title"); !title.IsNull() {
		bodyStart = frontmatter.end
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
	lead := ""
	scanner := bufio.NewScanner(strings.NewReader(body.String()))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			break
		}
		lead += scanner.Text() + "\n"
	}

	return opt.NewNotEmptyString(strings.TrimSpace(lead))
}

// parseTags extracts tags as #hashtags, :colon:tags: or from the YAML frontmatter.
func parseTags(frontmatter frontmatter, root ast.Node, source []byte) ([]string, error) {
	tags := make([]string, 0)

	// Parse from YAML frontmatter, either:
	// * a list of strings
	// * a single space-separated string
	findFMTags := func(key string) []string {
		if tags, ok := frontmatter.getStrings(key); ok {
			return tags
		} else if tags := frontmatter.getString(key); !tags.IsNull() {
			return strings.Fields(tags.Unwrap())
		} else {
			return []string{}
		}
	}

	for _, key := range []string{"tag", "tags", "keyword", "keywords"} {
		for _, t := range findFMTags(key) {
			// Trims any # prefix to support hashtags embedded in YAML
			// frontmatter, as in Simple Markdown Zettelkasten:
			// http://evantravers.com/articles/2020/11/23/zettelkasten-updates/
			tags = append(tags, strings.TrimPrefix(t, "#"))
		}
	}

	// Parse #hashtags and :colon:tags:
	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if tagsNode, ok := n.(*extensions.Tags); ok && entering {
			for _, tag := range tagsNode.Tags {
				tags = append(tags, tag)
			}
		}
		return ast.WalkContinue, nil
	})

	return strutil.RemoveDuplicates(tags), err
}

// parseLinks extracts outbound links from the note.
func parseLinks(root ast.Node, source []byte) ([]note.Link, error) {
	links := make([]note.Link, 0)

	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch link := n.(type) {
			case *ast.Link:
				href := string(link.Destination)
				if href != "" {
					links = append(links, note.Link{
						Title:    string(link.Text(source)),
						Href:     href,
						Rels:     strings.Fields(string(link.Title)),
						External: strutil.IsURL(href),
						Snippet:  extractLines(n.Parent(), source),
					})
				}

			case *ast.AutoLink:
				if href := string(link.URL(source)); href != "" && link.AutoLinkType == ast.AutoLinkURL {
					links = append(links, note.Link{
						Title:    string(link.Label(source)),
						Href:     href,
						Rels:     []string{},
						External: true,
						Snippet:  extractLines(n.Parent(), source),
					})
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return links, err
}

func extractLines(n ast.Node, source []byte) string {
	if n == nil {
		return ""
	}
	segs := n.Lines()
	if segs.Len() == 0 {
		return ""
	}
	start := segs.At(0).Start
	end := segs.At(segs.Len() - 1).Stop
	return string(source[start:end])
}

// frontmatter contains metadata parsed from a YAML frontmatter.
type frontmatter struct {
	values map[string]interface{}
	start  int
	end    int
}

var frontmatterRegex = regexp.MustCompile(`(?ms)^\s*-+\s*$.*?^\s*-+\s*$`)

func parseFrontmatter(context parser.Context, source []byte) (frontmatter, error) {
	var front frontmatter

	index := frontmatterRegex.FindIndex(source)
	if index == nil {
		return front, nil
	}

	front.start = index[0]
	front.end = index[1]
	front.values = map[string]interface{}{}

	values, err := meta.TryGet(context)
	if err != nil {
		return front, err
	}
	// Convert keys to lowercase, because we don't want to be case sensitive.
	for k, v := range values {
		front.values[strings.ToLower(k)] = v
	}

	return front, nil
}

// getString returns the first string value found for any of the given keys.
func (m frontmatter) getString(keys ...string) opt.String {
	if m.values == nil {
		return opt.NullString
	}

	for _, key := range keys {
		key = strings.ToLower(key)
		if val, ok := m.values[key]; ok {
			if val, ok := val.(string); ok {
				return opt.NewNotEmptyString(val)
			}
		}
	}
	return opt.NullString
}

// getStrings returns the first string list found for any of the given keys.
func (m frontmatter) getStrings(keys ...string) ([]string, bool) {
	if m.values == nil {
		return nil, false
	}

	for _, key := range keys {
		key = strings.ToLower(key)
		if val, ok := m.values[key]; ok {
			if val, ok := val.([]interface{}); ok {
				strings := []string{}
				for _, v := range val {
					strings = append(strings, fmt.Sprint(v))
				}
				return strings, true
			}
		}
	}
	return nil, false
}
