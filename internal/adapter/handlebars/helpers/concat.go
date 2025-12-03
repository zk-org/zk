package helpers

import (
	"github.com/aymerick/raymond"
)

// RegisterConcat registers a {{concat}} template helper which concatenates two
// strings.
//
// {{concat '> ' 'A quote'}} -> "> A quote"
func RegisterConcat() {
	raymond.RegisterHelper("concat", func(a, b string) string {
		return a + b
	})
}
