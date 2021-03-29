package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mickael-menu/zk/adapter"
	"github.com/mickael-menu/zk/cmd"
	"github.com/mickael-menu/zk/core/style"
	executil "github.com/mickael-menu/zk/util/exec"
)

var Version = "dev"
var Build = "dev"

var cli struct {
	Init  cmd.Init  `cmd group:"zk" help:"Create a new notebook in the given directory."`
	Index cmd.Index `cmd group:"zk" help:"Index the notes to be searchable."`

	New  cmd.New  `cmd group:"notes" help:"Create a new note in the given notebook directory."`
	List cmd.List `cmd group:"notes" help:"List notes matching the given criteria."`
	Edit cmd.Edit `cmd group:"notes" help:"Edit notes matching the given criteria."`

	NoInput     NoInput `help:"Never prompt or ask for confirmation."`
	NotebookDir string  `placeholder:"PATH" help:"Run as if zk was started in <PATH> instead of the current working directory."`

	ShowHelp ShowHelp         `cmd hidden default:"1"`
	LSP      cmd.LSP          `cmd hidden`
	Version  kong.VersionFlag `hidden help:"Print zk version."`
}

// NoInput is a flag preventing any user prompt when enabled.
type NoInput bool

func (f NoInput) BeforeApply(container *adapter.Container) error {
	container.Terminal.NoInput = true
	return nil
}

// ShowHelp is the default command run. It's equivalent to `zk --help`.
type ShowHelp struct{}

func (cmd *ShowHelp) Run(container *adapter.Container) error {
	parser, err := kong.New(&cli, options(container)...)
	if err != nil {
		return err
	}
	ctx, err := parser.Parse([]string{"--help"})
	if err != nil {
		return err
	}
	return ctx.Run(container)
}

func main() {
	// Create the dependency graph.
	container, err := adapter.NewContainer(Version)
	fatalIfError(err)

	// Open the notebook if there's any.
	searchPaths, err := notebookSearchPaths()
	fatalIfError(err)
	container.OpenNotebook(searchPaths)

	// Run the alias or command.
	if isAlias, err := runAlias(container, os.Args[1:]); isAlias {
		fatalIfError(err)
	} else {
		ctx := kong.Parse(&cli, options(container)...)
		err := ctx.Run(container)
		ctx.FatalIfErrorf(err)
	}
}

func options(container *adapter.Container) []kong.Option {
	term := container.Terminal
	return []kong.Option{
		kong.Bind(container),
		kong.Name("zk"),
		kong.UsageOnError(),
		kong.HelpOptions{
			Compact:   true,
			FlagsLast: true,
		},
		kong.Vars{
			"version": "zk " + strings.TrimPrefix(Version, "v"),
		},
		kong.Groups(map[string]string{
			"filter": "Filtering",
			"sort":   "Sorting",
			"format": "Formatting",
			"notes":  term.MustStyle("NOTES", style.RuleYellow, style.RuleBold) + "\n" + term.MustStyle("Edit or browse your notes", style.RuleBold),
			"zk":     term.MustStyle("NOTEBOOK", style.RuleYellow, style.RuleBold) + "\n" + term.MustStyle("A notebook is a directory containing a collection of notes", style.RuleBold),
		}),
	}
}

func fatalIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "zk: error: %v\n", err)
		os.Exit(1)
	}
}

// runAlias will execute a user alias if the command is one of them.
func runAlias(container *adapter.Container, args []string) (bool, error) {
	if len(args) < 1 {
		return false, nil
	}

	runningAlias := os.Getenv("ZK_RUNNING_ALIAS")
	for alias, cmdStr := range container.Config.Aliases {
		if alias == runningAlias || alias != args[0] {
			continue
		}

		// Prevent infinite loop if an alias calls itself.
		os.Setenv("ZK_RUNNING_ALIAS", alias)

		// Move to the provided working directory if it is not the current one,
		// before running the alias.
		cmdStr = `cd "` + container.WorkingDir + `" && ` + cmdStr

		cmd := executil.CommandFromString(cmdStr, args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				os.Exit(err.ExitCode())
				return true, nil
			} else {
				return true, err
			}
		}
		return true, nil
	}

	return false, nil
}

// notebookSearchPaths returns the places where zk will look for a notebook.
// The first successful candidate will be used as the working directory from
// which path arguments are relative from.
//
// By order of precedence:
//   1. --notebook-dir flag
//   2. current working directory
//   3. ZK_NOTEBOOK_DIR environment variable
func notebookSearchPaths() ([]string, error) {
	// 1. --notebook-dir flag
	notebookDir, err := parseNotebookDirFlag()
	if err != nil {
		return []string{}, err
	}
	if notebookDir != "" {
		// If --notebook-dir is used, we want to only check there to report errors.
		return []string{notebookDir}, nil
	}

	candidates := []string{}

	// 2. current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, wd)

	// 3. ZK_NOTEBOOK_DIR environment variable
	if notebookDir, ok := os.LookupEnv("ZK_NOTEBOOK_DIR"); ok {
		candidates = append(candidates, notebookDir)
	}

	return candidates, nil
}

// parseNotebookDir returns the path to the notebook specified with the
// --notebook-dir flag.
//
// We need to parse the --notebook-dir flag before Kong, because we might need
// it to resolve zk command aliases before parsing the CLI.
func parseNotebookDirFlag() (string, error) {
	foundFlag := false
	for _, arg := range os.Args {
		if arg == "--notebook-dir" {
			foundFlag = true
		} else if foundFlag {
			return arg, nil
		}
	}
	if foundFlag {
		return "", errors.New("--notebook-dir requires an argument")
	}
	return "", nil
}
