package exec

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// CommandFromString returns a Cmd running the given command.
func CommandFromString(command string, args ...string) *exec.Cmd {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CmdLine:       fmt.Sprintf(` /v:on/s/c "%s %s"`, command, strings.Join(args[:], " ")),
		CreationFlags: 0,
	}
	return cmd
}
