package lsp

import (
	"fmt"

	"github.com/mickael-menu/zk/internal/core"
)

const cmdIndex = "zk.index"

func executeCommandIndex(notebooks *core.NotebookStore, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("zk.index expects a notebook path as first argument")
	}
	path, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("zk.index expects a notebook path as first argument, got: %v", args[0])
	}

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

	notebook, err := notebooks.Open(path)
	if err != nil {
		return nil, err
	}

	return notebook.Index(opts)
}
