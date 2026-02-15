package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Strict formats for API input - only RFC3339 with timezone
var strictTimestampFormats = []string{
	time.RFC3339,     // "2006-01-02T15:04:05Z07:00" - PRIMARY FORMAT
	time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
}

// Legacy formats for database reads only (historical data compatibility)
// These are NOT accepted for API input
var legacyTimestampFormats = []string{
	"2006-01-02 15:04:05.999999-07:00", // SQLite/PostgreSQL with fractional seconds and numeric TZ
	"2006-01-02 15:04:05.999999Z",      // PostgreSQL TEXT with fractional seconds and Z
	"2006-01-02 15:04:05.999999",       // PostgreSQL TEXT with fractional seconds
	"2006-01-02 15:04:05-07:00",        // SQLite/PostgreSQL with numeric TZ
	"2006-01-02 15:04:05Z",             // PostgreSQL TEXT with Z
	"2006-01-02 15:04:05",              // MySQL without TZ (assume UTC)
	"2006-01-02T15:04:05Z",             // UTC shorthand
	"2006-01-02T15:04:05",              // No TZ (assume UTC)
	"2006-01-02",                       // Date only (assume 00:00:00 UTC)
}

// Timestamp handles datetime columns across SQLite (TEXT), MySQL (DATETIME), PostgreSQL (TIMESTAMP)
// All times are stored and returned in UTC.
type Timestamp struct {
	Time  time.Time
	Valid bool
}

// NewTimestamp creates a Timestamp from a time.Time
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{Time: t.UTC(), Valid: true}
}

// TimestampNow returns the current time as a Timestamp
func TimestampNow() Timestamp {
	return Timestamp{Time: time.Now().UTC(), Valid: true}
}

// String returns the RFC3339 representation of the timestamp or "null" if invalid.
func (t Timestamp) String() string {
	if !t.Valid {
		return "null"
	}
	return t.Time.UTC().Format(time.RFC3339)
}

// IsZero returns true if timestamp is null or zero time
func (t Timestamp) IsZero() bool {
	return !t.Valid || t.Time.IsZero()
}

// Value returns the database driver value for storage or nil if invalid.
func (t Timestamp) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.Time.UTC(), nil
}

// Scan reads from database - accepts legacy formats for compatibility
func (t *Timestamp) Scan(value any) error {
	if value == nil {
		t.Valid = false
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time, t.Valid = v.UTC(), true
		return nil
	case string:
		if v == "" {
			t.Valid = false
			return nil
		}
		// Try strict formats first
		for _, format := range strictTimestampFormats {
			if parsed, err := time.Parse(format, v); err == nil {
				t.Time, t.Valid = parsed.UTC(), true
				return nil
			}
		}
		// Fall back to legacy formats for database reads
		for _, format := range legacyTimestampFormats {
			if parsed, err := time.Parse(format, v); err == nil {
				t.Time, t.Valid = parsed.UTC(), true
				return nil
			}
		}
		return fmt.Errorf("Timestamp: cannot parse %q", v)
	case []byte:
		return t.Scan(string(v))
	default:
		return fmt.Errorf("Timestamp: cannot scan %T", value)
	}
}

// MarshalJSON always outputs RFC3339 in UTC
func (t Timestamp) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(t.Time.UTC().Format(time.RFC3339))
}

// UnmarshalJSON ONLY accepts RFC3339 format with timezone - strict API input validation
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Valid = false
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Timestamp: expected string, got %s", string(data))
	}

	// STRICT: Only accept RFC3339 formats for API input
	for _, format := range strictTimestampFormats {
		if parsed, err := time.Parse(format, s); err == nil {
			t.Time, t.Valid = parsed.UTC(), true
			return nil
		}
	}

	return fmt.Errorf("Timestamp: invalid format %q (must be RFC3339: 2006-01-02T15:04:05Z or 2006-01-02T15:04:05-07:00)", s)
}

// Before reports whether t is before u
func (t Timestamp) Before(u Timestamp) bool {
	if !t.Valid || !u.Valid {
		return false
	}
	return t.Time.Before(u.Time)
}

// After reports whether t is after u
func (t Timestamp) After(u Timestamp) bool {
	if !t.Valid || !u.Valid {
		return false
	}
	return t.Time.After(u.Time)
}

// UTC returns the time in UTC
func (t Timestamp) UTC() time.Time {
	return t.Time.UTC()
}

// Add returns t + duration
func (t Timestamp) Add(d time.Duration) Timestamp {
	if !t.Valid {
		return t
	}
	return Timestamp{Time: t.Time.Add(d), Valid: true}
}

// Sub returns the duration t - u
func (t Timestamp) Sub(u Timestamp) time.Duration {
	if !t.Valid || !u.Valid {
		return 0
	}
	return t.Time.Sub(u.Time)
}
