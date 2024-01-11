package helpers

import (
	"strings"

	"github.com/aymerick/raymond"
	"github.com/zk-org/zk/internal/util"
	"github.com/zk-org/zk/internal/util/exec"
)

// RegisterShell registers the {{sh}} template helper, which runs shell commands.
//
// {{#sh "tr '[a-z]' '[A-Z]'"}}Hello, world!{{/sh}} -> HELLO, WORLD!
// {{sh "echo 'Hello, world!'"}} -> Hello, world!
func RegisterShell(logger util.Logger) {
	raymond.RegisterHelper("sh", func(arg string, options *raymond.Options) string {
		cmd := exec.CommandFromString(arg)

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
