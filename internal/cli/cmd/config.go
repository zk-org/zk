package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/zk-org/zk/internal/cli"
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

	err := cmd.FormatFlags.Paginate(container, count, func(out io.Writer, i int) error {
		o := keys[i]
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
		return nil
	})
	if err != nil {
		return err
	}

	if !cmd.Quiet {
		fmt.Fprintf(os.Stderr, "\nFound %d %s\n", count, cmd.List)
	}
	return nil
}

func (cmd *Config) mapTemplate() string {
	return cmd.Template("short", defaultMapFormats)
}

var defaultMapFormats = map[string]string{
	"json":  `"%s":%s`,
	"short": `%s`,
	"full":  `%12s    %s`,
}
