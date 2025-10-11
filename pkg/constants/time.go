// Package constants provides centralized constants for the application.
// This helps avoid magic numbers and provides clear semantic meaning
// for time-related values used throughout the codebase.
package constants

import "time"

// Time conversion constants for better readability
const (
	// Seconds
	SecondsPerMinute = 60
	SecondsPerHour   = 3600
	SecondsPerDay    = 86400
	SecondsPerWeek   = 604800
	SecondsPerMonth  = 2592000 // 30 days approximation

	// Minutes
	MinutesPerHour = 60
	MinutesPerDay  = 1440
	MinutesPerWeek = 10080

	// Hours
	HoursPerDay  = 24
	HoursPerWeek = 168
)

// Duration constants using Go's time.Duration
const (
	// Token expiration durations
	DefaultAccessTokenDuration  = 24 * time.Hour
	DefaultRefreshTokenDuration = 7 * 24 * time.Hour

	// Payment related durations
	PaymentExpirationDuration = 30 * time.Minute
	PaymentRetryDelay         = 5 * time.Minute

	// Reservation durations
	ReservationExpirationDuration = 3 * 24 * time.Hour // 3 days

	// Cache durations
	DefaultCacheTTL = 1 * time.Hour
	BookCacheTTL    = 2 * time.Hour
	AuthorCacheTTL  = 24 * time.Hour

	// API timeouts
	DefaultHTTPTimeout = 30 * time.Second
	GatewayHTTPTimeout = 60 * time.Second
	DatabaseTimeout    = 10 * time.Second

	// Background job intervals
	PaymentExpiryCheckInterval = 5 * time.Minute
	CacheCleanupInterval       = 1 * time.Hour
)

// FormatDuration returns a human-readable string for a duration
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return string(rune(days)) + " days"
}

// ParseDurationOrDefault parses a duration string and returns the default if parsing fails
func ParseDurationOrDefault(s string, defaultDuration time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultDuration
	}
	return d
}
