package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
	executil "github.com/zk-org/zk/internal/util/exec"
	osutil "github.com/zk-org/zk/internal/util/os"
	"github.com/zk-org/zk/internal/util/ptr"
)

// Editor represents an external editor able to edit the notes.
type Editor struct {
	editor string
	shell  string
}

// NewEditor creates a new Editor from the given editor user setting or the
// matching environment variables.
func NewEditor(editor *string, shell string) (*Editor, error) {
	editor = ptr.First(
		osutil.GetOptEnv("ZK_EDITOR"),
		editor,
		osutil.GetOptEnv("VISUAL"),
		osutil.GetOptEnv("EDITOR"),
	)

	if editor == nil {
		return nil, fmt.Errorf("no editor set in config")
	}

	return &Editor{editor: *editor, shell: shell}, nil
}

// Open launches the editor with the notes at given paths.
func (e *Editor) Open(paths ...string) error {
	// If stdin is a pipe, pipe it to the editor, so it can consume the output from
	// that pipe. Allows providing input to commands like `zk new` when using Vim.
	//
	// Don't redirect a tty, since that messes up rendering of some TUI editors.
	//
	// See: https://github.com/zk-org/zk/issues/4
	// See: https://github.com/zk-org/zk/pull/693
	suffix := CMD_SUFFIX
	if isatty.IsTerminal(os.Stdin.Fd()) {
		suffix = ""
	}
	cmd := executil.CommandFromString(e.shell, e.editor+" "+shellquote.Join(paths...)+suffix)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}
	switch err.(type) {
	case *exec.ExitError:
		return fmt.Errorf("operation aborted by editor: %s %s: %w", e.editor, strings.Join(paths, " "), err)
	default:
		return fmt.Errorf("failed to launch editor: %s %s: %w", e.editor, strings.Join(paths, " "), err)
	}
}
