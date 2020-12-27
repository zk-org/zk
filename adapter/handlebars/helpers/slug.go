package helpers

import (
	"github.com/aymerick/raymond"
	"github.com/gosimple/slug"
	"github.com/mickael-menu/zk/util"
)

func RegisterSlug(logger util.Logger, lang string) {
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
