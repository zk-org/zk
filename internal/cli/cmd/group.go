package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/errors"
	"github.com/zk-org/zk/internal/util/strings"
)

// Group manages the note groups in the notebook.
type Group struct {
	List GroupList `cmd group:"cmd" default:"withargs" help:"List all the note groups."`
}

// GroupList lists all the note groups.
type GroupList struct {
	Header     string `group:format                                help:"Arbitrary text printed at the start of the list."`
	Footer     string `group:format default:\n                     help:"Arbitrary text printed at the end of the list."`
	Delimiter  string "group:format short:d default:\n             help:\"Print groups delimited by the given separator.\""
	Delimiter0 bool   "group:format short:0 name:delimiter0        help:\"Print groups delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.\""
	NoPager    bool   `group:format short:P help:"Do not pipe output into a pager."`
	Quiet      bool   `group:format short:q help:"Do not print the total number of groups found."`
}

func (cmd *GroupList) Run(container *cli.Container) error {
	cmd.Header = strings.ExpandWhitespaceLiterals(cmd.Header)
	cmd.Footer = strings.ExpandWhitespaceLiterals(cmd.Footer)
	cmd.Delimiter = strings.ExpandWhitespaceLiterals(cmd.Delimiter)

	if cmd.Delimiter0 {
		if cmd.Delimiter != "\n" {
			return errors.New("--delimiter and --delimiter0 can't be used together")
		}
		if cmd.Header != "" {
			return errors.New("--footer and --delimiter0 can't be used together")
		}
		if cmd.Footer != "\n" {
			return errors.New("--footer and --delimiter0 can't be used together")
		}

		cmd.Delimiter = "\x00"
		cmd.Footer = "\x00"
	}

	groups := container.Config.Groups
	count := len(groups)
	var err error = nil
	if count > 0 {
		group_names := make([]string, 0, count)
		for group := range groups {
			group_names = append(group_names, group)
		}

		// Sort the keys
		sort.Strings(group_names)
		err = container.Paginate(cmd.NoPager, func(out io.Writer) error {
			if cmd.Header != "" {
				fmt.Fprint(out, cmd.Header)
			}

			for i, name := range group_names {
				if i > 0 {
					fmt.Fprint(out, cmd.Delimiter)
				}

				fmt.Fprintf(out, name)
			}
			if cmd.Footer != "" {
				fmt.Fprint(out, cmd.Footer)
			}

			return nil
		})
	}

	if err == nil && !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\nFound %d %s\n", count, strings.Pluralize("group", count))
	}

	return err
}
