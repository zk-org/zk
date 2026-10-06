package exec

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// ResolveShell returns the shell to use. On Windows, this always returns "cmd"
// as the shell configuration is not applicable.
func ResolveShell(configShell *string) string {
	return "cmd"
}

// CommandFromString returns a Cmd running the given command.
// The shell parameter is ignored on Windows as it uses cmd directly.
func CommandFromString(shell, command string, args ...string) *exec.Cmd {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CmdLine:       fmt.Sprintf(` /v:on/s/c "%s %s"`, command, strings.Join(args[:], " ")),
		CreationFlags: 0,
	}
	return cmd
}
