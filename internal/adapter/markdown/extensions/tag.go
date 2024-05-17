package extensions

import (
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Tags represents a list of inline tags in a Markdown document.
type Tags struct {
	ast.BaseInline
	// Tags in this list.
	Tags []string
}

func (n *Tags) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Tags"] = strings.Join(n.Tags, ", ")
	ast.DumpHelper(n, source, level, m, nil)
}

// KindTags is a NodeKind of the Tags node.
var KindTags = ast.NewNodeKind("Tags")

func (n *Tags) Kind() ast.NodeKind {
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
	// Indicates whether #hashtags are parsed.
	HashtagEnabled bool
	// Indicates whether Bear's multi-word tags are parsed. Hashtags must be enabled as well.
	MultiWordTagEnabled bool
	// Indicates whether :colon:tags: are parsed.
	ColontagEnabled bool
}

func (t *TagExt) Extend(m goldmark.Markdown) {
	parsers := []util.PrioritizedValue{}

	if t.HashtagEnabled {
		parsers = append(parsers, util.Prioritized(&hashtagParser{
			multiWordTagEnabled: t.MultiWordTagEnabled,
		}, 2000))
	}

	if t.ColontagEnabled {
		parsers = append(parsers, util.Prioritized(&colontagParser{}, 2000))
	}

	if len(parsers) > 0 {
		m.Parser().AddOptions(parser.WithInlineParsers(parsers...))
	}
}

// hashtagParser parses #hashtags, including Bear's #multi words# tags
type hashtagParser struct {
	multiWordTagEnabled bool
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
		tag                   string // Accumulator for the hashtag
		multiWordTagCandidate string // Accumulator for a potential Bear multi-word tag
	)

	var (
		escaping            = false // Found a backslash, next character will be literal
		parsingMultiWordTag = false // Finished parsing a hashtag, now attempt parsing a Bear multi-word tag
		endPos              = 0     // Last position of the tag in the line
		multiWordTagEndPos  = 0     // Last position of the multi-word tag in the line
	)

	appendChar := func(c rune) {
		if parsingMultiWordTag {
			multiWordTagCandidate += string(c)
		} else {
			tag += string(c)
		}
	}

	for i, char := range string(line) {
		if i == 0 {
			// Skip the first character, as it is #
			continue
		}
		if parsingMultiWordTag {
			multiWordTagEndPos = i
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

		} else if parsingMultiWordTag {
			// Parsing a multi-word tag candidate.
			if isValidTagChar(char, '#') || unicode.IsSpace(char) {
				appendChar(char)
			} else if char == '#' {
				// A valid multi-word tag must not have a space before the closing #.
				if !unicode.IsSpace(previousChar) {
					tag = multiWordTagCandidate
					endPos = multiWordTagEndPos
				}
				break
			}
			previousChar = char

		} else if !p.multiWordTagEnabled && char == '#' {
			// A tag terminated with a # is invalid when not in a multi-word tag.
			return nil

		} else if p.multiWordTagEnabled && unicode.IsSpace(char) {
			// Found a space, let's try to parse a multi-word tag.
			previousChar = char
			multiWordTagCandidate = tag
			parsingMultiWordTag = true
			appendChar(char)

		} else if !isValidTagChar(char, '#') {
			// Found an invalid character, the hashtag is complete.
			break

		} else {
			appendChar(char)
		}
	}

	tag = strings.TrimSpace(tag)
	if len(tag) == 0 || !isValidHashTag(tag) {
		return nil
	}

	block.Advance(endPos)

	return &Tags{
		BaseInline: ast.BaseInline{},
		Tags:       []string{tag},
	}
}

func isValidHashTag(tag string) bool {
	for _, char := range tag {
		if !unicode.IsNumber(char) {
			return true
		}
	}
	return false
}

// colontagParser parses :colon:separated:tags:.
type colontagParser struct{}

func (p *colontagParser) Trigger() []byte {
	return []byte{':'}
}

func (p *colontagParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	previousChar := block.PrecendingCharacter()
	line, _ := block.PeekLine()

	// A colontag can't be directly preceded by a : or any other valid character.
	if isValidTagChar(previousChar, '\x00') {
		return nil
	}

	var (
		tag  string       // Accumulator for the current colontag
		tags = []string{} // All colontags found
	)

	var (
		escaping = false // Found a backslash, next character will be literal
		endPos   = 0     // Last position of the colontags in the line
	)

	appendChar := func(c rune) {
		tag += string(c)
	}

	for i, char := range string(line[1:]) {
		endPos = i

		if escaping {
			// Currently escaping? The character will be appended literally.
			appendChar(char)
			escaping = false

		} else if char == '\\' {
			// Found a backslash, next character will be escaped.
			escaping = true

		} else if char == ':' {
			tag = strings.TrimSpace(tag)
			if !isValidTag(tag) {
				break
			}
			tags = append(tags, tag)
			tag = ""

		} else if !isValidTagChar(char, ':') {
			// Found an invalid character, the colontag is complete.
			break

		} else {
			appendChar(char)
		}
	}

	if len(tags) == 0 {
		return nil
	}

	block.Advance(endPos)

	return &Tags{
		BaseInline: ast.BaseInline{},
		Tags:       tags,
	}
}

func isValidTagChar(r rune, excluded rune) bool {
	return r != excluded && (unicode.IsLetter(r) || unicode.IsNumber(r) ||
		r == '/' || r == '@' || r == '\'' || r == '~' ||
		r == '-' || r == '_' || r == '$' || r == '%' ||
		r == '&' || r == '+' || r == '=' || r == ':' ||
		r == '#')
}

func isValidTag(tag string) bool {
	if len(tag) == 0 {
		return false
	}

	// Prevent Markdown table syntax to be parsed a a colon tag, e.g. |:---:|
	// https://github.com/zk-org/zk/issues/185
	for _, c := range tag {
		if c != '-' {
			return true
		}
	}

	return false
}
