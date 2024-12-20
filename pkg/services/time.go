package services

import "time"

// Time implements the `TimeProvider` interface.
type Time struct{}

// Now returns the current time.
func (_ *Time) Now() time.Time {
	return time.Now()
}
