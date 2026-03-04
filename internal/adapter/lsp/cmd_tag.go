package lsp

import (
	"fmt"

	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
)

const cmdTagList = "zk.tag.list"

type cmdTagListOpts struct {
	Sort []string `json:"sort"`
}

func executeCommandTagList(logger util.Logger, notebook *core.Notebook, args []any) (any, error) {
	var opts cmdTagListOpts
	if len(args) > 1 {
		arg, ok := args[1].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s expects a dictionary of options as second argument, got: %v", cmdTagList, args[1])
		}
		err := unmarshalJSON(arg, &opts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s args, got: %v: %w", cmdTagList, arg, err)
		}
	}

	var sorters []core.CollectionSorter
	var err error
	if opts.Sort != nil {
		sorters, err = core.CollectionSortersFromStrings(opts.Sort)
		if err != nil {
			return nil, err
		}
	}
	return notebook.FindCollections(core.CollectionKindTag, sorters)
}
