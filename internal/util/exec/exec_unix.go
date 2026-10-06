//go:build !windows

package exec

import (
	"os/exec"

	osutil "github.com/zk-org/zk/internal/util/os"
	"github.com/zk-org/zk/internal/util/ptr"
)

// ResolveShell returns the shell to use for running commands, checking in order:
// ZK_SHELL environment variable, config/tool.shell, SHELL environment variable, or "sh" as fallback.
func ResolveShell(configShell *string) string {
	if shell := ptr.First(
		osutil.GetOptEnv("ZK_SHELL"),
		configShell,
		osutil.GetOptEnv("SHELL"),
	); shell != nil {
		return *shell
	}
	return "sh"
}

// CommandFromString returns a Cmd running the given command with the specified shell.
func CommandFromString(shell, command string, args ...string) *exec.Cmd {
	args = append([]string{"-c", command, "--"}, args...)
	return exec.Command(shell, args...)
}
