package date

import (
	"strconv"
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
	if i, err := strconv.ParseInt(date, 10, 0); err == nil && i >= 1000 && i < 5000 {
		return time.Date(int(i), time.January, 0, 0, 0, 0, 0, time.UTC), nil
	}
	return naturaldate.Parse(date, time.Now().UTC(), naturaldate.WithDirection(naturaldate.Past))
}
