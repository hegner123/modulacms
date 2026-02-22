package db

import (
	"database/sql"
	"encoding/json"
	"time"

)

// NullInt32 wraps sql.NullInt32 with JSON marshaling support.
type NullInt32 struct {
	sql.NullInt32
}

// MarshalJSON marshals NullInt32 to JSON, returning null if invalid.
func (n NullInt32) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int32)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals JSON into NullInt32, handling null values.
func (n *NullInt32) UnmarshalJSON(data []byte) error {
	var x *int32
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		n.Int32 = *x
		n.Valid = true
	} else {
		n.Valid = false
	}
	return nil
}

// NullInt64 wraps sql.NullInt64 with JSON marshaling support.
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON marshals NullInt64 to JSON, returning null if invalid.
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Int64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals JSON into NullInt64, handling null values.
func (n *NullInt64) UnmarshalJSON(data []byte) error {
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		n.Int64 = *x
		n.Valid = true
	} else {
		n.Valid = false
	}
	return nil
}

// NewNullInt64 creates a NullInt64 from a plain int64.
func NewNullInt64(i int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: i, Valid: true}}
}

// NullString wraps sql.NullString with JSON marshaling support.
// Serializes to the string value or null, instead of {"String":"...","Valid":true}.
type NullString struct {
	sql.NullString
}

// NewNullString creates a NullString from a plain string.
// Empty string and "null" produce an invalid (null) NullString.
func NewNullString(s string) NullString {
	return NullString{StringToNullString(s)}
}

// MarshalJSON marshals NullString to JSON, returning null if invalid.
func (n NullString) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals JSON into NullString, handling null values.
func (n *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		n.String = *s
		n.Valid = true
	} else {
		n.Valid = false
	}
	return nil
}

// NullTime wraps sql.NullTime with JSON marshaling support.
type NullTime struct {
	sql.NullTime
}

// MarshalJSON marshals NullTime to JSON, returning null if invalid.
func (n NullTime) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(n.Time)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON unmarshals JSON into NullTime, handling null values.
func (n *NullTime) UnmarshalJSON(data []byte) error {
	var x *time.Time
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		n.Time = *x
		n.Valid = true
	} else {
		n.Valid = false
	}
	return nil
}
