package cmd

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
)

// TagPositions is the `zk tag positions` subcommand.
type TagPositions struct {
	Sort    []string `group:sort short:"s" help:"Sort the output." default:"name"`
	NoPager bool     `group:format short:P help:"Do not pipe output into a pager."`
}

func (cmd *TagPositions) Run(container *cli.Container) error {
	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	sorters, err := core.TaggedNoteSorterFromStrings(cmd.Sort)
	if err != nil {
		return err
	}
	sortersLength := len(sorters)
	if sortersLength > 0 {
		if sorters[0].Field != core.TagSortName {
			return errors.New("If tagged notes are sorted, the sort must first be over the tag name.")
		}
	}

	taggedNotes, err := notebook.FindTaggedNotesAll(sorters)
	if err != nil {
		return err
	}

	groupedTaggedNotes := core.GroupTaggedNotes(taggedNotes)

	err = container.Paginate(cmd.NoPager, func(out io.Writer) error {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(groupedTaggedNotes)
	})

	return err
}
