package markdown

import (
	"github.com/yuin/goldmark/ast"
)

// LinkPosition holds the source position of a link.
type LinkPosition struct {
	Start int // Byte offset of '['
	End   int // Byte offset after ')'
}

// GetLinkPosition calculates the source position of a link on-demand.
// Returns nil if position cannot be determined.
func GetLinkPosition(link *ast.Link, source []byte) *LinkPosition {
	// Use potentially cached position:
	if startVal, ok := link.AttributeString("linkStart"); ok {
		if endVal, ok := link.AttributeString("linkEnd"); ok {
			start, ok1 := startVal.(int)
			end, ok2 := endVal.(int)
			if ok1 && ok2 {
				return &LinkPosition{Start: start, End: end}
			}
		}
	}

	start, end := calculateLinkPositions(link, source)
	if start < 0 || end <= start {
		return nil
	}

	// Cache for future calls.
	link.SetAttributeString("linkStart", start)
	link.SetAttributeString("linkEnd", end)

	return &LinkPosition{Start: start, End: end}
}

// calculateLinkPositions finds the byte offsets for a Link node by scanning the source.
// Returns (start of '[', end after ')').
func calculateLinkPositions(link *ast.Link, source []byte) (int, int) {
	// ast.Link node doesn't store position information for delimiters. The
	// parser knows these positions internally (in linkLabelState.Segment),
	// but only transfers the URL and title to the final node. We need to
	// scan for the exact positions.

	textStart, textEnd := findFirstAndLastTextPositions(link)
	if textStart < 0 {
		return -1, -1
	}

	// Scan backwards from textStart to find '['.
	linkStart := -1
	for i := textStart - 1; i >= 0; i-- {
		if source[i] == '[' {
			linkStart = i
			break
		}
		if source[i] == ']' || source[i] == '(' {
			break
		}
	}

	linkEnd := -1
	i := textEnd

	if source[i] != ']' {
		// TODO: Should warn here? Is this even reachable?
		return linkStart, textEnd
	}
	i++ // skip ']'

	if i >= len(source) {
		// Link of the style [link].
		return linkStart, textEnd
	}

	if source[i] == '(' {
		// Inline link: scan to find matching parenthesis.
		i++ // skip '('
		parenDepth := 1
		for i < len(source) && parenDepth > 0 {
			// Ignore escaped characters (potentially, parenthesis).
			if source[i] == '\\' && i+1 < len(source) {
				i += 2
				continue
			}
			if source[i] == '(' {
				parenDepth++
			} else if source[i] == ')' {
				parenDepth--
				if parenDepth == 0 {
					linkEnd = i + 1
					break
				}
			}
			i++
		}
	} else if source[i] == '[' {
		// Reference link: [text][label]. Scan to find matching brace.
		i++ // skip '['
		for i < len(source) {
			// Ignore escaped characters (potentially, parenthesis).
			if source[i] == '\\' && i+1 < len(source) {
				i += 2
				continue
			}
			if source[i] == ']' {
				linkEnd = i + 1
				break
			}
			i++
		}
	} else {
		// Unexpected character after ']'. Include the ']' in the range.
		// TODO: I don't think this can happen.
		linkEnd = textEnd + 1
	}

	return linkStart, linkEnd
}

// findFirstAndLastTextPositions finds the first and last text positions within a node.
// Handles nested text nodes (e.g.: emphasis, bold, etc.).
func findFirstAndLastTextPositions(parent ast.Node) (int, int) {
	var firstPos, lastPos = -1, -1

	ast.Walk(parent, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if text, ok := n.(*ast.Text); ok {
			if firstPos < 0 {
				firstPos = text.Segment.Start
			}
			lastPos = text.Segment.Stop
		}
		return ast.WalkContinue, nil
	})

	return firstPos, lastPos
}
