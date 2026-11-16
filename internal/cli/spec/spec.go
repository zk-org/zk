// Package spec describes zk's commands, flags and arguments independently of
// the library used to read them from a command line.
//
// It is the contract between the execution layer — the commands themselves,
// which only need their parameters filled in and a notebook to work on — and
// the interaction layers built on top of it. Today those are the Cobra
// front-end, which zk is built with by default, and the Kong one it kept for
// reference. Driving zk some other way, over a socket or an RPC, means
// reading the same specs rather than restating every flag.
//
// Nothing in here knows about any of those libraries, so a front-end is free
// to present a command however it likes, as long as it fills in the targets
// and calls Run.
package spec

import (
	"fmt"
	"strings"
)

// Kind is the type of value a flag or an argument holds.
type Kind int

const (
	KindBool Kind = iota + 1
	KindInt
	KindString
	KindStringList
	KindStringMap
)

// Target is a pointer to the value a flag or an argument fills in. It is one
// of *bool, *int, *string, *[]string or *map[string]string.
type Target any

// KindOf returns the kind of value the target holds. It panics on a target
// type zk doesn't use, which is a programming error in a command's spec.
func KindOf(target Target) Kind {
	switch target.(type) {
	case *bool:
		return KindBool
	case *int:
		return KindInt
	case *string:
		return KindString
	case *[]string:
		return KindStringList
	case *map[string]string:
		return KindStringMap
	default:
		panic(fmt.Sprintf("spec: unsupported target type %T", target))
	}
}

// Flag describes a command line option.
type Flag struct {
	// Name is the long name, without the leading dashes, e.g. "notebook-dir".
	Name string
	// Short is the one letter alias, without the leading dash, if any.
	Short string
	Help  string
	// Group is the help section the flag is listed under, e.g. GroupFilter.
	// Ungrouped flags are listed with the global ones.
	Group string
	// Placeholder stands for the flag's value in the help, e.g. the PATH in
	// `--notebook-dir=PATH`. When empty, one is derived from the flag's kind.
	Placeholder string
	// Hidden keeps the flag out of the help, but still accepted.
	Hidden bool
	// Required reports the flag as missing when it is not set.
	Required bool
	// OneOf restricts the accepted values, when set.
	OneOf []string
	// Default is the value the target holds when the flag is not set. It must
	// match the target's kind.
	Default any
	Target  Target
	// Complete suggests values for the flag, for a front-end able to offer
	// them, such as a shell completion script. A nil Complete means zk has
	// nothing to suggest; use CompleteFiles to ask for file names instead.
	Complete Completer
	// CompleteFiles asks the front-end to suggest file names.
	CompleteFiles bool
}

// Kind returns the kind of value the flag holds.
func (f Flag) Kind() Kind { return KindOf(f.Target) }

// DefaultString returns the flag's default value as a string, which is how
// the parsers and the help want it.
func (f Flag) DefaultString() string {
	if f.Default == nil {
		return ""
	}
	switch value := f.Default.(type) {
	case bool:
		if value {
			return "true"
		}
		return "false"
	case int:
		if value == 0 {
			return ""
		}
		return fmt.Sprint(value)
	case string:
		return value
	default:
		panic(fmt.Sprintf("spec: unsupported default type %T for --%s", f.Default, f.Name))
	}
}

// Arg describes a positional argument.
type Arg struct {
	// Name is how the argument is referred to in the help, e.g. "path" is
	// shown as <path>.
	Name     string
	Help     string
	Optional bool
	// Variadic collects every remaining argument.
	Variadic bool
	// Default is the value used when the argument is left out.
	Default string
	// Path expands `~` and environment variables in the value, and makes it
	// absolute.
	Path   bool
	Target Target
	// Complete suggests values for the argument. See Flag.Complete.
	Complete      Completer
	CompleteFiles bool
}

// Label returns how the argument is shown in the help, e.g. "[<path> ...]".
func (a Arg) Label() string {
	label := "<" + a.Name + ">"
	if a.Variadic {
		label += " ..."
	}
	if a.Optional {
		label = "[" + label + "]"
	}
	return label
}

// Hook runs after a front-end has filled in a command's parameters and
// before the command itself runs.
type Hook func(cmd *Command) error

// Completer suggests the values a flag or an argument accepts, given what
// the user has typed so far.
type Completer func(prefix string) []string

// Values returns a Completer suggesting a fixed set of values.
func Values(values ...string) Completer {
	return func(prefix string) []string {
		out := []string{}
		for _, value := range values {
			if strings.HasPrefix(value, prefix) {
				out = append(out, value)
			}
		}
		return out
	}
}

// Group is a section the help lists commands or flags under.
type Group struct {
	ID          string
	Title       string
	Description string
}

// Flag group identifiers.
const (
	GroupFilter = "filter"
	GroupSort   = "sort"
	GroupFormat = "format"
)

// Command describes a command and, through Run, how to execute it.
type Command struct {
	Name   string
	Short  string
	Long   string
	Group  string
	Hidden bool
	Args   []Arg
	Flags  []Flag
	// Children are the subcommands, e.g. `list` under `tag`.
	Children []*Command
	// DefaultChild names the subcommand to run when the command is invoked
	// without one, e.g. `zk tag` runs `zk tag list`.
	DefaultChild string
	// SkipAutoIndex keeps the runner from indexing the notebook before the
	// command runs.
	SkipAutoIndex bool
	// Run executes the command, once its flags and arguments have been filled
	// in. It is nil for a command which only groups subcommands.
	Run func() error
}

// Lookup returns the child command with the given name.
func (c *Command) Lookup(name string) *Command {
	for _, child := range c.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// UsageArgs returns the positional arguments as shown in the usage line, e.g.
// "[<path> ...]". Nested optional arguments have their brackets closed at the
// end, the way `[<a> [<b>]]` reads.
func (c *Command) UsageArgs() string {
	labels := []string{}
	optional := 0
	for _, arg := range c.Args {
		label := arg.Label()
		if arg.Optional {
			optional++
			label = strings.TrimRight(label, "]")
		}
		labels = append(labels, label)
	}
	if len(labels) == 0 {
		return ""
	}
	return strings.Join(labels, " ") + strings.Repeat("]", optional)
}

// App is the whole command tree, with what the front-ends need to present it.
type App struct {
	Name    string
	Version string
	// Flags available to every command.
	Flags []Flag
	// Commands in the order they are listed in the help.
	Commands []*Command
	// CommandGroups and FlagGroups title the help sections, in the order they
	// appear.
	CommandGroups []Group
	FlagGroups    []Group
}

// Lookup returns the top-level command with the given name.
func (a *App) Lookup(name string) *Command {
	for _, cmd := range a.Commands {
		if cmd.Name == name {
			return cmd
		}
	}
	return nil
}

// FlagGroupTitle returns the help section title for a flag group.
func (a *App) FlagGroupTitle(id string) string {
	for _, group := range a.FlagGroups {
		if group.ID == id {
			return group.Title
		}
	}
	return ""
}
