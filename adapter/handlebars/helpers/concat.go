package helpers

import (
	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/util"
)

// RegisterConcat registers a {{concat}} template helper which concatenates two
// strings.
//
// {{concat '> ' 'A quote'}} -> "> A quote"
//
func RegisterConcat(logger util.Logger) {
	raymond.RegisterHelper("concat", func(a, b string) string {
		return a + b
	})
}
