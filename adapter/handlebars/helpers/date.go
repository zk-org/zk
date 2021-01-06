package helpers

import (
	"time"

	"github.com/aymerick/raymond"
	"github.com/lestrrat-go/strftime"
	"github.com/mickael-menu/zk/util"
	"github.com/mickael-menu/zk/util/date"
)

// RegisterDate registers the {{date}} template helpers which format a given date.
//
// It supports various styles: short, medium, long, full, year, time,
// timestamp, timestamp-unix or a custom strftime format.
//
// {{date "medium"}} -> Nov 17, 2009
// {{date "%Y-%m"}} -> 2009-11
// {{date created "%Y-%m-%d"}} -> 2008-12-05
func RegisterDate(logger util.Logger, date date.Provider) {
	raymond.RegisterHelper("date", func(arg1 interface{}, arg2 interface{}) string {
		format := "%Y-%m-%d"
		date := date.Date()

		switch arg := arg1.(type) {
		case string:
			format = findFormat(arg)
		case time.Time:
			date = arg
		}

		if arg, ok := arg2.(string); ok {
			format = findFormat(arg)
		}

		res, err := strftime.Format(format, date, strftime.WithUnixSeconds('s'))
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
