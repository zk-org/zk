// +build !windows

package exec

import (
	"os/exec"

	osutil "github.com/zk-org/zk/internal/util/os"
)

// CommandFromString returns a Cmd running the given command with $SHELL.
func CommandFromString(command string, args ...string) *exec.Cmd {
	shell := osutil.GetOptEnv("ZK_SHELL").
		Or(osutil.GetOptEnv("SHELL")).
		OrString("sh").
		Unwrap()

	args = append([]string{"-c", command, "--"}, args...)
	return exec.Command(shell, args...)
}
