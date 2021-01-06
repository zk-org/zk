package helpers

import (
	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/strings"
)

// RegisterPrepend registers a {{prepend}} template helper which prepend a
// given string at the beginning of each line.
//
// {{prepend '> ' 'A quote'}} -> "> A quote"
// {{#prepend '> '}}A quote{{/prepend}} -> "> A quote"
//
// A quote on
// several lines
// {{/prepend}}
//
// > A quote on
// > several lines
func RegisterPrepend(logger util.Logger) {
	raymond.RegisterHelper("prepend", func(prefix string, opt interface{}) string {
		switch arg := opt.(type) {
		case *raymond.Options:
			return strings.Prepend(arg.Fn(), prefix)
		case string:
			return strings.Prepend(arg, prefix)
		default:
			logger.Printf("the {{prepend}} template helper is expecting a string as argument, received: %v", opt)
			return ""
		}
	})
}
