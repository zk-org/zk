package note

import (
	"fmt"
	"os"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/internal/core/zk"
	"github.com/mickael-menu/zk/internal/util/errors"
	executil "github.com/mickael-menu/zk/internal/util/exec"
	"github.com/mickael-menu/zk/internal/util/opt"
	osutil "github.com/mickael-menu/zk/internal/util/os"
)

// Edit starts the editor with the notes at given paths.
func Edit(zk *zk.Zk, paths ...string) error {
	editor := editor(zk)
	if editor.IsNull() {
		return fmt.Errorf("no editor set in config")
	}

	// /dev/tty is restored as stdin, in case the user used a pipe to feed
	// initial note content to `zk new`. Without this, Vim doesn't work
	// properly in this case.
	// See https://github.com/mickael-menu/zk/issues/4
	cmd := executil.CommandFromString(editor.String() + " " + shellquote.Join(paths...) + " </dev/tty")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.Wrapf(cmd.Run(), "failed to launch editor: %s %s", editor, strings.Join(paths, " "))
}

// editor returns the editor command to use to edit a note.
func editor(zk *zk.Zk) opt.String {
	return osutil.GetOptEnv("ZK_EDITOR").
		Or(zk.Config.Tool.Editor).
		Or(osutil.GetOptEnv("VISUAL")).
		Or(osutil.GetOptEnv("EDITOR"))
}
