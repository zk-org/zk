package date

import (
	"time"

	"github.com/tj/go-naturaldate"
)

// Provider returns a date instance.
type Provider interface {
	Date() time.Time
}

// Now is a date provider returning the current date.
type Now struct{}

func (n *Now) Date() time.Time {
	return time.Now()
}

// Frozen is a date provider returning always the same date.
type Frozen struct {
	date time.Time
}

func NewFrozenNow() Frozen {
	return Frozen{time.Now()}
}

func NewFrozen(date time.Time) Frozen {
	return Frozen{date}
}

func (n *Frozen) Date() time.Time {
	return n.date
}

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
	if t, err := time.ParseInLocation("2006-01-02T15:04", date, time.Local); err == nil {
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
