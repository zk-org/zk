package helpers

import (
	"github.com/mickael-menu/zk/internal/core"
	"github.com/mickael-menu/zk/internal/util"
)

// NewLinkHelper creates a new template helper to generate an internal link
// using a LinkFormatter.
//
// {{link "path/to/note.md" "An interesting subject"}} -> (depends on the LinkFormatter)
//   [[path/to/note]]
//   [An interesting subject](path/to/note)
func NewLinkHelper(formatter core.LinkFormatter, logger util.Logger) interface{} {
	return func(path string, opt interface{}) string {
		title, _ := opt.(string)
		link, err := formatter(path, title)
		if err != nil {
			logger.Err(err)
			return ""
		}

		return link
	}
}
