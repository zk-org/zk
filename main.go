package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/mickael-menu/zk/cmd"
	executil "github.com/mickael-menu/zk/util/exec"
)

var Version = "dev"
var Build = "dev"

var cli struct {
	Index   cmd.Index        `cmd help:"Index the notes in the given directory to be searchable"`
	Init    cmd.Init         `cmd help:"Create a slip box in the given directory"`
	List    cmd.List         `cmd help:"List notes matching given criteria"`
	Edit    cmd.Edit         `cmd help:"Edit notes matching given criteria"`
	New     cmd.New          `cmd help:"Create a new note in the given slip box directory"`
	NoInput NoInput          `help:"Never prompt or ask for confirmation"`
	Version kong.VersionFlag `help:"Print zk version"`
}

// NoInput is a flag preventing any user prompt when enabled.
type NoInput bool

func (f NoInput) BeforeApply(container *cmd.Container) error {
	container.Terminal.NoInput = true
	return nil
}

func main() {
	// Create the dependency graph.
	container := cmd.NewContainer()

	indexZk(container)

	if isAlias, err := runAlias(container, os.Args[1:]); isAlias {
		fatalIfError(err)

	} else {
		ctx := kong.Parse(&cli,
			kong.Bind(container),
			kong.Name("zk"),
			kong.Vars{
				"version": Version,
			},
		)

		err := ctx.Run(container)
		ctx.FatalIfErrorf(err)
	}
}

func fatalIfError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "zk: error: %v\n", err)
		os.Exit(1)
	}
}

// indexZk will index any slip box in the working directory.
func indexZk(container *cmd.Container) {
	(&cmd.Index{Quiet: true}).Run(container)
}

// runAlias will execute a user alias if the command is one of them.
func runAlias(container *cmd.Container, args []string) (bool, error) {
	if zk, err := container.OpenZk(); err == nil && len(args) >= 1 {
		for alias, cmdStr := range zk.Config.Aliases {
			if alias != args[0] {
				continue
			}

			cmd := executil.CommandFromString(cmdStr, args...)
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
	}

	return false, nil
}
