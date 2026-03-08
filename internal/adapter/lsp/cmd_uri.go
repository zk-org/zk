package lsp

import (
	"errors"
	"fmt"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/zk-org/zk/internal/core"
)

const cmdURI = "zk.uri"

type cmdURIOpts struct {
	Location *protocol.Location `json:"location"`
}

// executeCommandURI returns the zk:// URI of the given location, to be linked
// to from outside systems.
//
// The URI is returned rather than acted upon, as a language server has no way
// to reach the editor's clipboard: it is up to the client to decide what to do
// with it.
func executeCommandURI(notebook *core.Notebook, documents *documentStore, args []any) (any, error) {
	var opts cmdURIOpts

	if len(args) > 1 {
		arg, ok := args[1].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s expects a dictionary of options as second argument, got: %v", cmdURI, args[1])
		}
		err := unmarshalJSON(arg, &opts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s args, got: %v: %w", cmdURI, arg, err)
		}
	}

	if opts.Location == nil {
		return nil, errors.New("'location' not provided")
	}

	path, err := uriToPath(opts.Location.URI)
	if err != nil {
		return nil, err
	}
	relPath, err := notebook.RelPath(path)
	if err != nil {
		return nil, err
	}

	// LSP positions are zero-based, while zk:// URIs use the one-based
	// convention of text editors.
	start := opts.Location.Range.Start
	line := int(start.Line) + 1
	column := int(start.Character) + 1
	if doc, ok := documents.Get(opts.Location.URI); ok {
		// Character is a UTF-16 offset, which is not a column.
		column = doc.RuneColumnAt(start)
	}

	uri := core.NoteURI{
		NotebookPath: notebook.Path,
		Path:         relPath,
		Line:         line,
		Column:       column,
	}

	return map[string]any{
		"uri":      uri.String(),
		"notebook": uri.NotebookPath,
		"path":     uri.Path,
		"line":     line,
		"column":   column,
	}, nil
}
