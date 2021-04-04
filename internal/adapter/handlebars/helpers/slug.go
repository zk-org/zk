package helpers

import (
	"github.com/aymerick/raymond"
	"github.com/gosimple/slug"
	"github.com/mickael-menu/zk/internal/util"
)

// RegisterSlug registers a {{slug}} template helper to slugify text.
//
// {{slug "This will be slugified!"}} -> this-will-be-slugified
// {{#slug}}This will be slugified!{{/slug}} -> this-will-be-slugified
func RegisterSlug(lang string, logger util.Logger) {
	raymond.RegisterHelper("slug", func(opt interface{}) string {
		switch arg := opt.(type) {
		case *raymond.Options:
			return slug.MakeLang(arg.Fn(), lang)
		case string:
			return slug.MakeLang(arg, lang)
		default:
			logger.Printf("the {{slug}} template helper is expecting a string as argument, received: %v", opt)
			return ""
		}
	})
}
