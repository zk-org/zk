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

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/cmd"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
	executil "github.com/zk-org/zk/internal/util/exec"
	"github.com/zk-org/zk/internal/util/paths"
)

var Version = "dev"

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
		return
	}

	var globals cmd.Globals
	app := cmd.App(container, &globals)

	// Everything the commands share is set up here, once the front-end has
	// read the global flags and knows which command was selected.
	before := func(command *spec.Command) error {
		if globals.NoInput {
			container.Terminal.NoInput = true
		}
		container.Terminal.ForceInput = globals.ForceInput
		if globals.Debug {
			setupDebugMode()
		}
		if globals.DebugStyle {
			container.Styler.Styler = core.TagStyler
		}

		if command.SkipAutoIndex {
			return nil
		}
		// Index the current notebook, so that every command works on an
		// up-to-date index without having to ask for it.
		notebook, err := container.CurrentNotebook()
		if err != nil {
			// Not being in a notebook is reported by the command itself, if
			// it needs one.
			return nil
		}
		index := cmd.Index{Quiet: true, Automatic: true}
		return index.RunWithNotebook(container, notebook)
	}

	fatalIfError(runFrontend(app, args, before))
}

// errorLeader prefixes every error zk reports.
const errorLeader = "zk: error: "

func fatalIfError(err error) {
	if err == nil {
		return
	}

	// An error spanning several lines, such as one listing the values a flag
	// accepts, is indented under the leader so that it reads as one message.
	lines := strings.Split(err.Error(), "\n")
	fmt.Fprintf(os.Stderr, "%s%s\n", errorLeader, lines[0])
	for _, line := range lines[1:] {
		fmt.Fprintf(os.Stderr, "%*s%s\n", len(errorLeader), " ", line)
	}

	os.Exit(1)
}

func setupDebugMode() {
	c := make(chan os.Signal, 1)
	go func() {
		stacktrace := make([]byte, 8192)
		for range c {
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
	for alias, spec := range container.Config.Aliases {
		if alias == runningAlias || alias != args[0] {
			continue
		}

		// Prevent infinite loop if an alias calls itself.
		os.Setenv("ZK_RUNNING_ALIAS", alias)

		cmdStr := spec.Command

		// Move to the current notebook's root directory before running the alias.
		if notebook, err := container.CurrentNotebook(); err == nil {
			cmdStr = `cd "` + notebook.Path + `" && ` + cmdStr
		}

		shell := executil.ResolveShell(container.Config.Tool.Shell)
		cmd := executil.CommandFromString(shell, cmdStr, args[1:]...)
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
// We need to parse these flags before the front-end does, because resolving
// a notebook's aliases needs them and happens before the command line is
// parsed.
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
				valueAbs, err := paths.ExpandPath(value)
				if err != nil {
					return "", newArgs, err
				}
				path, err := filepath.Abs(valueAbs)
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
