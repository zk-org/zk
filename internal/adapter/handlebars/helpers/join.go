package helpers

import (
	"strings"

	"github.com/aymerick/raymond"
)

// RegisterJoin registers a {{join}} template helper which concatenates list
// items with the given separator.
//
// {{join list ', '}} -> item1, item2, item3
func RegisterJoin() {
	raymond.RegisterHelper("join", func(list []string, delimiter string) string {
		return strings.Join(list, delimiter)
	})
}
