package lsp

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/core"
)

const cmdLink = "zk.link"

type cmdLinkOpts struct {
	Path     *string            `json:"path"`
	Location *protocol.Location `json:"location"`
	Title    *string            `json:"title"`
}

func executeCommandLink(notebook *core.Notebook, documents *documentStore, context *glsp.Context, args []any) (any, error) {
	var opts cmdLinkOpts

	if len(args) > 1 {
		arg, ok := args[1].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s expects a dictionary of options as second argument, got: %v", cmdLink, args[1])
		}
		err := unmarshalJSON(arg, &opts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s args, got: %v: %w", cmdLink, arg, err)
		}
	}

	if opts.Path == nil {
		return nil, errors.New("'path' not provided")
	}

	note, err := notebook.FindByHref(*opts.Path, false)

	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, errors.New("Requested note to link to not found!")
	}

	info := &linkInfo{
		note:     note,
		location: opts.Location,
		title:    opts.Title,
	}

	err = linkNote(notebook, documents, context, info)

	if err != nil {
		return nil, err
	}

	return map[string]any{
		"path": filepath.Join(notebook.Path, note.Path),
	}, nil
}
