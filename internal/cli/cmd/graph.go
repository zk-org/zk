package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	stdstrings "strings"

	"github.com/zk-org/zk/internal/adapter/fzf"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/strings"
)

// Graph produces a directed graph of the notes matching a set of criteria.
type Graph struct {
	Format string `group:format short:f                        help:"Format of the graph among: json, web." enum:"json,web" required`
	Quiet  bool   `group:format short:q help:"Do not print the total number of notes found."`
	cli.Filtering
}

func (cmd *Graph) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	format, err := notebook.NewNoteFormatter("{{json .}}")
	if err != nil {
		return err
	}

	findOpts, err := cmd.NewNoteFindOpts(notebook)
	if err != nil {
		return errors.Wrapf(err, "incorrect criteria")
	}

	notes, err := notebook.FindNotes(findOpts)
	if err != nil {
		return err
	}

	filter := container.NewNoteFilter(fzf.NoteFilterOpts{
		Interactive:  cmd.Interactive,
		AlwaysFilter: false,
		NotebookDir:  notebook.Path,
	})

	notes, err = filter.Apply(notes)
	if err != nil {
		if err == fzf.ErrCancelled {
			return nil
		}
		return err
	}

	noteIDs := make([]core.NoteID, 0, len(notes))
	for _, note := range notes {
		noteIDs = append(noteIDs, note.ID)
	}
	links, err := notebook.FindLinksBetweenNotes(noteIDs)
	if err != nil {
		return err
	}

	switch cmd.Format {
	case "json":
		err = cmd.printJSON(notes, links, format)
	case "web":
		err = cmd.openWebView(notes, links)
	}
	if err != nil {
		return err
	}

	if !cmd.Quiet {
		count := len(notes)
		fmt.Fprintf(os.Stderr, "\nFound %d %s\n", count, strings.Pluralize("note", count))
	}

	return nil
}

func (cmd *Graph) printJSON(notes []core.ContextualNote, links []core.ResolvedLink, format core.NoteFormatter) error {
	fmt.Print("{\n  \"notes\": [\n")
	for i, note := range notes {
		if i > 0 {
			fmt.Print(",\n")
		}
		ft, err := format(note)
		if err != nil {
			return err
		}
		fmt.Printf("    %s", ft)
	}

	fmt.Print("\n  ],\n  \"links\": [\n")
	for i, link := range links {
		if i > 0 {
			fmt.Print(",\n")
		}
		ft, err := json.Marshal(link)
		if err != nil {
			return err
		}
		fmt.Printf("    %s", string(ft))
	}

	fmt.Print("\n  ]\n}\n")
	return nil
}

func (cmd *Graph) openWebView(notes []core.ContextualNote, links []core.ResolvedLink) error {
	page, err := renderWebGraphPage(graphPayloadFrom(notes, links))
	if err != nil {
		return err
	}

	file, err := os.CreateTemp("", "zk-graph-*.html")
	if err != nil {
		return errors.Wrap(err, "failed to create a temporary graph page")
	}
	defer file.Close()

	_, err = file.WriteString(page)
	if err != nil {
		return errors.Wrap(err, "failed to write the temporary graph page")
	}

	// Tests and non-interactive checks should not try to open a GUI browser.
	if _, runningInTesh := os.LookupEnv("RUNNING_TESH"); runningInTesh {
		return nil
	}

	return openInBrowser(file.Name())
}

type graphPayload struct {
	Nodes []graphNode `json:"nodes"`
	Links []graphLink `json:"links"`
}

type graphNode struct {
	ID    int64    `json:"id"`
	Title string   `json:"title"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
}

type graphLink struct {
	Source int64  `json:"source"`
	Target int64  `json:"target"`
	Type   string `json:"type"`
}

func graphPayloadFrom(notes []core.ContextualNote, links []core.ResolvedLink) graphPayload {
	nodes := make([]graphNode, 0, len(notes))
	for _, note := range notes {
		nodes = append(nodes, graphNode{
			ID:    int64(note.ID),
			Title: note.Title,
			Path:  note.Path,
			Tags:  note.Tags,
		})
	}

	graphLinks := make([]graphLink, 0, len(links))
	for _, link := range links {
		graphLinks = append(graphLinks, graphLink{
			Source: int64(link.SourceID),
			Target: int64(link.TargetID),
			Type:   string(link.Type),
		})
	}

	return graphPayload{
		Nodes: nodes,
		Links: graphLinks,
	}
}

func renderWebGraphPage(graph graphPayload) (string, error) {
	dataJSON, err := json.Marshal(graph)
	if err != nil {
		return "", errors.Wrap(err, "failed to serialize graph data")
	}
	return stdstrings.ReplaceAll(graphHTMLTemplate, "__DATA__", string(dataJSON)), nil
}

func openInBrowser(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to open graph view in browser")
	}
	return nil
}

//go:embed assets/graph.html
var graphHTMLTemplate string
