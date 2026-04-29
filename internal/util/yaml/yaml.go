package yaml

import (
	"bytes"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/opt"
	yml "gopkg.in/yaml.v3"
)

type FrontmatterExtended struct {
	Root   *yml.Node
	Values map[string]any
	Bytes  []byte
	Start  int
	End    int
	Tags   []core.Tag
}

func walkScalar(node *yml.Node) (any, error) {
	return node.Value, nil
}

func walkSequence(node *yml.Node) ([]any, error) {
	out := make([]any, 0, len(node.Content))
	for _, child := range node.Content {
		v, err := walkNode(child)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func walkMapping(node *yml.Node) (map[string]any, error) {
	nodeContent := node.Content
	nodeContentLength := len(nodeContent)
	nodeContentKeysLength := nodeContentLength / 2
	out := make(map[string]any, nodeContentKeysLength)
	for i := 0; i < nodeContentLength; i += 2 {
		v, err := walkNode(nodeContent[i+1])
		if err != nil {
			return nil, err
		}
		out[strings.ToLower(nodeContent[i].Value)] = v
	}
	return out, nil
}

func walkNode(node *yml.Node) (any, error) {
	switch node.Kind {
	case yml.ScalarNode:
		return walkScalar(node)
	case yml.SequenceNode:
		return walkSequence(node)
	case yml.MappingNode:
		return walkMapping(node)
	case yml.AliasNode:
		return walkNode(node.Alias)
	}
	panic("Unreachable: cases above are exhaustive.")
}

func byteOffset(src []byte, line int, col int) int {
	l := 1
	for i := 0; i < len(src); {
		if l == line {
			for c := 1; c < col; c++ {
				_, size := utf8.DecodeRune(src[i:])
				i += size
			}
			return i
		}
		if src[i] == '\n' {
			l++
		}
		i++
	}
	return -1
}

func trimTag(tag string) string {
	trimmedTag := strings.TrimPrefix(tag, `"`)
	trimmedTag = strings.TrimSuffix(trimmedTag, `"`)
	trimmedTag = strings.TrimSpace(trimmedTag)
	trimmedTag = strings.TrimPrefix(trimmedTag, "#")
	return trimmedTag
}

func collectTagsInternal(keyNode *yml.Node, valNode *yml.Node, frontMatterBytes []byte) ([]core.Tag, error) {
	key := keyNode.Value
	parsedTags := []core.Tag{}
	for _, keyword := range []string{"tag", "tags", "keyword", "keywords"} {
		if strings.Compare(strings.ToLower(key), keyword) == 0 {
			switch valNode.Kind {
			case yml.ScalarNode:
				line := valNode.Line
				column := valNode.Column
				tagString := valNode.Value
				tagStringPosition := byteOffset(frontMatterBytes, line, column)
				if len(tagString) > 0 {
					currentOffset := tagStringPosition
					tags := strings.Fields(tagString)
					for i := range tags {
						tag := tags[i]
						tagOffset := bytes.Index(frontMatterBytes[currentOffset:], []byte(tag))
						tagOffset += currentOffset
						trimmedTag := trimTag(tag)
						if len(trimmedTag) > 0 {
							parsedTags = append(parsedTags, core.Tag{Name: trimmedTag, Pos: tagOffset})
						}
						currentOffset = tagOffset + len(tag)
					}
				}
			case yml.SequenceNode:
				valNodeContent := valNode.Content
				for _, seqItemNode := range valNodeContent {
					if !(seqItemNode.Kind == yml.ScalarNode) {
						continue
					}
					tag := seqItemNode.Value
					line := seqItemNode.Line
					column := seqItemNode.Column
					trimmedTag := trimTag(tag)
					tagPosition := byteOffset(frontMatterBytes, line, column)
					parsedTags = append(parsedTags, core.Tag{Name: trimmedTag, Pos: tagPosition})
				}
			}
			break
		}
	}
	return parsedTags, nil
}

func collectTags(rootMapping *yml.Node, frontmatter *FrontmatterExtended) ([]core.Tag, error) {
	rootMappingContent := rootMapping.Content
	rootMappingContentLength := len(rootMappingContent)
	parsedTags := []core.Tag{}
	frontMatterBytes := frontmatter.Bytes
	for i := 0; i < rootMappingContentLength; i += 2 {
		res := []core.Tag{}
		keyNode := rootMappingContent[i]
		valNode := rootMappingContent[i+1]
		res, err := collectTagsInternal(keyNode, valNode, frontMatterBytes)
		if err != nil {
			return nil, err
		}
		parsedTags = append(parsedTags, res...)
	}
	return parsedTags, nil
}

var frontmatterRegex = regexp.MustCompile(`(?ms)\A\s*-+\s*$.*?^\s*-+\s*$`)

func ParseFrontmatter(source []byte) (*FrontmatterExtended, error) {
	index := frontmatterRegex.FindIndex(source)
	frontmatter := FrontmatterExtended{Values: map[string]any{}}

	if index == nil {
		return &frontmatter, nil
	}

	var rootNode yml.Node

	frontMatterStart := index[0]
	frontMatterEnd := index[1]
	frontmatter.Start = frontMatterStart
	frontmatter.End = frontMatterEnd

	frontMatterBytes := source[frontMatterStart:frontMatterEnd]
	err := yml.Unmarshal(frontMatterBytes, &rootNode)

	if err != nil {
		return nil, err
	}

	frontmatter.Root = &rootNode
	frontmatter.Bytes = frontMatterBytes

	if len(rootNode.Content) <= 0 {
		return &frontmatter, nil
	}

	mapping := rootNode.Content[0]
	if mapping.Kind != yml.MappingNode {
		return &frontmatter, nil
	}

	frontmatter.Values, err = walkMapping(mapping)

	if err != nil {
		return nil, err
	}

	tags, err := collectTags(mapping, &frontmatter)

	if err != nil {
		return nil, err
	}

	frontmatter.Tags = tags

	return &frontmatter, nil
}

func (frontmatter *FrontmatterExtended) GetValue(key string) (any, bool) {
	if frontmatter.Values == nil {
		return nil, false
	}
	value, ok := frontmatter.Values[key]
	return value, ok
}

func (frontmatter *FrontmatterExtended) GetTitleString() opt.String {
	titleValue, found := frontmatter.GetValue("title")
	if found {
		if titleStr, ok := titleValue.(string); ok {
			return opt.NewNotEmptyString(titleStr)
		}
	}
	return opt.NullString
}
