package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/core"
)

// aliasShortHelpMax caps how much of an alias's command stands in for a
// description it doesn't have. A long shell pipeline would push every other
// column of the help off the screen.
const aliasShortHelpMax = 60

// reservedAliasNames are the names an alias is not presented under, on top of
// zk's own commands. They are the ones the front-end adds for itself.
var reservedAliasNames = map[string]bool{
	"completion": true,
	"help":       true,
	"__complete": true,
}

// aliasCommands describes the aliases declared in the notebook's
// configuration, so that a front-end can present them alongside zk's own
// commands: they show up in `zk --help` with their description, and a shell
// suggests them as it does every other command.
//
// The aliases are not executed from here. They are run before the command
// line is parsed, by runAlias, because an alias may override one of zk's
// commands and has to win. An alias named after an existing command is left
// out for the same reason: zk would then have two commands by that name, and
// only the alias would ever run.
func aliasCommands(container *cli.Container, app *spec.App) []*spec.Command {
	names := make([]string, 0, len(container.Config.Aliases))
	for name := range container.Config.Aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	commands := []*spec.Command{}
	for _, name := range names {
		if app.Lookup(name) != nil || reservedAliasNames[name] {
			continue
		}
		alias := container.Config.Aliases[name]

		args := []string{}
		commands = append(commands, &spec.Command{
			Name:  name,
			Short: aliasShortHelp(alias),
			Long:  "Command alias running:\n\n    " + alias.Command,
			Group: GroupAliases,
			// The arguments belong to the shell command the alias runs, and
			// are expanded by it with `$@` or `$*`. zk neither parses them
			// nor indexes the notebook on their behalf.
			PassThroughArgs: true,
			SkipAutoIndex:   true,
			Args: []spec.Arg{
				{Name: "args", Optional: true, Variadic: true, Target: &args, CompleteFiles: true,
					Help: "Arguments passed to the aliased command."},
			},
			Run: func() error {
				// Reaching here means runAlias declined to run the alias.
				if os.Getenv("ZK_RUNNING_ALIAS") == name {
					return fmt.Errorf("alias `%s` cannot call itself", name)
				}
				return fmt.Errorf("alias `%s` must be the first argument, as in `zk %s`", name, name)
			},
		})
	}
	return commands
}

// aliasShortHelp returns the one line describing an alias. An alias without a
// description of its own is shown by the command it runs, which is better
// than nothing at all.
func aliasShortHelp(alias core.Alias) string {
	if alias.Description != "" {
		return singleLine(alias.Description, 0)
	}
	return singleLine(alias.Command, aliasShortHelpMax)
}

// singleLine collapses whitespace and, when max is not zero, truncates. Help
// and completions are line-based, and a shell reads a command's description
// up to the first tab.
func singleLine(text string, max int) string {
	text = strings.Join(strings.Fields(text), " ")
	// Counted and cut in runes: a command may well hold characters which take
	// several bytes, and cutting one in half would leave invalid UTF-8.
	runes := []rune(text)
	if max > 0 && len(runes) > max {
		return string(runes[:max-1]) + "…"
	}
	return text
}
