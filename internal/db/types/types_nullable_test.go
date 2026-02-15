package types

import (
	"encoding/json"
	"testing"
)

// ============================================================
// NullableString
// ============================================================

func TestNullableString_New(t *testing.T) {
	t.Parallel()
	ns := NewNullableString("hello")
	if !ns.Valid {
		t.Error("NewNullableString.Valid = false")
	}
	if ns.String != "hello" {
		t.Errorf("NewNullableString.String = %q", ns.String)
	}
	if ns.IsZero() {
		t.Error("NewNullableString.IsZero() = true")
	}
}

func TestNullableString_ZeroValue(t *testing.T) {
	t.Parallel()
	var ns NullableString
	if ns.Valid {
		t.Error("zero NullableString.Valid = true")
	}
	if !ns.IsZero() {
		t.Error("zero NullableString.IsZero() = false")
	}
}

func TestNullableString_IsZero_ValidButEmpty(t *testing.T) {
	t.Parallel()
	ns := NullableString{String: "", Valid: true}
	if !ns.IsZero() {
		t.Error("valid+empty NullableString.IsZero() = false, want true")
	}
}

func TestNullableString_Value(t *testing.T) {
	t.Parallel()
	// Null
	ns := NullableString{Valid: false}
	v, err := ns.Value()
	if err != nil {
		t.Fatalf("null Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null Value() = %v, want nil", v)
	}

	// Valid
	ns = NewNullableString("hello")
	v, err = ns.Value()
	if err != nil {
		t.Fatalf("valid Value() error = %v", err)
	}
	if v != "hello" {
		t.Errorf("valid Value() = %v, want %q", v, "hello")
	}
}

func TestNullableString_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantStr string
		wantOk  bool
		wantErr bool
	}{
		{name: "nil", input: nil, wantStr: "", wantOk: false, wantErr: false},
		{name: "string", input: "hello", wantStr: "hello", wantOk: true, wantErr: false},
		{name: "bytes", input: []byte("world"), wantStr: "world", wantOk: true, wantErr: false},
		{name: "int", input: 42, wantStr: "", wantOk: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ns NullableString
			err := ns.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if ns.Valid != tt.wantOk {
					t.Errorf("Scan(%v) Valid = %v, want %v", tt.input, ns.Valid, tt.wantOk)
				}
				if ns.String != tt.wantStr {
					t.Errorf("Scan(%v) String = %q, want %q", tt.input, ns.String, tt.wantStr)
				}
			}
		})
	}
}

func TestNullableString_JSON(t *testing.T) {
	t.Parallel()
	// Null
	ns := NullableString{Valid: false}
	data, err := json.Marshal(ns)
	if err != nil {
		t.Fatalf("MarshalJSON(null) error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON(null) = %s", data)
	}

	var got NullableString
	if err := json.Unmarshal([]byte("null"), &got); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if got.Valid {
		t.Error("UnmarshalJSON(null) Valid = true")
	}

	// Valid round-trip
	ns = NewNullableString("test")
	data, err = json.Marshal(ns)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid || got.String != "test" {
		t.Errorf("JSON round-trip: Valid=%v String=%q", got.Valid, got.String)
	}
}

// ============================================================
// NullableInt64
// ============================================================

func TestNullableInt64_New(t *testing.T) {
	t.Parallel()
	ni := NewNullableInt64(42)
	if !ni.Valid || ni.Int64 != 42 {
		t.Errorf("NewNullableInt64(42) = {%d, %v}", ni.Int64, ni.Valid)
	}
	if ni.IsZero() {
		t.Error("NewNullableInt64(42).IsZero() = true")
	}
}

func TestNullableInt64_IsZero(t *testing.T) {
	t.Parallel()
	// Invalid
	var ni NullableInt64
	if !ni.IsZero() {
		t.Error("zero NullableInt64.IsZero() = false")
	}
	// Valid with 0
	ni = NullableInt64{Int64: 0, Valid: true}
	if !ni.IsZero() {
		t.Error("valid+0 NullableInt64.IsZero() = false, want true")
	}
	// Valid with non-zero
	ni = NewNullableInt64(-1)
	if ni.IsZero() {
		t.Error("valid+-1 NullableInt64.IsZero() = true")
	}
}

func TestNullableInt64_Value(t *testing.T) {
	t.Parallel()
	ni := NullableInt64{Valid: false}
	v, err := ni.Value()
	if err != nil {
		t.Fatalf("null Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null Value() = %v", v)
	}

	ni = NewNullableInt64(99)
	v, err = ni.Value()
	if err != nil {
		t.Fatalf("valid Value() error = %v", err)
	}
	if v != int64(99) {
		t.Errorf("valid Value() = %v, want 99", v)
	}
}

func TestNullableInt64_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantVal int64
		wantOk  bool
		wantErr bool
	}{
		{name: "nil", input: nil, wantVal: 0, wantOk: false, wantErr: false},
		{name: "int64", input: int64(42), wantVal: 42, wantOk: true, wantErr: false},
		{name: "int32", input: int32(7), wantVal: 7, wantOk: true, wantErr: false},
		{name: "int", input: int(100), wantVal: 100, wantOk: true, wantErr: false},
		{name: "float64", input: float64(3.14), wantVal: 3, wantOk: true, wantErr: false},
		{name: "string", input: "nope", wantVal: 0, wantOk: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ni NullableInt64
			err := ni.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if ni.Valid != tt.wantOk {
					t.Errorf("Scan(%v) Valid = %v", tt.input, ni.Valid)
				}
				if ni.Int64 != tt.wantVal {
					t.Errorf("Scan(%v) Int64 = %d, want %d", tt.input, ni.Int64, tt.wantVal)
				}
			}
		})
	}
}

func TestNullableInt64_JSON(t *testing.T) {
	t.Parallel()
	// Null
	ni := NullableInt64{Valid: false}
	data, err := json.Marshal(ni)
	if err != nil {
		t.Fatalf("MarshalJSON(null) error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON(null) = %s", data)
	}

	var got NullableInt64
	if err := json.Unmarshal([]byte("null"), &got); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if got.Valid {
		t.Error("UnmarshalJSON(null) Valid = true")
	}

	// Valid
	ni = NewNullableInt64(42)
	data, _ = json.Marshal(ni)
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid || got.Int64 != 42 {
		t.Errorf("JSON round-trip: Valid=%v Int64=%d", got.Valid, got.Int64)
	}
}

// ============================================================
// NullableBool
// ============================================================

func TestNullableBool_New(t *testing.T) {
	t.Parallel()
	nb := NewNullableBool(true)
	if !nb.Valid || !nb.Bool {
		t.Errorf("NewNullableBool(true) = {%v, %v}", nb.Bool, nb.Valid)
	}
	if nb.IsZero() {
		t.Error("NewNullableBool(true).IsZero() = true")
	}
}

func TestNullableBool_IsZero(t *testing.T) {
	t.Parallel()
	var nb NullableBool
	if !nb.IsZero() {
		t.Error("zero NullableBool.IsZero() = false")
	}
	// Valid false is NOT zero (it's a meaningful value)
	nb = NewNullableBool(false)
	if nb.IsZero() {
		t.Error("valid+false NullableBool.IsZero() = true, want false")
	}
}

func TestNullableBool_Value(t *testing.T) {
	t.Parallel()
	nb := NullableBool{Valid: false}
	v, err := nb.Value()
	if err != nil {
		t.Fatalf("null Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null Value() = %v", v)
	}

	nb = NewNullableBool(true)
	v, err = nb.Value()
	if err != nil {
		t.Fatalf("valid Value() error = %v", err)
	}
	if v != true {
		t.Errorf("valid Value() = %v, want true", v)
	}
}

func TestNullableBool_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantVal bool
		wantOk  bool
		wantErr bool
	}{
		{name: "nil", input: nil, wantVal: false, wantOk: false, wantErr: false},
		{name: "bool true", input: true, wantVal: true, wantOk: true, wantErr: false},
		{name: "bool false", input: false, wantVal: false, wantOk: true, wantErr: false},
		{name: "int64 nonzero", input: int64(1), wantVal: true, wantOk: true, wantErr: false},
		{name: "int64 zero", input: int64(0), wantVal: false, wantOk: true, wantErr: false},
		{name: "int32 nonzero", input: int32(1), wantVal: true, wantOk: true, wantErr: false},
		{name: "int nonzero", input: int(1), wantVal: true, wantOk: true, wantErr: false},
		{name: "string", input: "true", wantVal: false, wantOk: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var nb NullableBool
			err := nb.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if nb.Valid != tt.wantOk {
					t.Errorf("Scan(%v) Valid = %v", tt.input, nb.Valid)
				}
				if nb.Bool != tt.wantVal {
					t.Errorf("Scan(%v) Bool = %v, want %v", tt.input, nb.Bool, tt.wantVal)
				}
			}
		})
	}
}

func TestNullableBool_JSON(t *testing.T) {
	t.Parallel()
	// Null
	nb := NullableBool{Valid: false}
	data, err := json.Marshal(nb)
	if err != nil {
		t.Fatalf("MarshalJSON(null) error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON(null) = %s", data)
	}

	var got NullableBool
	if err := json.Unmarshal([]byte("null"), &got); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if got.Valid {
		t.Error("UnmarshalJSON(null) Valid = true")
	}

	// Valid
	nb = NewNullableBool(true)
	data, _ = json.Marshal(nb)
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid || !got.Bool {
		t.Errorf("JSON round-trip: Valid=%v Bool=%v", got.Valid, got.Bool)
	}
}

// ============================================================
// JSONData
// ============================================================

func TestJSONData_New(t *testing.T) {
	t.Parallel()
	jd := NewJSONData(map[string]any{"key": "value"})
	if !jd.Valid {
		t.Error("NewJSONData.Valid = false")
	}
	if jd.IsZero() {
		t.Error("NewJSONData.IsZero() = true")
	}
}

func TestJSONData_ZeroValue(t *testing.T) {
	t.Parallel()
	var jd JSONData
	if !jd.IsZero() {
		t.Error("zero JSONData.IsZero() = false")
	}
}

func TestJSONData_Value(t *testing.T) {
	t.Parallel()
	// Null
	jd := JSONData{Valid: false}
	v, err := jd.Value()
	if err != nil {
		t.Fatalf("null Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null Value() = %v", v)
	}

	// Valid
	jd = NewJSONData(map[string]any{"a": float64(1)})
	v, err = jd.Value()
	if err != nil {
		t.Fatalf("valid Value() error = %v", err)
	}
	// Value returns json.Marshal output ([]byte)
	b, ok := v.([]byte)
	if !ok {
		t.Fatalf("valid Value() type = %T, want []byte", v)
	}
	if string(b) != `{"a":1}` {
		t.Errorf("valid Value() = %s", b)
	}
}

func TestJSONData_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantOk  bool
		wantErr bool
	}{
		{name: "nil", input: nil, wantOk: false, wantErr: false},
		{name: "string", input: `{"key":"val"}`, wantOk: true, wantErr: false},
		{name: "bytes", input: []byte(`{"key":"val"}`), wantOk: true, wantErr: false},
		{name: "empty string", input: "", wantOk: false, wantErr: false},
		{name: "empty bytes", input: []byte{}, wantOk: false, wantErr: false},
		{name: "invalid json", input: `{bad`, wantOk: false, wantErr: true},
		{name: "int", input: 42, wantOk: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var jd JSONData
			err := jd.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && jd.Valid != tt.wantOk {
				t.Errorf("Scan(%v) Valid = %v, want %v", tt.input, jd.Valid, tt.wantOk)
			}
		})
	}
}

func TestJSONData_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	jd := NewJSONData(map[string]any{"x": float64(42)})
	data, err := json.Marshal(jd)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}

	var got JSONData
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid {
		t.Error("round-trip Valid = false")
	}
	// Check data content
	m, ok := got.Data.(map[string]any)
	if !ok {
		t.Fatalf("round-trip Data type = %T, want map", got.Data)
	}
	if m["x"] != float64(42) {
		t.Errorf("round-trip Data[x] = %v", m["x"])
	}
}

func TestJSONData_JSON_Null(t *testing.T) {
	t.Parallel()
	jd := JSONData{Valid: false}
	data, err := json.Marshal(jd)
	if err != nil {
		t.Fatalf("MarshalJSON(null) error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON(null) = %s", data)
	}

	var got JSONData
	if err := json.Unmarshal([]byte("null"), &got); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if got.Valid {
		t.Error("UnmarshalJSON(null) Valid = true")
	}
}

func TestJSONData_String(t *testing.T) {
	t.Parallel()
	// Null
	jd := JSONData{Valid: false}
	if s := jd.String(); s != "" {
		t.Errorf("null String() = %q, want empty", s)
	}

	// Valid
	jd = NewJSONData(map[string]any{"a": float64(1)})
	s := jd.String()
	if s != `{"a":1}` {
		t.Errorf("valid String() = %q, want %q", s, `{"a":1}`)
	}
}
