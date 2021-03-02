package extensions

import (
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Tags represents a list of inline tags in a Markdown document.
type Tags struct {
	gast.BaseInline
	// Tags in this list.
	Tags []string
}

func (n *Tags) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Tags"] = strings.Join(n.Tags, ", ")
	gast.DumpHelper(n, source, level, m, nil)
}

// KindTags is a NodeKind of the Tags node.
var KindTags = gast.NewNodeKind("Tags")

func (n *Tags) Kind() gast.NodeKind {
	return KindTags
}

// TagExt is an extension parsing various flavors of tags.
//
// * #hashtags, including Bear's #multi words# tags
// * :colon:separated:tags:`, e.g. vimwiki and Org mode
//
// Are authorized in a tag:
// * unicode categories [L]etter and [N]umber
// * / @ ' ~ - _ $ % & + = and when possible # :
// * any character escaped with \, including whitespace
var TagExt = &tagExt{}

type tagExt struct{}

func (t *tagExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&hashtagParser{}, 2000),
			util.Prioritized(&colontagParser{}, 2000),
		),
	)
}

// hashtagParser parses #hashtags, including Bear's #multi words# tags
type hashtagParser struct{}

func (p *hashtagParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *hashtagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	var (
		tag string
	)

	var (
		escaping = false // Found a backslash, next character will be literal
		endPos   = 0     // Last position of the link in the line
	)

	for i, char := range string(line[1:]) {
		endPos = i

		if char == '\\' {
			escaping = true
			continue
		}

		if !escaping && !isValidTagChar(char, '#') {
			break
		}
		tag += string(char)

		escaping = false
	}

	if len(tag) == 0 || !isValidHashTag(tag) {
		return nil
	}

	block.Advance(endPos)

	return &Tags{
		BaseInline: gast.BaseInline{},
		Tags:       []string{tag},
	}
}

// colontagParser parses :colon:separated:tags:.
type colontagParser struct{}

func (p *colontagParser) Trigger() []byte {
	return []byte{':'}
}

func (p *colontagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	return nil
}

func isValidTagChar(r rune, excluded rune) bool {
	return r != excluded && (unicode.IsLetter(r) || unicode.IsNumber(r) ||
		r == '/' || r == '@' || r == '\'' || r == '~' ||
		r == '-' || r == '_' || r == '$' || r == '%' ||
		r == '&' || r == '+' || r == '=' || r == ':' ||
		r == '#')
}

func isValidHashTag(tag string) bool {
	for _, char := range tag {
		if !unicode.IsNumber(char) {
			return true
		}
	}
	return false
}
