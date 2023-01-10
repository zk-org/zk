package lsp

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const cmdLink = "zk.link"

type cmdLinkOpts struct {
	Path     *string            `json:"path"`
	Location *protocol.Location `json:"location"`
	Title    *string            `json:"title"`
}

func executeCommandLink(notebook *core.Notebook, documents *documentStore, context *glsp.Context, args []interface{}) (interface{}, error) {
	var opts cmdLinkOpts

	if len(args) > 1 {
		arg, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s expects a dictionary of options as second argument, got: %v", cmdLink, args[1])
		}
		err := unmarshalJSON(arg, &opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s args, got: %v", cmdLink, arg)
		}
	}

	if opts.Path == nil {
		return nil, errors.New("'path' not provided")
	}

	note, err := notebook.FindByHref(*opts.Path, false)

	if err != nil || note == nil {
		return nil, errors.New("Requested note to link to not found!")
	}

	info := &linkInfo{
		note:     note,
		location: opts.Location,
		title:    opts.Title,
	}

    ok, err := linkNote(notebook, documents, context, info)

    if !ok {
        return nil, err
    }

	return map[string]interface{}{
		"ok":   true,
		"path": filepath.Join(notebook.Path, note.Path),
	}, nil
}
