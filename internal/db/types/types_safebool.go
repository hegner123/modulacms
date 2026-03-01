package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// SafeBool handles non-nullable boolean columns across SQLite (INTEGER 0/1) and MySQL/PostgreSQL (BOOLEAN).
// SQLite stores booleans as int64, MySQL/PostgreSQL store them as native bool.
// SafeBool's Scan method accepts both representations, unifying all three drivers.
type SafeBool struct {
	Val bool
}

// NewSafeBool creates a SafeBool with the given value.
func NewSafeBool(b bool) SafeBool {
	return SafeBool{Val: b}
}

// Bool returns the underlying boolean value.
func (s SafeBool) Bool() bool {
	return s.Val
}

// Value returns the database driver value. go-sqlite3 converts bool to int64 automatically.
func (s SafeBool) Value() (driver.Value, error) {
	return s.Val, nil
}

// Scan reads a database value into SafeBool. Accepts bool (MySQL/PostgreSQL) and int64/int32/int (SQLite).
func (s *SafeBool) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("SafeBool: cannot scan nil (non-nullable)")
	}
	switch v := value.(type) {
	case bool:
		s.Val = v
	case int64:
		s.Val = v != 0
	case int32:
		s.Val = v != 0
	case int:
		s.Val = v != 0
	default:
		return fmt.Errorf("SafeBool: cannot scan %T", value)
	}
	return nil
}

// MarshalJSON encodes SafeBool as a JSON boolean.
func (s SafeBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Val)
}

// UnmarshalJSON decodes a JSON boolean into SafeBool.
func (s *SafeBool) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Val)
}

// String returns "true" or "false".
func (s SafeBool) String() string {
	if s.Val {
		return "true"
	}
	return "false"
}
