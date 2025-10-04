package timeutil

import (
	"time"
)

// Now returns the current UTC time
func Now() time.Time {
	return time.Now().UTC()
}

// StartOfDay returns the start of the day for the given time
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for the given time
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// StartOfMonth returns the start of the month for the given time
func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month for the given time
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// IsExpired checks if a timestamp has expired
func IsExpired(expiresAt time.Time) bool {
	return time.Now().UTC().After(expiresAt)
}

// DaysUntil calculates the number of days until a future date
func DaysUntil(future time.Time) int {
	duration := future.Sub(time.Now().UTC())
	return int(duration.Hours() / 24)
}

// FormatISO8601 formats a time in ISO8601 format
func FormatISO8601(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseISO8601 parses an ISO8601 formatted string
func ParseISO8601(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
