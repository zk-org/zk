package helpers

import (
	"time"

	"github.com/aymerick/raymond"
	"github.com/mickael-menu/zk/internal/util"
	dateutil "github.com/mickael-menu/zk/internal/util/date"
	"github.com/pkg/errors"
)

// RegisterGetDate registers the {{getdate}} template helper to use the `naturaldate` package to generate time.Time based on language strings.
// This can be used in combination with the `date` helper to generate dates in the user's language.
// {{date (get-date "last week") "timestamp"}}
func RegisterGetDate(logger util.Logger) {
	raymond.RegisterHelper("get-date", func(natural string) time.Time {
		date, err := dateutil.TimeFromNatural(natural)
		if err != nil {
			logger.Err(errors.Wrap(err, "the {{get-date}} template helper failed to parse the date"))
		}
		return date
	})
}
