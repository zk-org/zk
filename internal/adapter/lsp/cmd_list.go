package lsp

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
	"github.com/mickael-menu/zk/internal/util/errors"
	strutil "github.com/mickael-menu/zk/internal/util/strings"
)

const cmdList = "zk.list"

type cmdListOpts struct {
	Select         []string `json:"select"`
	Href           []string `json:"hrefs"`
	Limit          int      `json:"limit"`
	Match          string   `json:"match"`
	ExactMatch     bool     `json:"exactMatch"`
	Exclude        []string `json:"excludeHrefs"`
	Tag            []string `json:"tags"`
	Mention        []string `json:"mention"`
	MentionedBy    []string `json:"mentionedBy"`
	LinkTo         []string `json:"linkTo"`
	LinkedBy       []string `json:"linkedBy"`
	Orphan         bool     `json:"orphan"`
	Related        []string `json:"related"`
	MaxDistance    int      `json:"maxDistance"`
	Recursive      bool     `json:"recursive"`
	Created        string   `json:"created"`
	CreatedBefore  string   `json:"createdBefore"`
	CreatedAfter   string   `json:"createdAfter"`
	Modified       string   `json:"modified"`
	ModifiedBefore string   `json:"modifiedBefore"`
	ModifiedAfter  string   `json:"modifiedAfter"`
	Sort           []string `json:"sort"`
}

func executeCommandList(logger util.Logger, notebook *core.Notebook, args []interface{}) (interface{}, error) {
	var opts cmdListOpts
	if len(args) > 1 {
		arg, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s expects a dictionary of options as second argument, got: %v", cmdTagList, args[1])
		}
		err := unmarshalJSON(arg, &opts)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s args, got: %v", cmdTagList, arg)
		}
	}

	if len(opts.Select) == 0 {
		return nil, fmt.Errorf("%s expects a `select` option with the list of fields to return", cmdTagList)
	}
	var selection = newListSelection(opts.Select)

	var findOpts core.NoteFindOpts
	notes, err := notebook.FindNotes(findOpts)
	if err != nil {
		return nil, err
	}

	listNotes := []listNote{}
	for _, note := range notes {
		listNotes = append(listNotes, newListNote(note, selection, notebook.Path))
	}

	return listNotes, nil
}

type listSelection struct {
	Filename     bool
	FilenameStem bool
	Path         bool
	AbsPath      bool
	Title        bool
	Lead         bool
	Body         bool
	Snippets     bool
	RawContent   bool
	WordCount    bool
	Tags         bool
	Metadata     bool
	Created      bool
	Modified     bool
	Checksum     bool
}

func newListSelection(fields []string) listSelection {
	return listSelection{
		Filename:     strutil.Contains(fields, "filename"),
		FilenameStem: strutil.Contains(fields, "filenameStem"),
		Path:         strutil.Contains(fields, "path"),
		AbsPath:      strutil.Contains(fields, "absPath"),
		Title:        strutil.Contains(fields, "title"),
		Lead:         strutil.Contains(fields, "lead"),
		Body:         strutil.Contains(fields, "body"),
		Snippets:     strutil.Contains(fields, "snippets"),
		RawContent:   strutil.Contains(fields, "rawContent"),
		WordCount:    strutil.Contains(fields, "wordCount"),
		Tags:         strutil.Contains(fields, "tags"),
		Metadata:     strutil.Contains(fields, "metadata"),
		Created:      strutil.Contains(fields, "created"),
		Modified:     strutil.Contains(fields, "modified"),
		Checksum:     strutil.Contains(fields, "checksum"),
	}
}

func newListNote(note core.ContextualNote, selection listSelection, basePath string) listNote {
	var res listNote
	if selection.Filename {
		res.Filename = note.Filename()
	}
	if selection.FilenameStem {
		res.FilenameStem = note.FilenameStem()
	}
	if selection.Path {
		res.Path = note.Path
	}
	if selection.AbsPath {
		res.AbsPath = filepath.Join(basePath, note.Path)
	}
	if selection.Title {
		res.Title = note.Title
	}
	if selection.Lead {
		res.Lead = note.Lead
	}
	if selection.Body {
		res.Body = note.Body
	}
	if selection.Snippets {
		res.Snippets = note.Snippets
	}
	if selection.RawContent {
		res.RawContent = note.RawContent
	}
	if selection.WordCount {
		res.WordCount = note.WordCount
	}
	if selection.Tags {
		res.Tags = note.Tags
	}
	if selection.Metadata {
		res.Metadata = note.Metadata
	}
	if selection.Created {
		res.Created = &note.Created
	}
	if selection.Modified {
		res.Modified = &note.Modified
	}
	if selection.Checksum {
		res.Checksum = note.Checksum
	}
	return res
}

type listNote struct {
	Filename     string                 `json:"filename,omitempty"`
	FilenameStem string                 `json:"filenameStem,omitempty"`
	Path         string                 `json:"path,omitempty"`
	AbsPath      string                 `json:"absPath,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Lead         string                 `json:"lead,omitempty"`
	Body         string                 `json:"body,omitempty"`
	Snippets     []string               `json:"snippets,omitempty"`
	RawContent   string                 `json:"rawContent,omitempty"`
	WordCount    int                    `json:"wordCount,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Created      *time.Time             `json:"created,omitempty"`
	Modified     *time.Time             `json:"modified,omitempty"`
	Checksum     string                 `json:"checksum,omitempty"`
}
