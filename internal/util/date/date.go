package date

import (
	"time"

	naturaldate "github.com/tj/go-naturaldate"
)

// TimeFromNatural parses a human date into a time.Time.
func TimeFromNatural(date string) (time.Time, error) {
	if date == "" {
		return time.Now(), nil
	}
	if t, err := time.Parse(time.RFC3339, date); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006", date, time.Local); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("15:04", date, time.Local); err == nil {
		return t, nil
	}
	return naturaldate.Parse(date, time.Now(), naturaldate.WithDirection(naturaldate.Past))
}
