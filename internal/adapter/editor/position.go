package editor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kballard/go-shellquote"
)

// Position is a 1-based line and column location in a file. A Line of 0 means
// that no particular position is requested.
type Position struct {
	Line   int
	Column int
}

// IsValid returns whether this position points at an actual location.
func (p Position) IsValid() bool {
	return p.Line > 0
}

// positionSyntax is the command-line convention an editor uses to open a file
// at a given location.
type positionSyntax int

const (
	// The editor has no known syntax, so the position is dropped.
	posNone positionSyntax = iota
	// +<line> <path>
	posPlusLine
	// +<line>:<column> <path>
	posPlusLineColon
	// +<line>,<column> <path>
	posPlusLineComma
	// <path>:<line>:<column>
	posSuffixColon
	// --goto <path>:<line>:<column>
	posGoto
)

// editorPositionSyntaxes maps the name of an editor binary to the syntax it
// uses to open a file at a given line and column.
var editorPositionSyntaxes = map[string]positionSyntax{
	"vi":   posPlusLine,
	"view": posPlusLine,
	"vim":  posPlusLine,
	"nvim": posPlusLine,
	"gvim": posPlusLine,
	"mvim": posPlusLine,

	"nano": posPlusLineComma,

	"emacs":       posPlusLineColon,
	"emacsclient": posPlusLineColon,
	"kak":         posPlusLineColon,
	"micro":       posPlusLineColon,

	"hx":           posSuffixColon,
	"helix":        posSuffixColon,
	"subl":         posSuffixColon,
	"sublime_text": posSuffixColon,

	"code":          posGoto,
	"code-insiders": posGoto,
	"code-oss":      posGoto,
	"codium":        posGoto,
	"vscodium":      posGoto,
	"cursor":        posGoto,
	"windsurf":      posGoto,
}

// editorBinary extracts the name of the editor binary from an editor command,
// e.g. "nvim" for `/usr/bin/nvim -p`.
//
// Returns an empty string when the command can't be parsed, which falls back to
// opening files without any position.
func editorBinary(editor string) string {
	words, err := shellquote.Split(editor)
	if err != nil || len(words) == 0 {
		return ""
	}
	return strings.TrimSuffix(strings.ToLower(filepath.Base(words[0])), ".exe")
}

// openArgs returns the arguments to pass to the given editor binary to open
// the file at path, at the given position.
//
// Editors without a known position syntax receive the bare path, as there's no
// portable way to express a location.
func openArgs(binary string, path string, pos Position) []string {
	if !pos.IsValid() {
		return []string{path}
	}

	column := pos.Column
	if column < 1 {
		column = 1
	}

	switch editorPositionSyntaxes[binary] {
	case posPlusLine:
		return []string{fmt.Sprintf("+%d", pos.Line), path}
	case posPlusLineColon:
		return []string{fmt.Sprintf("+%d:%d", pos.Line, column), path}
	case posPlusLineComma:
		return []string{fmt.Sprintf("+%d,%d", pos.Line, column), path}
	case posSuffixColon:
		return []string{fmt.Sprintf("%s:%d:%d", path, pos.Line, column)}
	case posGoto:
		return []string{"--goto", fmt.Sprintf("%s:%d:%d", path, pos.Line, column)}
	default:
		return []string{path}
	}
}
