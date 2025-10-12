package sqlutil

import (
	"database/sql"
	"time"
)

// NullTimeToTime converts a sql.NullTime to time.Time.
// Returns zero time (time.Time{}) if the value is not valid.
//
// This helper eliminates the repetitive null-checking pattern:
//
//	// Before:
//	paymentDate := row.PaymentDate.Time
//	if !row.PaymentDate.Valid {
//	    paymentDate = time.Time{}
//	}
//
//	// After:
//	paymentDate := sqlutil.NullTimeToTime(row.PaymentDate)
func NullTimeToTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// TimeToNullTime converts a time.Time to sql.NullTime.
// The result is invalid if the input is a zero time.
//
// This helper is useful for the reverse conversion when inserting
// or updating database records with nullable timestamp fields.
func TimeToNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// NullStringToPtr converts a sql.NullString to *string.
// Returns nil if the value is not valid.
//
// This helper eliminates the null-checking pattern:
//
//	// Before:
//	var cardMask *string
//	if row.CardMask.Valid {
//	    cardMask = &row.CardMask.String
//	}
//
//	// After:
//	cardMask := sqlutil.NullStringToPtr(row.CardMask)
func NullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// NullStringToString converts a sql.NullString to string.
// Returns empty string if the value is not valid.
//
// This helper eliminates the null-checking pattern:
//
//	// Before:
//	description := ""
//	if row.Description.Valid {
//	    description = row.Description.String
//	}
//
//	// After:
//	description := sqlutil.NullStringToString(row.Description)
func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// StringToNullString converts a string to sql.NullString.
// The result is invalid if the input is empty.
//
// This helper is useful for the reverse conversion when inserting
// or updating database records with nullable string fields.
func StringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// PtrToNullString converts a *string to sql.NullString.
// The result is invalid if the pointer is nil.
//
// This helper is useful when converting from domain entities with
// nullable pointer fields to database nullable types.
func PtrToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullInt64ToPtr converts a sql.NullInt64 to *int64.
// Returns nil if the value is not valid.
func NullInt64ToPtr(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

// PtrToNullInt64 converts an *int64 to sql.NullInt64.
// The result is invalid if the pointer is nil.
func PtrToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// NullBoolToPtr converts a sql.NullBool to *bool.
// Returns nil if the value is not valid.
func NullBoolToPtr(nb sql.NullBool) *bool {
	if nb.Valid {
		return &nb.Bool
	}
	return nil
}

// PtrToNullBool converts a *bool to sql.NullBool.
// The result is invalid if the pointer is nil.
func PtrToNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}
