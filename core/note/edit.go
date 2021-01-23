package note

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/core/zk"
	"github.com/mickael-menu/zk/util/errors"
	"github.com/mickael-menu/zk/util/opt"
	osutil "github.com/mickael-menu/zk/util/os"
)

// Edit starts the editor with the notes at given paths.
func Edit(zk *zk.Zk, paths ...string) error {
	editor := editor(zk)
	if editor.IsNull() {
		return fmt.Errorf("no editor set in config")
	}

	wrap := errors.Wrapperf("failed to launch editor: %v", editor)

	args, err := shellquote.Split(editor.String())
	if err != nil {
		return wrap(err)
	}
	if len(args) == 0 {
		return wrap(fmt.Errorf("editor command is not valid: %v", editor))
	}
	args = append(args, paths...)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return wrap(cmd.Run())
}

// editor returns the editor command to use to edit a note.
func editor(zk *zk.Zk) opt.String {
	return zk.Config.Editor.
		Or(osutil.GetOptEnv("VISUAL")).
		Or(osutil.GetOptEnv("EDITOR"))
}
