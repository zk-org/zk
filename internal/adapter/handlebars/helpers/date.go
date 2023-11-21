package helpers

import (
	"os"
	"time"

	"github.com/aymerick/raymond"
	"github.com/lestrrat-go/strftime"
	"github.com/mickael-menu/zk/internal/util"
	dateutil "github.com/mickael-menu/zk/internal/util/date"
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

// RegisterDateFrom registers the {{date-from}} template helper to use the `naturaldate` package to generate time.Time based on language strings,
// based on a reference time which can also be a time string or a time.Time object.
// This can be used in combination with the `format-date` helper to generate dates in the user's language.
// {{format-date (date-from "2006-01-02" "last week") "timestamp"}}
//
// Format is: {{date-from "REFERENCE TIME" "NATURAL DATE TIMESTAMP"}}
// Reference Time can either be a time.Time object or a string. Parsing matches the {{date}} helper, but without natural language parsing
// Natural Date Timestamp behaves the same as the {{date}} helper, but uses the given reference time to offset the generated time.Time object.
func RegisterDateFrom(logger util.Logger) {
	raymond.RegisterHelper("date-from", func(arg1 any, arg2 string) time.Time {
		var refTime time.Time

		switch ref := arg1.(type) {
		case string:
			var err error
			refTime, err = dateutil.ParseTimestamp(ref)
			if err != nil {
				logger.Printf("the {{date-from}} template helper failed to parse the reference date: %v", err)
				return refTime
			}
		case time.Time:
			refTime = ref
		default:
			logger.Printf("the {{date-from}} template helper expects a date string ir a time object as its first argument")
			return refTime
		}

		t, err := dateutil.TimeFromReference(arg2, refTime)
		if err != nil {
			logger.Err(errors.Wrap(err, "the {{date-from}} template helper failed to parse the date"))
		}
		return t
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
