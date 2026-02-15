package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NullableString handles nullable string columns across SQLite (TEXT), MySQL (VARCHAR), PostgreSQL (TEXT)
type NullableString struct {
	String string
	Valid  bool
}

// NewNullableString creates a valid NullableString with the given value.
func NewNullableString(s string) NullableString {
	return NullableString{String: s, Valid: true}
}

// Value returns the database driver value.
func (n NullableString) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.String, nil
}

// Scan reads a database value into NullableString.
func (n *NullableString) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.String = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		n.String, n.Valid = v, true
	case []byte:
		n.String, n.Valid = string(v), true
	default:
		return fmt.Errorf("NullableString: cannot scan %T", value)
	}
	return nil
}

// MarshalJSON encodes NullableString as JSON.
func (n NullableString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

// UnmarshalJSON decodes JSON into NullableString.
func (n *NullableString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.String = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.String)
}

// IsZero reports whether NullableString is null or empty.
func (n NullableString) IsZero() bool { return !n.Valid || n.String == "" }

// NullableInt64 handles nullable integer columns across all databases
type NullableInt64 struct {
	Int64 int64
	Valid bool
}

// NewNullableInt64 creates a valid NullableInt64 with the given value.
func NewNullableInt64(i int64) NullableInt64 {
	return NullableInt64{Int64: i, Valid: true}
}

// Value returns the database driver value.
func (n NullableInt64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Int64, nil
}

// Scan reads a database value into NullableInt64.
func (n *NullableInt64) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.Int64 = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		n.Int64, n.Valid = v, true
	case int32:
		n.Int64, n.Valid = int64(v), true
	case int:
		n.Int64, n.Valid = int64(v), true
	case float64:
		n.Int64, n.Valid = int64(v), true
	default:
		return fmt.Errorf("NullableInt64: cannot scan %T", value)
	}
	return nil
}

// MarshalJSON encodes NullableInt64 as JSON.
func (n NullableInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int64)
}

// UnmarshalJSON decodes JSON into NullableInt64.
func (n *NullableInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Int64 = 0
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Int64)
}

// IsZero reports whether NullableInt64 is null or zero.
func (n NullableInt64) IsZero() bool { return !n.Valid || n.Int64 == 0 }

// NullableFloat64 handles nullable float columns across SQLite (REAL), MySQL (FLOAT), PostgreSQL (REAL)
type NullableFloat64 struct {
	Float64 float64
	Valid   bool
}

// NewNullableFloat64 creates a valid NullableFloat64 with the given value.
func NewNullableFloat64(f float64) NullableFloat64 {
	return NullableFloat64{Float64: f, Valid: true}
}

// Value returns the database driver value.
func (n NullableFloat64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Float64, nil
}

// Scan reads a database value into NullableFloat64.
func (n *NullableFloat64) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.Float64 = 0
		return nil
	}
	switch v := value.(type) {
	case float64:
		n.Float64, n.Valid = v, true
	case float32:
		n.Float64, n.Valid = float64(v), true
	case int64:
		n.Float64, n.Valid = float64(v), true
	case int32:
		n.Float64, n.Valid = float64(v), true
	case int:
		n.Float64, n.Valid = float64(v), true
	default:
		return fmt.Errorf("NullableFloat64: cannot scan %T", value)
	}
	return nil
}

// MarshalJSON encodes NullableFloat64 as JSON.
func (n NullableFloat64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Float64)
}

// UnmarshalJSON decodes JSON into NullableFloat64.
func (n *NullableFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Float64 = 0
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Float64)
}

// IsZero reports whether NullableFloat64 is null or zero.
func (n NullableFloat64) IsZero() bool { return !n.Valid || n.Float64 == 0 }

// NullableBool handles nullable boolean columns across all databases
type NullableBool struct {
	Bool  bool
	Valid bool
}

// NewNullableBool creates a valid NullableBool with the given value.
func NewNullableBool(b bool) NullableBool {
	return NullableBool{Bool: b, Valid: true}
}

// Value returns the database driver value.
func (n NullableBool) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Bool, nil
}

// Scan reads a database value into NullableBool.
func (n *NullableBool) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.Bool = false
		return nil
	}
	switch v := value.(type) {
	case bool:
		n.Bool, n.Valid = v, true
	case int64:
		n.Bool, n.Valid = v != 0, true
	case int32:
		n.Bool, n.Valid = v != 0, true
	case int:
		n.Bool, n.Valid = v != 0, true
	default:
		return fmt.Errorf("NullableBool: cannot scan %T", value)
	}
	return nil
}

// MarshalJSON encodes NullableBool as JSON.
func (n NullableBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Bool)
}

// UnmarshalJSON decodes JSON into NullableBool.
func (n *NullableBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Bool = false
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Bool)
}

// IsZero reports whether NullableBool is null.
func (n NullableBool) IsZero() bool { return !n.Valid }

// JSONData handles JSON columns across SQLite (TEXT), MySQL (JSON), PostgreSQL (JSONB)
// It stores arbitrary JSON data that can be marshaled/unmarshaled.
type JSONData struct {
	Data  any
	Valid bool
}

// NewJSONData creates a valid JSONData with the given value.
func NewJSONData(data any) JSONData {
	return JSONData{Data: data, Valid: true}
}

// Value returns the database driver value.
func (j JSONData) Value() (driver.Value, error) {
	if !j.Valid || j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

// Scan reads a database value into JSONData.
func (j *JSONData) Scan(value any) error {
	if value == nil {
		j.Valid = false
		j.Data = nil
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("JSONData: cannot scan %T", value)
	}
	if len(data) == 0 {
		j.Valid = false
		j.Data = nil
		return nil
	}
	j.Valid = true
	return json.Unmarshal(data, &j.Data)
}

// MarshalJSON encodes JSONData as JSON.
func (j JSONData) MarshalJSON() ([]byte, error) {
	if !j.Valid || j.Data == nil {
		return []byte("null"), nil
	}
	return json.Marshal(j.Data)
}

// UnmarshalJSON decodes JSON into JSONData.
func (j *JSONData) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		j.Valid = false
		j.Data = nil
		return nil
	}
	j.Valid = true
	return json.Unmarshal(data, &j.Data)
}

// IsZero reports whether JSONData is null or nil.
func (j JSONData) IsZero() bool { return !j.Valid || j.Data == nil }

// String returns the JSON representation as a string
func (j JSONData) String() string {
	if !j.Valid || j.Data == nil {
		return ""
	}
	data, err := json.Marshal(j.Data)
	if err != nil {
		return ""
	}
	return string(data)
}
