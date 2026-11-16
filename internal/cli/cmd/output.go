package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/cli/spec"
	"github.com/zk-org/zk/internal/util/strings"
)

// FormatFlags holds the output formatting options shared by commands that print lists of items.
type FormatFlags struct {
	Format     string
	Header     string
	Footer     string
	Delimiter  string
	Delimiter0 bool
	NoPager    bool
	Quiet      bool
}

// Prepare normalises the formatting flags and rejects mutually exclusive options.
func (f *FormatFlags) Prepare(jsonHeader, jsonFooter string, supportsJSONL bool) error {
	f.Header = strings.ExpandWhitespaceLiterals(f.Header)
	f.Footer = strings.ExpandWhitespaceLiterals(f.Footer)
	f.Delimiter = strings.ExpandWhitespaceLiterals(f.Delimiter)

	if f.Delimiter0 {
		if f.Delimiter != "\n" {
			return errors.New("--delimiter and --delimiter0 can't be used together")
		}
		if f.Header != "" {
			return errors.New("--header and --delimiter0 can't be used together")
		}
		if f.Footer != "\n" {
			return errors.New("--footer and --delimiter0 can't be used together")
		}

		f.Delimiter = "\x00"
		f.Footer = "\x00"
	}

	if f.Format == "json" || (supportsJSONL && f.Format == "jsonl") {
		if f.Header != "" {
			return errors.New("--header can't be used with JSON format")
		}
		if f.Footer != "\n" {
			return errors.New("--footer can't be used with JSON format")
		}
		if f.Delimiter != "\n" {
			return errors.New("--delimiter can't be used with JSON format")
		}

		switch f.Format {
		case "json":
			f.Delimiter = ","
			f.Header = jsonHeader
			f.Footer = jsonFooter
		case "jsonl":
			// > The last character in the file may be a line separator, and it
			// > will be treated the same as if there was no line separator
			// > present.
			// > https://jsonlines.org/
			f.Footer = "\n"
		}
	}

	return nil
}

// Paginate prints the given number of items using the configured header,
// delimiter and footer, piping the result through the user's pager if needed.
// The render callback is responsible for printing the item at index i.
func (f FormatFlags) Paginate(container *cli.Container, count int, render func(out io.Writer, i int) error) error {
	if count == 0 {
		return nil
	}

	return container.Paginate(f.NoPager, func(out io.Writer) error {
		if f.Header != "" {
			fmt.Fprint(out, f.Header)
		}
		for i := 0; i < count; i++ {
			if i > 0 {
				fmt.Fprint(out, f.Delimiter)
			}
			if err := render(out, i); err != nil {
				return err
			}
		}
		if f.Footer != "" {
			fmt.Fprint(out, f.Footer)
		}
		return nil
	})
}

// Template resolves the configured format to a handlebars template, falling to the given default.
func (f FormatFlags) Template(defaultFormat string, predefined map[string]string) string {
	format := f.Format
	if format == "" {
		format = defaultFormat
	}

	if templ, ok := predefined[format]; ok {
		return templ
	}

	return strings.ExpandWhitespaceLiterals(format)
}

// PrintFound prints the total number of found items on the container's
// standard error, unless the quiet flag is set. The noun must already be
// pluralised for the count.
func (f FormatFlags) PrintFound(container *cli.Container, count int, noun string) {
	if !f.Quiet {
		fmt.Fprintf(container.Stderr, "\nFound %d %s\n", count, noun)
	}
}

// FormatFlagsSpec is what one command's output flags differ in: the formats it
// knows, and the word its help uses for the things it prints.
type FormatFlagsSpec struct {
	// Noun names what is being printed, as the help text says it: "notes" for
	// `zk list`, "tags" for `zk tag list`.
	Noun string
	// FormatHelp is the help of --format, which names that command's
	// predefined formats.
	FormatHelp string
	// Formats are the names a shell offers for --format.
	Formats []string
	// Extra are further flags belonging to the formatting group, inserted
	// after --format: `zk list` puts --link there.
	Extra []spec.Flag
}

// Spec describes the shared output flags, so that a command declares them in
// one line rather than repeating all seven.
//
// It is the counterpart of the struct above: the fields are shared, so their
// descriptions are too, and a change to the wording reaches every command
// which prints a list.
func (f *FormatFlags) Spec(s FormatFlagsSpec) []spec.Flag {
	flags := []spec.Flag{
		{Name: "format", Short: "f", Group: spec.GroupFormat, Placeholder: "TEMPLATE", Target: &f.Format,
			Help:     s.FormatHelp,
			Complete: spec.Values(s.Formats...)},
	}
	flags = append(flags, s.Extra...)

	return append(flags,
		spec.Flag{Name: "header", Group: spec.GroupFormat, Target: &f.Header,
			Help: "Arbitrary text printed at the start of the list."},
		// The default of --footer is the two characters `\n`, which
		// ExpandWhitespaceLiterals turns into a newline, while --delimiter
		// defaults to a real one. Both end up as a newline, but the help shows
		// them differently, as it always has.
		spec.Flag{Name: "footer", Group: spec.GroupFormat, Default: `\n`, Target: &f.Footer,
			Help: "Arbitrary text printed at the end of the list."},
		spec.Flag{Name: "delimiter", Short: "d", Group: spec.GroupFormat, Default: "\n", Target: &f.Delimiter,
			Help: fmt.Sprintf("Print %s delimited by the given separator.", s.Noun)},
		spec.Flag{Name: "delimiter0", Short: "0", Group: spec.GroupFormat, Target: &f.Delimiter0,
			Help: fmt.Sprintf("Print %s delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.", s.Noun)},
		spec.Flag{Name: "no-pager", Short: "P", Group: spec.GroupFormat, Target: &f.NoPager,
			Help: "Do not pipe output into a pager."},
		spec.Flag{Name: "quiet", Short: "q", Group: spec.GroupFormat, Target: &f.Quiet,
			Help: fmt.Sprintf("Do not print the total number of %s found.", s.Noun)},
	)
}
