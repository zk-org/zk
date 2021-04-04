package exec

import (
	"fmt"
	"os/exec"
	"syscall"
)

// CommandFromString returns a Cmd running the given command.
func CommandFromString(command string) *exec.Cmd {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CmdLine:       fmt.Sprintf(` /v:on/s/c "%s"`, command),
		CreationFlags: 0,
	}
	return cmd
}
