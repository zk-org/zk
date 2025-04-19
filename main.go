package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/cmd"
	"github.com/zk-org/zk/internal/core"
	executil "github.com/zk-org/zk/internal/util/exec"
)

var Version = "dev"
var Build = "dev"

var root struct {
	Init  cmd.Init  `cmd group:"zk" help:"Create a new notebook in the given directory."`
	Index cmd.Index `cmd group:"zk" help:"Index the notes to be searchable."`
	Config cmd.Config `cmd group:"zk" help:"List configuration parameters."`

	New   cmd.New   `cmd group:"notes" help:"Create a new note in the given notebook directory."`
	List  cmd.List  `cmd group:"notes" help:"List notes matching the given criteria."`
	Graph cmd.Graph `cmd group:"notes" help:"Produce a graph of the notes matching the given criteria."`
	Edit  cmd.Edit  `cmd group:"notes" help:"Edit notes matching the given criteria."`
	Tag   cmd.Tag   `cmd group:"notes" help:"Manage the note tags."`

	NotebookDir string  `type:path placeholder:PATH help:"Turn off notebook auto-discovery and set manually the notebook where commands are run."`
	WorkingDir  string  `short:W type:path placeholder:PATH help:"Run as if zk was started in <PATH> instead of the current working directory."`
	NoInput     NoInput `help:"Never prompt or ask for confirmation."`
	// ForceInput is a debugging flag overriding the default value of interaction prompts.
	ForceInput string `hidden xor:"input"`
	Debug      bool   `default:"0" hidden help:"Print a debug stacktrace on SIGINT."`
	DebugStyle bool   `default:"0" hidden help:"Force styling output as XML tags."`

	ShowHelp ShowHelp         `cmd hidden default:"1"`
	LSP      cmd.LSP          `cmd hidden`
	Version  kong.VersionFlag `hidden help:"Print zk version."`
}

// NoInput is a flag preventing any user prompt when enabled.
type NoInput bool

func (f NoInput) BeforeApply(container *cli.Container) error {
	container.Terminal.NoInput = true
	return nil
}

// ShowHelp is the default command run. It's equivalent to `zk --help`.
type ShowHelp struct{}

func (cmd *ShowHelp) Run(container *cli.Container) error {
	parser, err := kong.New(&root, options(container)...)
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
	args := os.Args[1:]

	// Create the dependency graph.
	container, err := cli.NewContainer(Version)
	fatalIfError(err)

	// Open the notebook if there's any.
	dirs, args, err := parseDirs(args)
	fatalIfError(err)
	searchDirs, err := notebookSearchDirs(dirs)
	fatalIfError(err)
	err = container.SetCurrentNotebook(searchDirs)
	fatalIfError(err)

	// Run the alias or command.
	if isAlias, err := runAlias(container, args); isAlias {
		fatalIfError(err)
	} else {
		parser, err := kong.New(&root, options(container)...)
		fatalIfError(err)
		ctx, err := parser.Parse(args)
		fatalIfError(err)

		if root.Debug {
			setupDebugMode()
		}
		if root.DebugStyle {
			container.Styler.Styler = core.TagStyler
		}

		container.Terminal.ForceInput = root.ForceInput

		// Index the current notebook except if the user is running the `index`
		// command, otherwise it would hide the stats.
		if ctx.Command() != "index" {
			if notebook, err := container.CurrentNotebook(); err == nil {
				index := cmd.Index{Quiet: true}
				err = index.RunWithNotebook(container, notebook)
				ctx.FatalIfErrorf(err)
			}
		}

		err = ctx.Run(container)
		ctx.FatalIfErrorf(err)
	}
}

func options(container *cli.Container) []kong.Option {
	term := container.Terminal
	return []kong.Option{
		kong.Bind(container),
		kong.Name("zk"),
		kong.UsageOnError(),
		kong.HelpOptions{
			Compact:             true,
			FlagsLast:           true,
			WrapUpperBound:      100,
			NoExpandSubcommands: true,
		},
		kong.Vars{
			"version": "zk " + strings.TrimPrefix(Version, "v"),
		},
		kong.Groups(map[string]string{
			"cmd":    "Commands:",
			"filter": "Filtering",
			"sort":   "Sorting",
			"format": "Formatting",
			"notes":  term.MustStyle("NOTES", core.StyleYellow, core.StyleBold) + "\n" + term.MustStyle("Edit or browse your notes", core.StyleBold),
			"zk":     term.MustStyle("NOTEBOOK", core.StyleYellow, core.StyleBold) + "\n" + term.MustStyle("A notebook is a directory containing a collection of notes", core.StyleBold),
		}),
	}
}

func fatalIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "zk: error: %v\n", err)
		os.Exit(1)
	}
}

func setupDebugMode() {
	c := make(chan os.Signal)
	go func() {
		stacktrace := make([]byte, 8192)
		for _ = range c {
			length := runtime.Stack(stacktrace, true)
			fmt.Fprintf(os.Stderr, "%s\n", string(stacktrace[:length]))
			os.Exit(1)
		}
	}()
	signal.Notify(c, os.Interrupt)
}

// runAlias will execute a user alias if the command is one of them.
func runAlias(container *cli.Container, args []string) (bool, error) {
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

		// Move to the current notebook's root directory before running the alias.
		if notebook, err := container.CurrentNotebook(); err == nil {
			cmdStr = `cd "` + notebook.Path + `" && ` + cmdStr
		}

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

// notebookSearchDirs returns the places where zk will look for a notebook.
// The first successful candidate will be used as the working directory from
// which path arguments are relative from.
//
// By order of precedence:
//  1. --notebook-dir flag
//  2. current working directory
//  3. ZK_NOTEBOOK_DIR environment variable
func notebookSearchDirs(dirs cli.Dirs) ([]cli.Dirs, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// 1. --notebook-dir flag
	if dirs.NotebookDir != "" {
		// If --notebook-dir is used, we want to only check there to report
		// "notebook not found" errors.
		if dirs.WorkingDir == "" {
			dirs.WorkingDir = wd
		}
		return []cli.Dirs{dirs}, nil
	}

	candidates := []cli.Dirs{}

	// 2. current working directory
	wdDirs := dirs
	if wdDirs.WorkingDir == "" {
		wdDirs.WorkingDir = wd
	}
	wdDirs.NotebookDir = wdDirs.WorkingDir
	candidates = append(candidates, wdDirs)

	// 3. ZK_NOTEBOOK_DIR environment variable
	if notebookDir, ok := os.LookupEnv("ZK_NOTEBOOK_DIR"); ok {
		dirs := dirs
		dirs.NotebookDir = notebookDir
		if dirs.WorkingDir == "" {
			dirs.WorkingDir = notebookDir
		}
		candidates = append(candidates, dirs)
	}

	return candidates, nil
}

// parseDirs returns the paths specified with the --notebook-dir and
// --working-dir flags.
//
// We need to parse these flags before Kong, because we might need it to
// resolve zk command aliases before parsing the CLI.
func parseDirs(args []string) (cli.Dirs, []string, error) {
	var d cli.Dirs
	var err error

	// Split str by first "=" if present and return the split pair, otherwise return nil
	makeSplitPair := func(str string) (pair []string) {
		re := regexp.MustCompile(`=`)
		slice := re.FindStringIndex(str)
		if slice == nil {
			return nil
		}
		return []string{str[:slice[0]], str[slice[1]:]}
	}

	// Peek ahead at next value  and pair with current if it exists, otherwise return nil
	makePeekPair := func(args []string, index int) (pair []string) {
		if len(args) <= (index + 1) {
			return nil
		}
		return []string{args[index], args[index+1]}
	}

	matchesLongOrShort := func(str string, long string, short string) bool {
		return str == long || (short != "" && str == short)
	}

	findFlag := func(long string, short string, args []string) (string, []string, error) {
		newArgs := []string{}
		for i, arg := range args {
			// We can be given "--notebook-dir x" (two args) or "--notebook-dir=x" (one arg)
			// so we must test against the current argument split into two, and
			// the current argument + the next.
			splitPair := makeSplitPair(arg)
			peekPair := makePeekPair(args, i)
			var option string
			var value string

			if splitPair != nil && matchesLongOrShort(splitPair[0], long, short) {
				option = splitPair[0]
				value = splitPair[1]
				// skip 1 ahead
				newArgs = append(newArgs, args[i+1:]...)
			} else if peekPair != nil && matchesLongOrShort(peekPair[0], long, short) {
				option = peekPair[0]
				value = peekPair[1]
				// skip 2 ahead (arg and value)
				newArgs = append(newArgs, args[i+2:]...)
			} else {
				// we either had no split pair or peek pair, or they didn't match the
				// needle, so just save the given arg and keep looking.
				newArgs = append(newArgs, arg)
			}

			if option != "" && value != "" {
				path, err := filepath.Abs(value)
				return path, newArgs, err
			} else if option != "" && value == "" {
				return "", newArgs, errors.New(option + " requires a path argument")
			} else if len(args) == (i+1) && matchesLongOrShort(arg, long, short) {
				return "", newArgs, errors.New(arg + " requires a path argument")
			}
		}
		return "", newArgs, nil
	}

	d.NotebookDir, args, err = findFlag("--notebook-dir", "", args)
	if err != nil {
		return d, args, err
	}
	d.WorkingDir, args, err = findFlag("--working-dir", "-W", args)
	if err != nil {
		return d, args, err
	}

	return d, args, nil
}
