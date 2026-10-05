package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/zk-org/zk/internal/cli"
	"github.com/zk-org/zk/internal/util/strings"
)

// AliasList lists all the aliases.
type Config struct {
	List string `short:"l" placeholder:"OBJECT" help:"List configuration objects. Listable objects are: aliases, filters and extras."`
	FormatFlags
}

func (cmd *Config) Run(container *cli.Container) error {
	if err := cmd.FormatFlags.Prepare("{", "}\n", false); err != nil {
		return err
	}

	var objects = make(map[string]string)

	switch cmd.List {
	case "filters":
		objects = container.Config.Filters
	case "aliases":
		objects = container.Config.Aliases
	case "extras":
		objects = container.Config.Extra
	default:
		fmt.Print("Listable config objects are: filters, aliases and extras.\n")
		os.Exit(1)
	}

	count := len(objects)
	keys := make([]string, count)
	i := 0
	for k := range objects {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	format := cmd.mapTemplate()

	var err = container.Paginate(cmd.NoPager, func(out io.Writer) error {
		if cmd.Header != "" {
			fmt.Fprint(out, cmd.Header)
		}
		for i, o := range keys {

			if i > 0 {
				fmt.Fprint(out, cmd.Delimiter)
			}
			if cmd.Format == "" || cmd.Format == "short" {
				fmt.Fprintf(out, format, o)
			} else if cmd.Format == "json" {
				jsonData, err := json.Marshal(objects[o])
				if err != nil {
					fmt.Println("Error marshaling JSON:", err)
					os.Exit(1)
				}
				fmt.Fprintf(out, format, o, jsonData)
			} else {
				fmt.Fprintf(out, format, o, objects[o])
			}

			i += 1
		}
		if cmd.Footer != "" {
			fmt.Fprint(out, cmd.Footer)
		}
		return nil
	})

	if err == nil && !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\nFound %d %s\n", count, cmd.List)
	}
	return err
}

func (cmd *Config) mapTemplate() string {
	format := cmd.Format
	if format == "" {
		format = "short"
	}

	templ, ok := defaultMapFormats[format]
	if !ok {
		templ = strings.ExpandWhitespaceLiterals(format)
	}

	return templ
}

var defaultMapFormats = map[string]string{
	"json":  `"%s":%s`,
	"short": `%s`,
	"full":  `%12s    %s`,
}
