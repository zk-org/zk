// Package cobracli drives zk's commands from a command line, using Cobra.
//
// It reads the command tree from internal/cli/spec and presents it, which is
// what makes shell completion possible: Cobra knows every command, every flag
// and the values they accept, and can hand a shell a script which asks zk
// itself what to suggest.
package cobracli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zk-org/zk/internal/cli/spec"
)

// Run presents the app's commands on the command line, and runs the one the
// given arguments select.
//
// before runs once the selected command's flags and arguments are filled in
// and before the command itself, which is where the caller gets to act on the
// global flags and prepare the notebook.
func Run(app *spec.App, args []string, before spec.Hook) error {
	b := &builder{app: app, before: before, specs: map[*cobra.Command]*spec.Command{}}
	root := b.build()

	root.SetArgs(withDefaultChild(root, b.specs, args))
	_, err := root.ExecuteC()
	return err
}

type builder struct {
	app    *spec.App
	before spec.Hook
	// specs maps a Cobra command back to the spec it was built from.
	specs map[*cobra.Command]*spec.Command
}

func (b *builder) build() *cobra.Command {
	// zk lists its commands in the order they are declared, which reads far
	// better than alphabetically: `init` comes before `index`, and `new`
	// before `list`.
	cobra.EnableCommandSorting = false

	root := &cobra.Command{
		Use: b.app.Name,
		// `zk` on its own prints its help, which is the most useful thing it
		// can do without a command.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unexpected argument %s", args[0])
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Required flags and value restrictions are checked here rather than left
	// to Cobra, which reports them in its own words, and before the hook so
	// that a bad command line is reported before any work is done.
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// Suggesting a notebook's notes and tags needs an up-to-date index
		// just as much as listing them does. A failure here is not worth
		// reporting: it would be printed among the suggestions.
		if isCompletionRequest(cmd) {
			if b.before != nil {
				_ = b.before(&spec.Command{Name: cmd.Name()})
			}
			return nil
		}

		command, ok := b.specs[cmd]
		if !ok {
			return nil
		}
		if err := spec.ValidateFlags(command.Flags, func(name string) bool {
			flag := cmd.Flags().Lookup(name)
			return flag != nil && flag.Changed
		}); err != nil {
			return err
		}
		if b.before != nil {
			return b.before(command)
		}
		return nil
	}

	groups := map[string]string{}
	for _, group := range b.app.CommandGroups {
		groups[group.ID] = label(group)
		root.AddGroup(&cobra.Group{ID: group.ID, Title: label(group)})
	}
	for _, group := range b.app.FlagGroups {
		groups[group.ID] = label(group)
	}

	spec.BindFlags(root.PersistentFlags(), b.app.Flags)

	for _, command := range b.app.Commands {
		root.AddCommand(b.command(command))
	}

	// `--version` is accepted but, like zk's other diagnostic flags, kept out
	// of the help.
	root.Version = strings.TrimPrefix(b.app.Version, "v")
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	root.InitDefaultVersionFlag()
	if flag := root.Flags().Lookup("version"); flag != nil {
		flag.Hidden = true
		flag.Usage = "Print zk version."
	}

	addCompletionCommand(root)
	setHelp(root, groups)
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return spec.FlagError(err)
	})

	return root
}

func (b *builder) command(command *spec.Command) *cobra.Command {
	c := &cobra.Command{
		Use:     command.Name,
		Short:   command.Short,
		Long:    command.Long,
		GroupID: command.Group,
		Hidden:  command.Hidden,
		Args:    argsValidator(command),
	}
	b.specs[c] = command

	if command.Run != nil {
		c.RunE = func(_ *cobra.Command, args []string) error {
			if err := command.BindArgs(args); err != nil {
				return err
			}
			return command.Run()
		}
	}

	spec.BindFlags(c.Flags(), command.Flags)
	for _, flag := range command.Flags {
		if flag.Required {
			_ = c.MarkFlagRequired(flag.Name)
		}
	}

	if len(command.Args) > 0 {
		labels := make([]string, 0, len(command.Args))
		for _, arg := range command.Args {
			labels = append(labels, arg.Label()+"\t"+arg.Help)
		}
		c.Annotations = map[string]string{
			usageArgsAnnotation: command.UsageArgs(),
			argsAnnotation:      strings.Join(labels, "\n"),
		}
	}

	for _, child := range command.Children {
		c.AddCommand(b.command(child))
	}

	registerCompletions(c, command)

	return c
}

// argsValidator rejects an unexpected positional argument. The binding
// reports a missing one, which it can name.
func argsValidator(command *spec.Command) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		max, variadic := command.MaxArgs()
		if variadic || len(args) <= max {
			return nil
		}
		return fmt.Errorf("unexpected argument %s", args[max])
	}
}

// withDefaultChild returns args with a parent command's default subcommand
// inserted, so that `zk tag --format json` runs `zk tag list --format json`.
func withDefaultChild(root *cobra.Command, specs map[*cobra.Command]*spec.Command, args []string) []string {
	target, rest, err := root.Find(args)
	if err != nil || target == root {
		return args
	}
	command, ok := specs[target]
	if !ok || command.DefaultChild == "" {
		return args
	}
	// `zk tag --help` documents the parent command itself.
	for _, arg := range rest {
		if arg == "-h" || arg == "--help" {
			return args
		}
	}

	for i, arg := range args {
		if arg == target.Name() {
			out := append([]string{}, args[:i+1]...)
			out = append(out, command.DefaultChild)
			return append(out, args[i+1:]...)
		}
	}
	return args
}

// label joins a group's title and description the way the help renderer reads
// them back.
func label(group spec.Group) string {
	if group.Description == "" {
		return group.Title
	}
	return group.Title + "\n" + group.Description
}
