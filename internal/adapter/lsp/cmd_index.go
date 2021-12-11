package lsp

import (
	"fmt"

	"github.com/mickael-menu/zk/internal/core"
)

const cmdIndex = "zk.index"

func executeCommandIndex(notebook *core.Notebook, args []interface{}) (interface{}, error) {
	opts := core.NoteIndexOpts{}
	if len(args) == 2 {
		options, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("zk.index expects a dictionary of options as second argument, got: %v", args[1])
		}
		if forceOption, ok := options["force"]; ok {
			opts.Force = toBool(forceOption)
		}
		if verboseOption, ok := options["verbose"]; ok {
			opts.Verbose = toBool(verboseOption)
		}
	}

	return notebook.Index(opts)
}
