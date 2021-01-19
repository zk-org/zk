// +build !windows

package exec

import (
	"os"
	"os/exec"
)

// CommandFromString returns a Cmd running the given command with $SHELL.
func CommandFromString(command string) *exec.Cmd {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}
	return exec.Command(shell, "-c", command)
}
