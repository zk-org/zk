package cmd

import (
	"fmt"
	"io"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util/strings"
)

// Tag manages the note tags in the notebook.
type Tag struct {
	List TagList `cmd group:"cmd" default:"withargs" help:"List all the note tags."`
}

// TagList lists all the note tags.
type TagList struct {
	FormatFlags
	Sort []string `group:sort short:s placeholder:TERM help:"Order the tags by the given criterion."`
}

func (cmd *TagList) Run(container *cli.Container) error {
	if err := cmd.FormatFlags.Prepare("[", "]\n", true); err != nil {
		return err
	}

	notebook, err := container.CurrentNotebook()
	if err != nil {
		return err
	}

	format, err := notebook.NewCollectionFormatter(cmd.tagTemplate())
	if err != nil {
		return err
	}

	sorters, err := core.CollectionSortersFromStrings(cmd.Sort)
	if err != nil {
		return err
	}

	tags, err := notebook.FindCollections(core.CollectionKindTag, sorters)
	if err != nil {
		return err
	}

	count := len(tags)
	err = cmd.FormatFlags.Paginate(container, count, func(out io.Writer, i int) error {
		ft, err := format(tags[i])
		if err != nil {
			return err
		}
		fmt.Fprint(out, ft)
		return nil
	})
	if err != nil {
		return err
	}

	cmd.FormatFlags.PrintFound(count, strings.Pluralize("tag", count))

	return nil
}

func (cmd *TagList) tagTemplate() string {
	return cmd.Template("full", defaultTagFormats)
}

var defaultTagFormats = map[string]string{
	"json":  `{{json .}}`,
	"jsonl": `{{json .}}`,
	"name":  `{{name}}`,
	"full":  `{{name}} ({{note-count}})`,
}
