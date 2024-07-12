package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/zk-org/zk/internal/util/errors"
	executil "github.com/zk-org/zk/internal/util/exec"
	"github.com/zk-org/zk/internal/util/opt"
	osutil "github.com/zk-org/zk/internal/util/os"
)

// Editor represents an external editor able to edit the notes.
type Editor struct {
	editor string
}

// NewEditor creates a new Editor from the given editor user setting or the
// matching environment variables.
func NewEditor(editor opt.String) (*Editor, error) {
	editor = osutil.GetOptEnv("ZK_EDITOR").
		Or(editor).
		Or(osutil.GetOptEnv("VISUAL")).
		Or(osutil.GetOptEnv("EDITOR"))

	if editor.IsNull() {
		return nil, fmt.Errorf("no editor set in config")
	}

	return &Editor{editor.Unwrap()}, nil
}

// Open launches the editor with the notes at given paths.
func (e *Editor) Open(paths ...string) error {
	// /dev/tty is restored as stdin, in case the user used a pipe to feed
	// initial note content to `zk new`. Without this, Vim doesn't work
	// properly in this case.
	// See https://github.com/zk-org/zk/issues/4
	cmd := executil.CommandFromString(e.editor + " " + shellquote.Join(paths...) + " </dev/tty")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	switch err.(type) {
	case *exec.ExitError:
		return errors.Wrapf(err, "operation aborted by editor: %s %s", e.editor, strings.Join(paths, " "))
	default:
		return errors.Wrapf(err, "failed to launch editor: %s %s", e.editor, strings.Join(paths, " "))

	}
}
