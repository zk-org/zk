// +build !windows

package exec

import (
	"os"
	"os/exec"

	"github.com/kballard/go-shellquote"
)

// CommandFromString returns a Cmd running the given command with $SHELL.
func CommandFromString(command string, args ...string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}
	args = append([]string{"-c", command + " " + shellquote.Join(args...)})
	return exec.Command(shell, args...)
}
