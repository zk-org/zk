package helpers

import (
	"strings"

	"github.com/aymerick/raymond"
	"github.com/zk-org/zk/internal/core"
	"github.com/zk-org/zk/internal/util"
)

// NewStyleHelper creates a new template helper which stylizes the text input
// according to predefined styling rules.
//
// {{style "date" created}}
// {{#style "red"}}Hello, world{{/style}}
func NewStyleHelper(styler core.Styler, logger util.Logger) interface{} {
	style := func(keys string, text string) string {
		rules := make([]core.Style, 0)
		for _, key := range strings.Fields(keys) {
			rules = append(rules, core.Style(key))
		}
		res, err := styler.Style(text, rules...)
		if err != nil {
			logger.Err(err)
			return text
		} else {
			return res
		}
	}

	return func(rules string, opt interface{}) string {
		switch arg := opt.(type) {
		case *raymond.Options:
			return style(rules, arg.Fn())
		case string:
			return style(rules, arg)
		default:
			logger.Printf("the {{style}} template helper is expecting a string as input, received: %v", opt)
			return ""
		}
	}
}
