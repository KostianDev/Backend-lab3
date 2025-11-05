package services

import "time"

// TimeProvider abstracts current time access to simplify testing.
type TimeProvider interface {
    Now() time.Time
}

// SystemTimeProvider implements TimeProvider using the real clock.
type SystemTimeProvider struct{}

func (SystemTimeProvider) Now() time.Time {
    return time.Now()
}
