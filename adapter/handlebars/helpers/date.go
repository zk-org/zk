package helpers

import (
	"github.com/aymerick/raymond"
	"github.com/lestrrat-go/strftime"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/date"
)

func RegisterDate(logger util.Logger, date date.Provider) {
	raymond.RegisterHelper("date", func(arg string) string {
		format := findFormat(arg)
		res, err := strftime.Format(format, date.Date(), strftime.WithUnixSeconds('s'))
		if err != nil {
			logger.Printf("the {{date}} template helper failed to format the date: %v", err)
			return ""
		}
		return res
	})
}

var (
	shortFormat         = `%m/%d/%Y`
	mediumFormat        = `%b %d, %Y`
	longFormat          = `%B %d, %Y`
	fullFormat          = `%A, %B %d, %Y`
	yearFormat          = `%Y`
	timeFormat          = `%H:%M`
	timestampFormat     = `%Y%m%d%H%M`
	timestampUnixFormat = `%s`
)

func findFormat(key string) string {
	switch key {
	case "short":
		return shortFormat
	case "medium":
		return mediumFormat
	case "long":
		return longFormat
	case "full":
		return fullFormat
	case "year":
		return yearFormat
	case "time":
		return timeFormat
	case "timestamp":
		return timestampFormat
	case "timestamp-unix":
		return timestampUnixFormat
	default:
		return key
	}
}
