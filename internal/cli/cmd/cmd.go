package cmd

import (
	"path/filepath"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/paths"
)

// Command groups, as presented in `zk --help`.
const (
	GroupNotebook = "zk"
	GroupNotes    = "notes"
)

// Globals holds the flags every command accepts.
type Globals struct {
	NotebookDir string
	WorkingDir  string
	NoInput     bool
	// ForceInput overrides the default answer to interaction prompts, to be
	// able to test them.
	ForceInput string
	Debug      bool
	DebugStyle bool
}

// Flags describes the global flags.
func (g *Globals) Flags() []spec.Flag {
	return []spec.Flag{
		{Name: "notebook-dir", Placeholder: "PATH", Target: &g.NotebookDir, CompleteFiles: true,
			Help: "Turn off notebook auto-discovery and set manually the notebook where commands are run."},
		{Name: "working-dir", Short: "W", Placeholder: "PATH", Target: &g.WorkingDir, CompleteFiles: true,
			Help: "Run as if zk was started in <PATH> instead of the current working directory."},
		{Name: "no-input", Target: &g.NoInput,
			Help: "Never prompt or ask for confirmation."},
		{Name: "force-input", Hidden: true, Target: &g.ForceInput},
		{Name: "debug", Hidden: true, Target: &g.Debug,
			Help: "Print a debug stacktrace on SIGINT."},
		{Name: "debug-style", Hidden: true, Target: &g.DebugStyle,
			Help: "Force styling output as XML tags."},
	}
}

// App describes zk's whole command line: every command, with the flags and
// arguments it takes and how to run it.
//
// This is the only place the commands are declared. A front-end reads it to
// present them, so a new flag shows up in the help, in the shell completions
// and anywhere else zk is driven from, without being restated.
func App(container *cli.Container, globals *Globals) *spec.App {
	term := container.Terminal

	return &spec.App{
		Name:    "zk",
		Version: container.Version,
		Flags:   globals.Flags(),
		Commands: []*spec.Command{
			InitSpec(container),
			IndexSpec(container),
			ConfigSpec(container),

			NewSpec(container),
			ListSpec(container),
			GraphSpec(container),
			EditSpec(container),
			OpenSpec(container),
			URISpec(container),
			TagSpec(container),

			LSPSpec(container),
		},
		CommandGroups: []spec.Group{
			{
				ID:          GroupNotebook,
				Title:       term.MustStyle("NOTEBOOK", core.StyleYellow, core.StyleBold),
				Description: term.MustStyle("A notebook is a directory containing a collection of notes", core.StyleBold),
			},
			{
				ID:          GroupNotes,
				Title:       term.MustStyle("NOTES", core.StyleYellow, core.StyleBold),
				Description: term.MustStyle("Edit or browse your notes", core.StyleBold),
			},
		},
		FlagGroups: []spec.Group{
			{ID: spec.GroupFormat, Title: "Formatting"},
			{ID: spec.GroupFilter, Title: "Filtering"},
			{ID: spec.GroupSort, Title: "Sorting"},
		},
	}
}

// expandPath resolves a path argument the way zk's path options always did:
// expanding `~` and environment variables, then making it absolute.
func expandPath(path string) (string, error) {
	expanded, err := paths.ExpandPath(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(expanded)
}
