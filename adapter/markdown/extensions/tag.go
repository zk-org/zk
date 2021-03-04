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
type TagExt struct {
	// Indicates whether Bear's word tags should be parsed.
	WordTagEnabled bool
}

func (t *TagExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&hashtagParser{
				wordTagEnabled: t.WordTagEnabled,
			}, 2000),
			util.Prioritized(&colontagParser{}, 2000),
		),
	)
}

// hashtagParser parses #hashtags, including Bear's #multi words# tags
type hashtagParser struct {
	wordTagEnabled bool
}

func (p *hashtagParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *hashtagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	previousChar := block.PrecendingCharacter()
	line, _ := block.PeekLine()

	// A hashtag can't be directly preceded by a # or any other valid character.
	if isValidTagChar(previousChar, '\x00') {
		return nil
	}

	var (
		tag              string // Accumulator for the hashtag
		wordTagCandidate string // Accumulator for a potential Bear word tag
	)

	var (
		escaping       = false // Found a backslash, next character will be literal
		parsingWordTag = false // Finished parsing a hashtag, now attempt parsing a Bear word tag
		endPos         = 0     // Last position of the tag in the line
		wordTagEndPos  = 0     // Last position of the word tag in the line
	)

	appendChar := func(c rune) {
		if parsingWordTag {
			wordTagCandidate += string(c)
		} else {
			tag += string(c)
		}
	}

	for i, char := range string(line[1:]) {
		if parsingWordTag {
			wordTagEndPos = i
		} else {
			endPos = i
		}

		if escaping {
			// Currently escaping? The character will be appended literally.
			appendChar(char)
			escaping = false

		} else if char == '\\' {
			// Found a backslash, next character will be escaped.
			escaping = true

		} else if parsingWordTag {
			// Parsing a word tag candidate.
			if isValidTagChar(char, '#') || unicode.IsSpace(char) {
				appendChar(char)
			} else if char == '#' {
				// A valid word tag must not have a space before the closing #.
				if !unicode.IsSpace(previousChar) {
					tag = wordTagCandidate
					endPos = wordTagEndPos
				}
				break
			}
			previousChar = char

		} else if !p.wordTagEnabled && char == '#' {
			// A tag terminated with a # is invalid when not in a word tag.
			return nil

		} else if p.wordTagEnabled && unicode.IsSpace(char) {
			// Found a space, let's try to parse a word tag.
			previousChar = char
			wordTagCandidate = tag
			parsingWordTag = true
			appendChar(char)

		} else if !isValidTagChar(char, '#') {
			// Found an invalid character, the hashtag is complete.
			break

		} else {
			appendChar(char)
		}
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
