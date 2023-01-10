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
	Title                   string             `json:"title"`
	Content                 string             `json:"content"`
	Dir                     string             `json:"dir"`
	Group                   string             `json:"group"`
	Template                string             `json:"template"`
	Extra                   map[string]string  `json:"extra"`
	Date                    string             `json:"date"`
	Edit                    jsonBoolean        `json:"edit"`
	DryRun                  jsonBoolean        `json:"dryRun"`
	InsertLinkAtLocation    *protocol.Location `json:"insertLinkAtLocation"`
	InsertContentAtLocation *protocol.Location `json:"insertContentAtLocation"`
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
		DryRun:    bool(opts.DryRun),
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

	if opts.InsertContentAtLocation != nil {
		go context.Call(protocol.ServerWorkspaceApplyEdit, protocol.ApplyWorkspaceEditParams{
			Edit: protocol.WorkspaceEdit{
				Changes: map[string][]protocol.TextEdit{
					opts.InsertContentAtLocation.URI: {{Range: opts.InsertContentAtLocation.Range, NewText: note.RawContent}},
				},
			},
		}, nil)
	}

	if !opts.DryRun && opts.InsertLinkAtLocation != nil {
        minNote := note.AsMinimalNote()

        info := &linkInfo{
            note: &minNote,
            location: opts.InsertLinkAtLocation,
            title: &opts.Title,
        }
        ok, err := linkNote(notebook, documents, context, info)

        if !ok {
            return nil, err
        }
	}

	absPath := filepath.Join(notebook.Path, note.Path)
	if !opts.DryRun && opts.Edit {
		go context.Call(protocol.ServerWindowShowDocument, protocol.ShowDocumentParams{
			URI:       pathToURI(absPath),
			TakeFocus: boolPtr(true),
		}, nil)
	}

	return map[string]interface{}{
		"path":    absPath,
		"content": note.RawContent,
	}, nil
}
