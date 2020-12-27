package date

import "time"

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
