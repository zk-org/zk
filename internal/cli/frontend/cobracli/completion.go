package cobracli

import (
	"github.com/spf13/cobra"
	"github.com/zk-org/zk/internal/cli/spec"
)

// addCompletionCommand makes `zk completion <shell>` available, which is the
// whole point of driving zk with Cobra: the script it prints asks zk itself
// what to suggest, so completions follow the commands instead of drifting
// from them.
func addCompletionCommand(root *cobra.Command) {
	root.InitDefaultCompletionCmd()
	root.CompletionOptions.DisableDescriptions = false

	// `zk help <command>` works, but is left out of the command listing:
	// zk documents `zk <command> --help` instead, and has always done so.
	root.InitDefaultHelpCmd()
	if help := findCommand(root, "help"); help != nil {
		help.Hidden = true
		help.Short = "Show the help of a command."
	}

	if cmd := findCommand(root, "completion"); cmd != nil {
		cmd.Short = "Print a shell completion script."
		cmd.Long = "Print the script which teaches your shell to complete zk's commands and flags, and, inside a notebook, its notes and tags.\n\nRun `zk completion <shell> --help` to see where your shell expects it."
		for _, shell := range cmd.Commands() {
			shell.Short = "Print the " + shell.Name() + " completion script."
		}
	}
}

// isCompletionRequest reports whether the command is the hidden one a shell
// runs to ask zk for suggestions.
func isCompletionRequest(cmd *cobra.Command) bool {
	return cmd.Name() == cobra.ShellCompRequestCmd || cmd.Name() == cobra.ShellCompNoDescRequestCmd
}

func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// registerCompletions wires the values a command's spec says it accepts into
// Cobra's completion machinery.
func registerCompletions(c *cobra.Command, command *spec.Command) {
	for _, flag := range command.Flags {
		switch {
		case flag.Complete != nil:
			complete := flag.Complete
			_ = c.RegisterFlagCompletionFunc(flag.Name, func(_ *cobra.Command, _ []string, prefix string) ([]string, cobra.ShellCompDirective) {
				return complete(prefix), cobra.ShellCompDirectiveNoFileComp
			})
		case flag.CompleteFiles:
			// Cobra falls back to file names on its own.
		case flag.Kind() != spec.KindBool:
			_ = c.RegisterFlagCompletionFunc(flag.Name, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return nil, cobra.ShellCompDirectiveNoFileComp
			})
		}
	}

	if len(command.Args) == 0 {
		if len(command.Children) == 0 {
			c.ValidArgsFunction = noCompletion
		}
		return
	}

	// Positional arguments are completed from the last one the command
	// accepts, as the earlier ones are already filled in.
	c.ValidArgsFunction = func(_ *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
		arg := command.Args[min(len(args), len(command.Args)-1)]
		switch {
		case arg.Complete != nil:
			return arg.Complete(prefix), cobra.ShellCompDirectiveNoFileComp
		case arg.CompleteFiles:
			return nil, cobra.ShellCompDirectiveDefault
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

func noCompletion(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
