package cmd

import (
	"fmt"
	"io"
	"os"

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
	if count > 0 {
		err = container.Paginate(cmd.NoPager, func(out io.Writer) error {
			if cmd.Header != "" {
				fmt.Fprint(out, cmd.Header)
			}
			for i, tag := range tags {
				if i > 0 {
					fmt.Fprint(out, cmd.Delimiter)
				}

				ft, err := format(tag)
				if err != nil {
					return err
				}
				fmt.Fprint(out, ft)
			}
			if cmd.Footer != "" {
				fmt.Fprint(out, cmd.Footer)
			}

			return nil
		})
	}

	if err == nil && !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\nFound %d %s\n", count, strings.Pluralize("tag", count))
	}

	return err
}

func (cmd *TagList) tagTemplate() string {
	format := cmd.Format
	if format == "" {
		format = "full"
	}

	templ, ok := defaultTagFormats[format]
	if !ok {
		templ = strings.ExpandWhitespaceLiterals(format)
	}

	return templ
}

var defaultTagFormats = map[string]string{
	"json":  `{{json .}}`,
	"jsonl": `{{json .}}`,
	"name":  `{{name}}`,
	"full":  `{{name}} ({{note-count}})`,
}
