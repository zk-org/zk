package editor

import (
	"testing"

	"github.com/kballard/go-shellquote"
	"github.com/zk-org/zk/internal/util/test/assert"
)

func TestEditorBinary(t *testing.T) {
	tests := []struct {
		editor   string
		expected string
	}{
		{"nvim", "nvim"},
		{"/usr/bin/nvim", "nvim"},
		{"./nvim", "nvim"},
		{"code -w", "code"},
		{"Code.exe", "code"},
		{"NVIM", "nvim"},
		{"'my editor'", "my editor"},
		{"flatpak run com.visualstudio.code", "flatpak"},
		{"", ""},
		{`"unterminated`, ""},
	}

	for _, test := range tests {
		t.Run(test.editor, func(t *testing.T) {
			assert.Equal(t, editorBinary(test.editor), test.expected)
		})
	}
}

func TestOpenArgs(t *testing.T) {
	pos := Position{Line: 12, Column: 4}

	tests := []struct {
		binary   string
		expected []string
	}{
		{"vim", []string{"+12", "/notes/a.md"}},
		{"nvim", []string{"+12", "/notes/a.md"}},
		{"vi", []string{"+12", "/notes/a.md"}},
		{"gvim", []string{"+12", "/notes/a.md"}},
		{"nano", []string{"+12,4", "/notes/a.md"}},
		{"emacs", []string{"+12:4", "/notes/a.md"}},
		{"emacsclient", []string{"+12:4", "/notes/a.md"}},
		{"kak", []string{"+12:4", "/notes/a.md"}},
		{"micro", []string{"+12:4", "/notes/a.md"}},
		{"hx", []string{"/notes/a.md:12:4"}},
		{"helix", []string{"/notes/a.md:12:4"}},
		{"subl", []string{"/notes/a.md:12:4"}},
		{"code", []string{"--goto", "/notes/a.md:12:4"}},
		{"codium", []string{"--goto", "/notes/a.md:12:4"}},
		{"code-insiders", []string{"--goto", "/notes/a.md:12:4"}},
		// Unknown editors open the file without any position.
		{"ed", []string{"/notes/a.md"}},
		{"", []string{"/notes/a.md"}},
	}

	for _, test := range tests {
		t.Run(test.binary, func(t *testing.T) {
			assert.Equal(t, openArgs(test.binary, "/notes/a.md", pos), test.expected)
		})
	}
}

func TestOpenArgsWithoutPosition(t *testing.T) {
	for _, binary := range []string{"vim", "nano", "emacs", "hx", "code", "ed"} {
		t.Run(binary, func(t *testing.T) {
			assert.Equal(t, openArgs(binary, "/notes/a.md", Position{}), []string{"/notes/a.md"})
			// A column alone is not a position.
			assert.Equal(t, openArgs(binary, "/notes/a.md", Position{Column: 4}), []string{"/notes/a.md"})
		})
	}
}

func TestOpenArgsDefaultsColumn(t *testing.T) {
	pos := Position{Line: 12}

	assert.Equal(t, openArgs("vim", "/notes/a.md", pos), []string{"+12", "/notes/a.md"})
	assert.Equal(t, openArgs("nano", "/notes/a.md", pos), []string{"+12,1", "/notes/a.md"})
	assert.Equal(t, openArgs("emacs", "/notes/a.md", pos), []string{"+12:1", "/notes/a.md"})
	assert.Equal(t, openArgs("hx", "/notes/a.md", pos), []string{"/notes/a.md:12:1"})
	assert.Equal(t, openArgs("code", "/notes/a.md", pos), []string{"--goto", "/notes/a.md:12:1"})
}

func TestOpenArgsQuoting(t *testing.T) {
	pos := Position{Line: 12, Column: 4}

	assert.Equal(t,
		shellquote.Join(openArgs("code", "/my notes/a b.md", pos)...),
		"--goto '/my notes/a b.md:12:4'",
	)
	assert.Equal(t,
		shellquote.Join(openArgs("vim", "/my notes/a b.md", pos)...),
		"+12 '/my notes/a b.md'",
	)
}
