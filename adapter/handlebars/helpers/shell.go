package helpers

import (
	"os/exec"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/kballard/go-shellquote"
	"github.com/mickael-menu/zk/util"
)

// RegisterShell registers the {{sh}} template helper, which runs shell commands.
//
// {{#sh "tr '[a-z]' '[A-Z]'"}}Hello, world!{{/sh}} -> HELLO, WORLD!
// {{sh "echo 'Hello, world!'"}} -> Hello, world!
func RegisterShell(logger util.Logger) {
	raymond.RegisterHelper("sh", func(arg string, options *raymond.Options) string {
		args, err := shellquote.Split(arg)
		if err != nil {
			logger.Printf("{{sh}} failed to parse command: %v: %v", arg, err)
			return ""
		}
		if len(args) == 0 {
			logger.Printf("{{sh}} expects a valid shell command, received: %v", arg)
			return ""
		}

		cmd := exec.Command(args[0], args[1:]...)
		// Feed any block content as piped input
		cmd.Stdin = strings.NewReader(options.Fn())

		output, err := cmd.Output()
		if err != nil {
			logger.Printf("{{sh}} command failed: %v", err)
			return ""
		}

		return strings.TrimSpace(string(output))
	})
}
