package helpers

import (
	"strings"

	"github.com/aymerick/raymond"
)

// RegisterList registers a {{list}} template helper which formats a slice of
// strings into a bulleted list.
func RegisterList() {
	itemify := func(text string) string {
		lines := strings.SplitAfter(strings.TrimRight(text, "\n"), "\n")
		return "  ‣ " + strings.Join(lines, "    ")
	}

	raymond.RegisterHelper("list", func(items []string) string {
		res := ""
		for _, item := range items {
			if item == "" {
				continue
			}

			res += itemify(item) + "\n"
		}

		return res
	})
}
