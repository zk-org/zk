package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func pathToURI(path string) string {
	u := &url.URL{
		Scheme: "file",
		Path:   path,
	}
	return u.String()
}

func uriToPath(uri string) (string, error) {
	s := strings.ReplaceAll(uri, "%5C", "/")
	parsed, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "file" {
		return "", errors.New("URI was not a file:// URI")
	}

	if runtime.GOOS == "windows" {
		// In Windows "file:///c:/tmp/foo.md" is parsed to "/c:/tmp/foo.md".
		// Strip the first character to get a valid path.
		if strings.Contains(parsed.Path[1:], ":") {
			// url.Parse() behaves differently with "file:///c:/..." and "file://c:/..."
			return parsed.Path[1:], nil
		} else {
			// if the windows drive is not included in Path it will be in Host
			return parsed.Host + "/" + parsed.Path[1:], nil
		}
	}
	return parsed.Path, nil
}

// jsonBoolean can be unmarshalled from integers or strings.
// Neovim cannot send a boolean easily, so it's useful to support integers too.
type jsonBoolean bool

func (b *jsonBoolean) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "1" || s == "true" {
		*b = true
	} else if s == "0" || s == "false" {
		*b = false
	} else {
		return fmt.Errorf("%s: failed to unmarshal as boolean", s)
	}
	return nil
}

type linkInfo struct {
    note *core.MinimalNote
	location *protocol.Location
	title    *string
}

func linkNote(notebook *core.Notebook, documents *documentStore, context *glsp.Context, info *linkInfo) (bool, error) {
	if info.location == nil {
		return false, errors.New("'location' not provided")
	}

	// Get current document to edit
	doc, ok := documents.Get(info.location.URI)
	if !ok {
		return false, fmt.Errorf("Cannot insert link in '%s'", info.location.URI)
	}

	formatter, err := notebook.NewLinkFormatter()
	if err != nil {
		return false, err
	}

	path := core.NotebookPath{
		Path:       info.note.Path,
		BasePath:   notebook.Path,
		WorkingDir: filepath.Dir(doc.Path),
	}

	var title *string
	title = info.title

	if title == nil {
		title = &info.note.Title
	}

	formatterContext, err := core.NewLinkFormatterContext(path, *title, info.note.Metadata)
	if err != nil {
		return false, err
	}

	link, err := formatter(formatterContext)
	if err != nil {
		return false, err
	}

	go context.Call(protocol.ServerWorkspaceApplyEdit, protocol.ApplyWorkspaceEditParams{
		Edit: protocol.WorkspaceEdit{
			Changes: map[string][]protocol.TextEdit{
				info.location.URI: {{Range: info.location.Range, NewText: link}},
			},
		},
	}, nil)

	return true, nil
}
