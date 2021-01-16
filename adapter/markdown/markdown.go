package markdown

import (
	"bufio"
	"strings"

	"github.com/mickael-menu/zk/core/note"
	"github.com/mickael-menu/zk/util/opt"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Parser parses the content of Markdown notes.
type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new Markdown Parser.
func NewParser() *Parser {
	return &Parser{
		md: goldmark.New(),
	}
}

// Parse implements note.Parse.
func (p *Parser) Parse(source string) (note.Content, error) {
	out := note.Content{}

	bytes := []byte(source)
	root := p.md.Parser().Parse(text.NewReader(bytes))

	title, titleNode, err := parseTitle(root, bytes)
	if err != nil {
		return out, err
	}

	out.Title = title
	out.Body = parseBody(titleNode, bytes)
	out.Lead = parseLead(out.Body)

	return out, nil
}

// parseTitle extracts the note title with its node.
func parseTitle(root ast.Node, source []byte) (title opt.String, node ast.Node, err error) {
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
		node = titleNode
		title = opt.NewNotEmptyString(string(titleNode.Text(source)))
	}
	return
}

// parseBody extracts the whole content after the title.
func parseBody(titleNode ast.Node, source []byte) opt.String {
	start := 0
	if titleNode != nil {
		if lines := titleNode.Lines(); lines.Len() > 0 {
			start = lines.At(lines.Len() - 1).Stop
		}
	}

	return opt.NewNotEmptyString(
		strings.TrimSpace(
			string(source[start:]),
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
