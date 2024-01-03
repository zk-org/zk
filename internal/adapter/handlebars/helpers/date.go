package helpers

import (
	"os"
	"time"

	"github.com/aymerick/raymond"
	"github.com/lestrrat-go/strftime"
	"github.com/zk-org/zk/internal/util"
	dateutil "github.com/zk-org/zk/internal/util/date"
	"github.com/pkg/errors"
	"github.com/rvflash/elapsed"
)

// RegisterDate registers the {{date}} template helper to use the `naturaldate` package to generate time.Time based on language strings.
// This can be used in combination with the `format-date` helper to generate dates in the user's language.
// {{format-date (date "last week") "timestamp"}}
func RegisterDate(logger util.Logger) {
	raymond.RegisterHelper("date", func(arg1 interface{}, arg2 interface{}) time.Time {
		var t time.Time
		switch date := arg1.(type) {
		case string:
			t, err := dateutil.TimeFromNatural(date)
			if err != nil {
				logger.Err(errors.Wrap(err, "the {{date}} template helper failed to parse the date"))
			}
			return t
		case time.Time:
			logger.Println("the {{date}} template helper was renamed to {{format-date}}, please update your configuration")
			os.Exit(1)
			return t
		default:
			logger.Println("the {{date}} template helper expects a natural human date as a string for its only argument")
			return t
		}
	})
}

// RegisterFormatDate registers the {{format-date}} template helpers which format a given date.
//
// It supports various styles: short, medium, long, full, year, time,
// timestamp, timestamp-unix or a custom strftime format.
//
// {{format-date now}} -> 2009-11-17
// {{format-date now "medium"}} -> Nov 17, 2009
// {{format-date now "%Y-%m"}} -> 2009-11
func RegisterFormatDate(logger util.Logger) {
	raymond.RegisterHelper("format-date", func(date time.Time, arg interface{}) string {
		format := "%Y-%m-%d"

		if arg, ok := arg.(string); ok {
			format = findFormat(arg)
		}

		if format == "elapsed" {
			return elapsed.Time(date)

		} else {
			res, err := strftime.Format(format, date, strftime.WithUnixSeconds('s'))
			if err != nil {
				logger.Printf("the {{format-date}} template helper failed to format the date: %v", err)
				return ""
			}
			return res
		}
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
