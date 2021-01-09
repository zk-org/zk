package helpers

import (
	"strings"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/core"
	"github.com/mickael-menu/zk/util"
)

// RegisterStyle register the {{style}} template helpers which stylizes the
// text input according to predefined styling rules.
//
// {{style "date" created}}
// {{#style "red"}}Hello, world{{/style}}
func RegisterStyle(styler core.Styler, logger util.Logger) {
	style := func(keys string, text string) string {
		rules := make([]core.StyleRule, 0)
		for _, key := range strings.Fields(keys) {
			rules = append(rules, core.StyleRule(key))
		}
		res, err := styler.Style(text, rules...)
		if err != nil {
			logger.Err(err)
			return text
		} else {
			return res
		}
	}

	raymond.RegisterHelper("style", func(rules string, opt interface{}) string {
		switch arg := opt.(type) {
		case *raymond.Options:
			return style(rules, arg.Fn())
		case string:
			return style(rules, arg)
		default:
			logger.Printf("the {{style}} template helper is expecting a string as input, received: %v", opt)
			return ""
		}
	})
}
