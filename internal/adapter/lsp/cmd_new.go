package lsp

import (
	"fmt"
	"path/filepath"

	"github.com/mickael-menu/zk/internal/core"
	dateutil "github.com/mickael-menu/zk/internal/util/date"
	"github.com/mickael-menu/zk/internal/util/errors"
	"github.com/mickael-menu/zk/internal/util/opt"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

const cmdNew = "zk.new"

type cmdNewOpts struct {
	Title                string             `json:"title,omitempty"`
	Content              string             `json:"content,omitempty"`
	Dir                  string             `json:"dir,omitempty"`
	Group                string             `json:"group,omitempty"`
	Template             string             `json:"template,omitempty"`
	Extra                map[string]string  `json:"extra,omitempty"`
	Date                 string             `json:"date,omitempty"`
	Edit                 jsonBoolean        `json:"edit,omitempty"`
	InsertLinkAtLocation *protocol.Location `json:"insertLinkAtLocation,omitempty"`
}

func executeCommandNew(notebook *core.Notebook, documents *documentStore, context *glsp.Context, args []interface{}) (interface{}, error) {
	var opts cmdNewOpts
	if len(args) > 1 {
		arg, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s expects a dictionary of options as second argument, got: %v", cmdNew, args[1])
		}
		err := unmarshalJSON(arg, &opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s args, got: %v", cmdNew, arg)
		}
	}

	date, err := dateutil.TimeFromNatural(opts.Date)
	if err != nil {
		return nil, errors.Wrapf(err, "%s, failed to parse the `date` option", opts.Date)
	}

	note, err := notebook.NewNote(core.NewNoteOpts{
		Title:     opt.NewNotEmptyString(opts.Title),
		Content:   opts.Content,
		Directory: opt.NewNotEmptyString(opts.Dir),
		Group:     opt.NewNotEmptyString(opts.Group),
		Template:  opt.NewNotEmptyString(opts.Template),
		Extra:     opts.Extra,
		Date:      date,
	})
	if err != nil {
		var noteExists core.ErrNoteExists
		if !errors.As(err, &noteExists) {
			return nil, err
		}
		note, err = notebook.FindNote(core.NoteFindOpts{
			IncludeHrefs: []string{noteExists.Name},
		})
		if err != nil {
			return nil, err
		}
	}
	if note == nil {
		return nil, errors.New("zk.new could not generate a new note")
	}

	if opts.InsertLinkAtLocation != nil {
		doc, ok := documents.Get(opts.InsertLinkAtLocation.URI)
		if !ok {
			return nil, fmt.Errorf("can't insert link in %s", opts.InsertLinkAtLocation.URI)
		}
		linkFormatter, err := notebook.NewLinkFormatter()
		if err != nil {
			return nil, err
		}

		currentDir := filepath.Dir(doc.Path)
		linkFormatterContext, err := core.NewLinkFormatterContext(note.AsMinimalNote(), notebook.Path, currentDir)
		if err != nil {
			return nil, err
		}

		link, err := linkFormatter(linkFormatterContext)
		if err != nil {
			return nil, err
		}

		go context.Call(protocol.ServerWorkspaceApplyEdit, protocol.ApplyWorkspaceEditParams{
			Edit: protocol.WorkspaceEdit{
				Changes: map[string][]protocol.TextEdit{
					opts.InsertLinkAtLocation.URI: {{Range: opts.InsertLinkAtLocation.Range, NewText: link}},
				},
			},
		}, nil)
	}

	absPath := filepath.Join(notebook.Path, note.Path)
	if opts.Edit {
		go context.Call(protocol.ServerWindowShowDocument, protocol.ShowDocumentParams{
			URI:       pathToURI(absPath),
			TakeFocus: boolPtr(true),
		}, nil)
	}

	return map[string]interface{}{"path": absPath}, nil
}
