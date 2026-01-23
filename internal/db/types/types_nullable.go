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

func NewNullableString(s string) NullableString {
	return NullableString{String: s, Valid: true}
}

func (n NullableString) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.String, nil
}

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

func (n NullableString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

func (n *NullableString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.String = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.String)
}

func (n NullableString) IsZero() bool { return !n.Valid || n.String == "" }

// NullableInt64 handles nullable integer columns across all databases
type NullableInt64 struct {
	Int64 int64
	Valid bool
}

func NewNullableInt64(i int64) NullableInt64 {
	return NullableInt64{Int64: i, Valid: true}
}

func (n NullableInt64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Int64, nil
}

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

func (n NullableInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int64)
}

func (n *NullableInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Int64 = 0
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Int64)
}

func (n NullableInt64) IsZero() bool { return !n.Valid || n.Int64 == 0 }

// NullableBool handles nullable boolean columns across all databases
type NullableBool struct {
	Bool  bool
	Valid bool
}

func NewNullableBool(b bool) NullableBool {
	return NullableBool{Bool: b, Valid: true}
}

func (n NullableBool) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Bool, nil
}

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

func (n NullableBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Bool)
}

func (n *NullableBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Bool = false
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Bool)
}

func (n NullableBool) IsZero() bool { return !n.Valid }

// JSONData handles JSON columns across SQLite (TEXT), MySQL (JSON), PostgreSQL (JSONB)
// It stores arbitrary JSON data that can be marshaled/unmarshaled.
type JSONData struct {
	Data  any
	Valid bool
}

func NewJSONData(data any) JSONData {
	return JSONData{Data: data, Valid: true}
}

func (j JSONData) Value() (driver.Value, error) {
	if !j.Valid || j.Data == nil {
		return nil, nil
	}
	return json.Marshal(j.Data)
}

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

func (j JSONData) MarshalJSON() ([]byte, error) {
	if !j.Valid || j.Data == nil {
		return []byte("null"), nil
	}
	return json.Marshal(j.Data)
}

func (j *JSONData) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		j.Valid = false
		j.Data = nil
		return nil
	}
	j.Valid = true
	return json.Unmarshal(data, &j.Data)
}

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
