// Package kongcli drives zk's commands from a command line, using Kong.
//
// zk was built with Kong before it moved to Cobra, for the shell completion
// Cobra can generate. This front-end is kept to show that the command tree in
// internal/cli/spec is not tied to the library presenting it, and to have
// something to compare against: built with `-tags zk_kong`, zk accepts the
// same commands and flags, read by a different parser.
//
// Kong builds its grammar from struct tags, so the specs are turned into a
// struct type at run time, parsed, and the values copied back into the
// targets the commands read.
package kongcli

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/zk-org/zk/internal/cli/spec"
)

// Run presents the app's commands on the command line, and runs the one the
// given arguments select.
//
// before runs once the selected command's flags and arguments are filled in
// and before the command itself, which is where the caller gets to act on the
// global flags and prepare the notebook.
func Run(app *spec.App, args []string, before spec.Hook) error {
	g := newGrammar(app)

	parser, err := kong.New(g.value.Interface(), g.options(app)...)
	if err != nil {
		return err
	}
	setDetails(parser.Model.Node, app.Commands)

	// Kong reports a missing required flag and a value outside an enum
	// itself, in the same words zk has always used.
	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}

	command := g.selected(ctx.Command())

	// Kong fills in its own struct, so the values have to be handed over to
	// the targets the commands were declared with.
	g.commit(command)

	if command == nil || command.Run == nil {
		// `zk` on its own, which resolves to the hidden default command.
		return showHelp(g, app)
	}

	if before != nil {
		if err := before(command); err != nil {
			return err
		}
	}
	return command.Run()
}

// showHelp prints the help the way `zk --help` would.
func showHelp(g *grammar, app *spec.App) error {
	parser, err := kong.New(g.value.Interface(), g.options(app)...)
	if err != nil {
		return err
	}
	setDetails(parser.Model.Node, app.Commands)
	ctx, err := parser.Parse([]string{"--help"})
	if err != nil {
		return err
	}
	return ctx.Run()
}

// grammar is the struct Kong parses into, built from the specs.
type grammar struct {
	value reflect.Value
	// bindings records where each spec target ended up in the struct, so the
	// parsed values can be copied back into it.
	bindings []binding
	// commands maps a command path, e.g. "tag list", to its spec.
	commands map[string]*spec.Command
}

// binding is the chain of struct field indices leading to a spec target.
type binding struct {
	path   []int
	target spec.Target
	// owner is the command the flag or argument belongs to, or nil for the
	// global flags. Only the selected command's values are handed over, so
	// that a command which was not run cannot overwrite them.
	owner *spec.Command
}

func newGrammar(app *spec.App) *grammar {
	g := &grammar{commands: map[string]*spec.Command{}}

	fields := []reflect.StructField{}
	for _, command := range app.Commands {
		fields = append(fields, g.commandField(command, path(nil, len(fields)), []string{command.Name}, false))
	}
	for _, flag := range app.Flags {
		fields = append(fields, g.flagField(flag, path(nil, len(fields)), nil))
	}

	// Kong needs a command to fall back on, as zk prints its help when run
	// without one.
	fields = append(fields, reflect.StructField{
		Name: "FShowHelp",
		Type: reflect.StructOf(nil),
		Tag:  `kong:"cmd,name='show-help',hidden,default='1'"`,
	})
	fields = append(fields, reflect.StructField{
		Name: "FVersion",
		Type: reflect.TypeOf(kong.VersionFlag(false)),
		Tag:  `kong:"name='version',hidden,help='Print zk version.'"`,
	})

	g.value = reflect.New(reflect.StructOf(fields))
	return g
}

// commandField returns the struct field standing for a command. Its type
// holds the command's subcommands, arguments and flags.
func (g *grammar) commandField(command *spec.Command, at []int, names []string, isDefault bool) reflect.StructField {
	g.commands[strings.Join(names, " ")] = command

	fields := []reflect.StructField{}
	for _, child := range command.Children {
		childNames := append(append([]string{}, names...), child.Name)
		fields = append(fields, g.commandField(child, path(at, len(fields)), childNames, child.Name == command.DefaultChild))
	}
	for _, arg := range command.Args {
		fields = append(fields, g.argField(arg, path(at, len(fields)), command))
	}
	for _, flag := range command.Flags {
		fields = append(fields, g.flagField(flag, path(at, len(fields)), command))
	}

	tags := []string{"cmd", "name='" + command.Name + "'"}
	switch {
	case command.Group != "":
		tags = append(tags, "group='"+command.Group+"'")
	case len(names) > 1:
		// A subcommand is listed under a plain "Commands:" heading.
		tags = append(tags, "group='cmd'")
	}
	if command.Hidden {
		tags = append(tags, "hidden")
	}
	if command.Short != "" {
		tags = append(tags, "help='"+escape(command.Short)+"'")
	}
	if isDefault {
		// Kong marks the default subcommand itself, not its parent.
		// "withargs" lets `zk tag --format json` reach `zk tag list`.
		tags = append(tags, "default='withargs'")
	}

	return reflect.StructField{
		Name: fieldName(command.Name),
		Type: reflect.StructOf(fields),
		Tag:  tag(tags),
	}
}

func (g *grammar) flagField(flag spec.Flag, at []int, owner *spec.Command) reflect.StructField {
	tags := []string{"name='" + flag.Name + "'"}
	if flag.Short != "" {
		tags = append(tags, "short='"+flag.Short+"'")
	}
	if flag.Group != "" {
		tags = append(tags, "group='"+flag.Group+"'")
	}
	if flag.Placeholder != "" {
		tags = append(tags, "placeholder='"+flag.Placeholder+"'")
	}
	if flag.Hidden {
		tags = append(tags, "hidden")
	}
	if flag.Required {
		tags = append(tags, "required")
	}
	if len(flag.OneOf) > 0 {
		tags = append(tags, "enum='"+strings.Join(flag.OneOf, ",")+"'")
	}
	if def := flag.DefaultString(); def != "" {
		tags = append(tags, "default='"+escapeDefault(def)+"'")
	}
	if flag.Kind() == spec.KindStringMap {
		tags = append(tags, "mapsep=','")
	}
	if flag.Help != "" {
		tags = append(tags, "help='"+escape(flag.Help)+"'")
	}

	g.bindings = append(g.bindings, binding{path: at, target: flag.Target, owner: owner})

	return reflect.StructField{
		Name: fieldName(flag.Name),
		Type: reflect.TypeOf(flag.Target).Elem(),
		Tag:  tag(tags),
	}
}

func (g *grammar) argField(arg spec.Arg, at []int, owner *spec.Command) reflect.StructField {
	tags := []string{"arg", "name='" + arg.Name + "'"}
	if arg.Optional {
		tags = append(tags, "optional")
	}
	if arg.Default != "" {
		tags = append(tags, "default='"+escapeDefault(arg.Default)+"'")
	}
	if arg.Path {
		tags = append(tags, "type='path'")
	}
	if arg.Help != "" {
		tags = append(tags, "help='"+escape(arg.Help)+"'")
	}

	g.bindings = append(g.bindings, binding{path: at, target: arg.Target, owner: owner})

	return reflect.StructField{
		Name: fieldName(arg.Name),
		Type: reflect.TypeOf(arg.Target).Elem(),
		Tag:  tag(tags),
	}
}

func (g *grammar) options(app *spec.App) []kong.Option {
	groups := map[string]string{"cmd": "Commands:"}
	for _, group := range app.CommandGroups {
		groups[group.ID] = label(group)
	}
	for _, group := range app.FlagGroups {
		groups[group.ID] = label(group)
	}

	return []kong.Option{
		kong.Name(app.Name),
		kong.HelpOptions{
			Compact:             true,
			FlagsLast:           true,
			WrapUpperBound:      100,
			NoExpandSubcommands: true,
		},
		kong.Vars{"version": app.Name + " " + strings.TrimPrefix(app.Version, "v")},
		kong.Groups(groups),
	}
}

// selected returns the spec of the command Kong resolved. Kong renders a
// command taking positional arguments as e.g. "open <uri>", so only the
// leading words are command names.
func (g *grammar) selected(resolved string) *spec.Command {
	names := []string{}
	for _, word := range strings.Fields(resolved) {
		if strings.HasPrefix(word, "<") || strings.HasPrefix(word, "[") {
			break
		}
		names = append(names, word)
	}
	return g.commands[strings.Join(names, " ")]
}

// commit copies the parsed values into the targets the selected command
// reads, along with the global flags.
func (g *grammar) commit(selected *spec.Command) {
	for _, bound := range g.bindings {
		if bound.owner != nil && bound.owner != selected {
			continue
		}
		value := g.value.Elem()
		for _, index := range bound.path {
			value = value.Field(index)
		}
		reflect.ValueOf(bound.target).Elem().Set(value)
	}
}

// path returns the field path to a new field at the given index, without
// sharing the parent's backing array.
func path(parent []int, index int) []int {
	out := make([]int, 0, len(parent)+1)
	out = append(out, parent...)
	return append(out, index)
}

// fieldName turns a command, flag or argument name into the exported Go field
// name reflect.StructOf requires.
func fieldName(name string) string {
	out := "F"
	for _, part := range strings.Split(name, "-") {
		if part != "" {
			out += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return out
}

// tag assembles a struct tag. The value has to be a valid Go quoted string,
// or reflect.StructTag.Lookup silently reports no tag at all and every item
// in it is lost.
func tag(tags []string) reflect.StructTag {
	return reflect.StructTag("kong:" + strconv.Quote(strings.Join(tags, ",")))
}

// escape encodes a value for a Kong tag, which reads it between single
// quotes. Only the quote itself needs escaping; a comma inside the quotes is
// not a separator.
func escape(value string) string {
	return strings.ReplaceAll(value, `'`, `\'`)
}

// escapeDefault encodes a default value for a Kong tag. Kong drops a real
// newline from a tag but turns the two characters `\n` back into one, which
// is what `--delimiter` relies on.
func escapeDefault(value string) string {
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, "\t", `\t`)
	return escape(value)
}

// setDetails fills in the long help of each command. Kong takes it from a
// Help() method on the command's struct, which a type built by reflection
// cannot have, so it is set on the parsed model instead.
func setDetails(node *kong.Node, commands []*spec.Command) {
	for _, command := range commands {
		for _, child := range node.Children {
			if child.Name != command.Name {
				continue
			}
			child.Detail = command.Long
			setDetails(child, command.Children)
		}
	}
}

func label(group spec.Group) string {
	if group.Description == "" {
		return group.Title
	}
	return group.Title + "\n" + group.Description
}
