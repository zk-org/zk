package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/strings"
)

// FormatFlags holds the output formatting options shared by commands that print lists of items.
type FormatFlags struct {
	Format     string `group:format short:f placeholder:TEMPLATE help:"Pretty print the list using a custom template or one of the predefined formats."`
	Header     string `group:format help:"Arbitrary text printed at the start of the list."`
	Footer     string "group:format default:\n help:\"Arbitrary text printed at the end of the list.\""
	Delimiter  string "group:format short:d default:\n help:\"Print items delimited by the given separator.\""
	Delimiter0 bool   "group:format short:0 name:delimiter0 help:\"Print items delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.\""
	NoPager    bool   `group:format short:P help:"Do not pipe output into a pager."`
	Quiet      bool   `group:format short:q help:"Do not print the total number of items found."`
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
