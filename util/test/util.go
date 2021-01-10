package test

import "time"

func Date(s string) time.Time {
	date, _ := time.Parse(time.RFC3339, s)
	return date
}
