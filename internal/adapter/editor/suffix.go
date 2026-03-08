//go:build !windows

package editor

import "os"

// cmdSuffix returns the shell suffix redirecting the editor's standard input to
// the controlling terminal.
//
// There isn't always one to redirect to: when zk is started by a desktop URI
// handler there is no terminal attached, and the redirection would prevent the
// editor from starting at all.
func cmdSuffix() string {
	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		return ""
	}
	tty.Close()
	return " </dev/tty"
}
