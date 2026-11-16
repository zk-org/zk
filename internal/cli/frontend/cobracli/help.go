package cobracli

import (
	"bytes"
	"fmt"
	"go/doc"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zk-org/zk/internal/cli/spec"
	"golang.org/x/term"
)

// Command annotations describing what cobra doesn't model but the help needs.
const (
	// The positional arguments as shown in the usage line, e.g. "[<path> ...]".
	usageArgsAnnotation = "zk_usage_args"
	// One "label\thelp" entry per line, for the "Arguments:" section.
	argsAnnotation = "zk_args"
)

const (
	helpIndent        = 2
	helpColumnPadding = 4
	// Help text is never wrapped wider than this, however wide the terminal
	// is: a line running the full width of a large one is hard to read.
	helpMaxWidth = 100
)

// setHelp installs zk's help renderer on cmd and all its descendants.
//
// Usage is never printed on error: an error is reported as a single
// "zk: error: …" line, and a wall of help text in between would only bury it.
func setHelp(cmd *cobra.Command, groups map[string]string) {
	cmd.InitDefaultHelpFlag()
	if flag := cmd.Flags().Lookup("help"); flag != nil {
		flag.Usage = "Show context-sensitive help."
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) { printHelp(cmd, groups) })
	cmd.SetUsageFunc(func(*cobra.Command) error { return nil })

	for _, child := range cmd.Commands() {
		setHelp(child, groups)
	}
}

func printHelp(cmd *cobra.Command, groups map[string]string) {
	out := cmd.OutOrStdout()
	w := newHelpWriter(out)

	w.Printf("Usage: %s", usageLine(cmd))

	if cmd.Short != "" {
		w.Print("")
		w.Wrap(cmd.Short)
	}
	if cmd.Long != "" {
		w.Print("")
		w.Wrap(cmd.Long)
	}

	writeArguments(w, cmd)
	writeCommands(w, cmd, groups)
	writeFlagGroups(w, cmd, groups)

	// Only the root command lists enough subcommands for the hint to be worth
	// printing.
	if !cmd.HasParent() && cmd.HasAvailableSubCommands() {
		w.Print("")
		w.Printf(`Run "%s <command> --help" for more information on a command.`, cmd.Name())
	}

	_ = w.Write(out)
}

// usageLine returns the command invocation shown after "Usage:", listing the
// required flags and then either the positional arguments or <command>.
func usageLine(cmd *cobra.Command) string {
	usage := cmd.CommandPath()

	if required := requiredFlagsSummary(cmd); required != "" {
		usage += " " + required
	}

	if args := cmd.Annotations[usageArgsAnnotation]; args != "" {
		usage += " " + args
	} else if cmd.HasAvailableSubCommands() {
		usage += " <command>"
	}

	return usage
}

func requiredFlagsSummary(cmd *cobra.Command) string {
	required := []string{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden && isRequired(flag) {
			required = append(required, "--"+flag.Name+placeholderSuffix(flag))
		}
	})
	return strings.Join(required, " ")
}

func isRequired(flag *pflag.Flag) bool {
	values, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]
	return ok && len(values) > 0 && values[0] == "true"
}

func writeArguments(w *helpWriter, cmd *cobra.Command) {
	entries := cmd.Annotations[argsAnnotation]
	if entries == "" {
		return
	}

	rows := [][2]string{}
	for _, entry := range strings.Split(entries, "\n") {
		label, help, _ := strings.Cut(entry, "\t")
		rows = append(rows, [2]string{label, help})
	}

	w.Print("")
	w.Print("Arguments:")
	writeTwoColumns(w.Indent(), rows)
}

func writeCommands(w *helpWriter, cmd *cobra.Command, groups map[string]string) {
	children := cmd.Commands()
	if len(children) == 0 {
		return
	}

	// Group IDs in order of appearance, with "" for the ungrouped commands,
	// which are always listed first.
	ids := []string{}
	byGroup := map[string][]*cobra.Command{}
	for _, child := range children {
		if child.Name() == cobra.ShellCompRequestCmd || child.Name() == cobra.ShellCompNoDescRequestCmd {
			continue
		}
		if _, ok := byGroup[child.GroupID]; !ok {
			ids = append(ids, child.GroupID)
		}
		byGroup[child.GroupID] = append(byGroup[child.GroupID], child)
	}
	ungroupedFirst(ids)

	for _, id := range ids {
		rows := [][2]string{}
		for _, child := range byGroup[id] {
			if child.Hidden {
				continue
			}
			// A subcommand is listed with its full path, e.g. "tag list".
			path := strings.TrimPrefix(child.CommandPath(), cmd.Root().Name()+" ")
			rows = append(rows, [2]string{path, child.Short})
		}

		title, description := "Commands:", ""
		if id != "" {
			title, description, _ = cut(groups[id])
		}

		w.Print("")
		if title != "" {
			w.Wrap(title)
		}
		if description != "" {
			w.Indent().Wrap(description)
			w.Print("")
		}
		writeTwoColumns(w.Indent(), rows)
	}
}

func writeFlagGroups(w *helpWriter, cmd *cobra.Command, groups map[string]string) {
	// The help flag comes first, then the flags shared by every command, then
	// the command's own. That keeps the global flags in the same place in
	// every command's help.
	persistent := cmd.Root().PersistentFlags()

	global := []*pflag.Flag{}
	if help := cmd.Flags().Lookup("help"); help != nil {
		global = append(global, help)
	}
	persistent.VisitAll(func(flag *pflag.Flag) {
		global = append(global, flag)
	})

	ids := []string{""}
	byGroup := map[string][]*pflag.Flag{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" || persistent.Lookup(flag.Name) != nil {
			return
		}
		id := spec.Annotation(flag, spec.GroupAnnotation)
		if _, ok := byGroup[id]; !ok && id != "" {
			ids = append(ids, id)
		}
		byGroup[id] = append(byGroup[id], flag)
	})

	// The global flags and the command's own ungrouped ones share the
	// "Flags:" section, separated by a blank line.
	byGroup[""] = append(global, prependSeparator(byGroup[""])...)

	for _, id := range ids {
		rows := flagRows(byGroup[id])
		if len(rows) == 0 {
			continue
		}

		title := "Flags:"
		if id != "" {
			title, _, _ = cut(groups[id])
		}

		w.Print("")
		if title != "" {
			w.Wrap(title)
		}
		writeTwoColumns(w.Indent(), rows)
	}
}

// prependSeparator inserts a blank separator row before the given flags, or
// returns them untouched when there are none.
func prependSeparator(flags []*pflag.Flag) []*pflag.Flag {
	if len(flags) == 0 {
		return flags
	}
	return append([]*pflag.Flag{nil}, flags...)
}

// flagRows formats a section's flags as two-column rows. A nil flag stands
// for a blank separator row.
func flagRows(flags []*pflag.Flag) [][2]string {
	// Long flags are aligned with the short ones only when the section has
	// any.
	haveShort := false
	for _, flag := range flags {
		if flag != nil && flag.Shorthand != "" && !flag.Hidden {
			haveShort = true
			break
		}
	}

	rows := [][2]string{}
	for _, flag := range flags {
		switch {
		case flag == nil:
			rows = append(rows, [2]string{"", ""})
		case flag.Hidden:
		default:
			rows = append(rows, [2]string{formatFlag(haveShort, flag), flag.Usage})
		}
	}

	// Drop a separator which ended up leading or trailing because every flag
	// around it was hidden.
	for len(rows) > 0 && rows[0][0] == "" {
		rows = rows[1:]
	}
	for len(rows) > 0 && rows[len(rows)-1][0] == "" {
		rows = rows[:len(rows)-1]
	}
	return rows
}

func formatFlag(haveShort bool, flag *pflag.Flag) string {
	var out string
	switch {
	case flag.Shorthand != "":
		out = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
	case haveShort:
		out = fmt.Sprintf("    --%s", flag.Name)
	default:
		out = fmt.Sprintf("--%s", flag.Name)
	}
	return out + placeholderSuffix(flag)
}

func placeholderSuffix(flag *pflag.Flag) string {
	if p := spec.Annotation(flag, spec.PlaceholderAnnotation); p != "" {
		return "=" + p
	}
	return ""
}

// cut splits a group label into its title and description.
func cut(label string) (title string, description string, found bool) {
	return strings.Cut(label, "\n")
}

// ungroupedFirst moves the ungrouped entries to the front, leaving the other
// groups in their order of appearance.
func ungroupedFirst(ids []string) {
	for i, id := range ids {
		if id == "" {
			copy(ids[1:i+1], ids[:i])
			ids[0] = ""
			return
		}
	}
}

// helpWriter accumulates the help text, wrapping it to the terminal width.
type helpWriter struct {
	indent string
	width  int
	lines  *[]string
}

func newHelpWriter(out io.Writer) *helpWriter {
	lines := []string{}
	width := guessWidth(out)
	if width > helpMaxWidth {
		width = helpMaxWidth
	}
	return &helpWriter{width: width, lines: &lines}
}

func (h *helpWriter) Printf(format string, args ...interface{}) {
	h.Print(fmt.Sprintf(format, args...))
}

func (h *helpWriter) Print(text string) {
	*h.lines = append(*h.lines, strings.TrimRight(h.indent+text, " "))
}

// Indent returns a new writer indented by two more characters.
func (h *helpWriter) Indent() *helpWriter {
	return &helpWriter{indent: h.indent + "  ", width: h.width - 2, lines: h.lines}
}

func (h *helpWriter) Wrap(text string) {
	buf := bytes.NewBuffer(nil)
	toText(buf, strings.TrimSpace(text), "", "    ", h.width)
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		h.Print(line)
	}
}

func (h *helpWriter) Write(w io.Writer) error {
	for _, line := range *h.lines {
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
	}
	return nil
}

func writeTwoColumns(w *helpWriter, rows [][2]string) {
	maxLeft := 375 * w.width / 1000
	if maxLeft < 30 {
		maxLeft = 30
	}

	// Find the width of the first column, ignoring the entries too wide to
	// leave room for their help text.
	leftSize := 0
	for _, row := range rows {
		if c := len(row[0]); c > leftSize && c < maxLeft {
			leftSize = c
		}
	}

	offset := strings.Repeat(" ", leftSize+helpColumnPadding)

	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		toText(buf, row[1], "", strings.Repeat(" ", helpIndent), w.width-leftSize-helpColumnPadding)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")

		line := fmt.Sprintf("%-*s", leftSize, row[0])
		if len(row[0]) < maxLeft {
			line += fmt.Sprintf("%*s%s", helpColumnPadding, "", lines[0])
			lines = lines[1:]
		}
		w.Print(line)
		for _, line := range lines {
			w.Printf("%s%s", offset, line)
		}
	}
}

// toText formats text the way go/doc does, which is how zk's help has always
// been wrapped.
func toText(w io.Writer, text string, indent string, preIndent string, width int) {
	doc.ToText(w, text, indent, preIndent, width) //nolint:staticcheck
}

// guessWidth returns the width the help is wrapped to, honoring COLUMNS as
// POSIX requires.
func guessWidth(w io.Writer) int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if width, err := strconv.Atoi(cols); err == nil && width > 0 {
			return width
		}
	}
	if f, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(int(f.Fd())); err == nil && width > 0 {
			return width
		}
	}
	return 80
}
