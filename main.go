package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/mickael-menu/zk/cmd"
	"github.com/mickael-menu/zk/core/style"
	executil "github.com/mickael-menu/zk/util/exec"
)

var Version = "dev"
var Build = "dev"

var cli struct {
	Init  cmd.Init  `cmd group:"zk" help:"Create a new slip box in the given directory."`
	Index cmd.Index `cmd group:"zk" help:"Index the notes to be searchable."`

	New  cmd.New  `cmd group:"notes" help:"Create a new note in the given slip box directory."`
	List cmd.List `cmd group:"notes" help:"List notes matching the given criteria."`
	Edit cmd.Edit `cmd group:"notes" help:"Edit notes matching the given criteria."`

	NoInput NoInput `help:"Never prompt or ask for confirmation."`

	ShowHelp ShowHelp         `cmd default:"1" hidden:true`
	Version  kong.VersionFlag `help:"Print zk version." hidden:true`
}

// NoInput is a flag preventing any user prompt when enabled.
type NoInput bool

func (f NoInput) BeforeApply(container *cmd.Container) error {
	container.Terminal.NoInput = true
	return nil
}

// ShowHelp is the default command run. It's equivalent to `zk --help`.
type ShowHelp struct{}

func (cmd *ShowHelp) Run(container *cmd.Container) error {
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
	container := cmd.NewContainer()

	indexZk(container)

	if isAlias, err := runAlias(container, os.Args[1:]); isAlias {
		fatalIfError(err)

	} else {
		ctx := kong.Parse(&cli, options(container)...)
		err := ctx.Run(container)
		ctx.FatalIfErrorf(err)
	}
}

func options(container *cmd.Container) []kong.Option {
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
			"version": Version,
		},
		kong.Groups(map[string]string{
			"filter": "Filtering",
			"sort":   "Sorting",
			"format": "Formatting",
			"notes":  term.MustStyle("NOTES", style.RuleYellow, style.RuleBold) + "\n" + term.MustStyle("Edit or browse your notes", style.RuleBold),
			"zk":     term.MustStyle("SLIP BOX", style.RuleYellow, style.RuleBold) + "\n" + term.MustStyle("A slip box is a directory containing your notes", style.RuleBold),
		}),
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
	if len(os.Args) > 1 && os.Args[1] != "index" {
		(&cmd.Index{Quiet: true}).Run(container)
	}
}

// runAlias will execute a user alias if the command is one of them.
func runAlias(container *cmd.Container, args []string) (bool, error) {
	runningAlias := os.Getenv("ZK_RUNNING_ALIAS")
	if zk, err := container.OpenZk(); err == nil && len(args) >= 1 {
		for alias, cmdStr := range zk.Config.Aliases {
			if alias == runningAlias || alias != args[0] {
				continue
			}

			// Prevent infinite loop if an alias calls itself.
			os.Setenv("ZK_RUNNING_ALIAS", alias)

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
	}

	return false, nil
}
