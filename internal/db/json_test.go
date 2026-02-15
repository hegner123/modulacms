package db

import (
	"database/sql"
	"encoding/json"
	"math"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// NullInt32
// ---------------------------------------------------------------------------

func TestNullInt32_MarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullInt32
		want string
	}{
		{name: "valid positive", in: NullInt32{sql.NullInt32{Int32: 42, Valid: true}}, want: "42"},
		{name: "valid negative", in: NullInt32{sql.NullInt32{Int32: -7, Valid: true}}, want: "-7"},
		{name: "valid zero", in: NullInt32{sql.NullInt32{Int32: 0, Valid: true}}, want: "0"},
		{name: "valid max int32", in: NullInt32{sql.NullInt32{Int32: math.MaxInt32, Valid: true}}, want: "2147483647"},
		{name: "valid min int32", in: NullInt32{sql.NullInt32{Int32: math.MinInt32, Valid: true}}, want: "-2147483648"},
		{name: "null", in: NullInt32{sql.NullInt32{Int32: 0, Valid: false}}, want: "null"},
		{name: "null with leftover value", in: NullInt32{sql.NullInt32{Int32: 99, Valid: false}}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.in.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNullInt32_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantVal   int32
		wantValid bool
		wantErr   bool
	}{
		{name: "valid positive", input: "42", wantVal: 42, wantValid: true},
		{name: "valid negative", input: "-7", wantVal: -7, wantValid: true},
		{name: "valid zero", input: "0", wantVal: 0, wantValid: true},
		{name: "valid max int32", input: "2147483647", wantVal: math.MaxInt32, wantValid: true},
		{name: "valid min int32", input: "-2147483648", wantVal: math.MinInt32, wantValid: true},
		{name: "null", input: "null", wantVal: 0, wantValid: false},
		{name: "error on string", input: `"hello"`, wantErr: true},
		{name: "error on boolean", input: "true", wantErr: true},
		{name: "error on object", input: `{"x":1}`, wantErr: true},
		{name: "error on array", input: `[1,2]`, wantErr: true},
		{name: "error on malformed", input: `{bad`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var n NullInt32
			err := n.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("UnmarshalJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() unexpected error: %v", err)
			}
			if n.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", n.Valid, tt.wantValid)
			}
			if n.Int32 != tt.wantVal {
				t.Errorf("Int32 = %d, want %d", n.Int32, tt.wantVal)
			}
		})
	}
}

func TestNullInt32_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullInt32
	}{
		{name: "valid value", in: NullInt32{sql.NullInt32{Int32: 123, Valid: true}}},
		{name: "valid zero", in: NullInt32{sql.NullInt32{Int32: 0, Valid: true}}},
		{name: "null", in: NullInt32{sql.NullInt32{Int32: 0, Valid: false}}},
		{name: "max int32", in: NullInt32{sql.NullInt32{Int32: math.MaxInt32, Valid: true}}},
		{name: "min int32", in: NullInt32{sql.NullInt32{Int32: math.MinInt32, Valid: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			var got NullInt32
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if got.Valid != tt.in.Valid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.in.Valid)
			}
			if got.Valid && got.Int32 != tt.in.Int32 {
				t.Errorf("Int32 = %d, want %d", got.Int32, tt.in.Int32)
			}
		})
	}
}

func TestNullInt32_UnmarshalJSON_ClearsValidOnNull(t *testing.T) {
	t.Parallel()
	// Start with a valid value, then unmarshal null -- Valid must become false
	n := NullInt32{sql.NullInt32{Int32: 42, Valid: true}}
	if err := n.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}
	if n.Valid {
		t.Error("expected Valid=false after unmarshalling null")
	}
}

// ---------------------------------------------------------------------------
// NullInt64
// ---------------------------------------------------------------------------

func TestNullInt64_MarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullInt64
		want string
	}{
		{name: "valid positive", in: NullInt64{sql.NullInt64{Int64: 100, Valid: true}}, want: "100"},
		{name: "valid negative", in: NullInt64{sql.NullInt64{Int64: -999, Valid: true}}, want: "-999"},
		{name: "valid zero", in: NullInt64{sql.NullInt64{Int64: 0, Valid: true}}, want: "0"},
		{name: "valid max int64", in: NullInt64{sql.NullInt64{Int64: math.MaxInt64, Valid: true}}, want: "9223372036854775807"},
		{name: "valid min int64", in: NullInt64{sql.NullInt64{Int64: math.MinInt64, Valid: true}}, want: "-9223372036854775808"},
		{name: "null", in: NullInt64{sql.NullInt64{Int64: 0, Valid: false}}, want: "null"},
		{name: "null with leftover value", in: NullInt64{sql.NullInt64{Int64: 55, Valid: false}}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.in.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNullInt64_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantVal   int64
		wantValid bool
		wantErr   bool
	}{
		{name: "valid positive", input: "100", wantVal: 100, wantValid: true},
		{name: "valid negative", input: "-999", wantVal: -999, wantValid: true},
		{name: "valid zero", input: "0", wantVal: 0, wantValid: true},
		{name: "valid max int64", input: "9223372036854775807", wantVal: math.MaxInt64, wantValid: true},
		{name: "valid min int64", input: "-9223372036854775808", wantVal: math.MinInt64, wantValid: true},
		{name: "null", input: "null", wantVal: 0, wantValid: false},
		{name: "error on string", input: `"hello"`, wantErr: true},
		{name: "error on boolean", input: "false", wantErr: true},
		{name: "error on object", input: `{}`, wantErr: true},
		{name: "error on array", input: `[]`, wantErr: true},
		{name: "error on malformed", input: `[bad`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var n NullInt64
			err := n.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("UnmarshalJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() unexpected error: %v", err)
			}
			if n.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", n.Valid, tt.wantValid)
			}
			if n.Int64 != tt.wantVal {
				t.Errorf("Int64 = %d, want %d", n.Int64, tt.wantVal)
			}
		})
	}
}

func TestNullInt64_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullInt64
	}{
		{name: "valid value", in: NullInt64{sql.NullInt64{Int64: 456, Valid: true}}},
		{name: "valid zero", in: NullInt64{sql.NullInt64{Int64: 0, Valid: true}}},
		{name: "null", in: NullInt64{sql.NullInt64{Int64: 0, Valid: false}}},
		{name: "max int64", in: NullInt64{sql.NullInt64{Int64: math.MaxInt64, Valid: true}}},
		{name: "min int64", in: NullInt64{sql.NullInt64{Int64: math.MinInt64, Valid: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			var got NullInt64
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if got.Valid != tt.in.Valid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.in.Valid)
			}
			if got.Valid && got.Int64 != tt.in.Int64 {
				t.Errorf("Int64 = %d, want %d", got.Int64, tt.in.Int64)
			}
		})
	}
}

func TestNullInt64_UnmarshalJSON_ClearsValidOnNull(t *testing.T) {
	t.Parallel()
	n := NullInt64{sql.NullInt64{Int64: 999, Valid: true}}
	if err := n.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}
	if n.Valid {
		t.Error("expected Valid=false after unmarshalling null")
	}
}

// ---------------------------------------------------------------------------
// NullString
// ---------------------------------------------------------------------------

func TestNullString_MarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullString
		want string
	}{
		{name: "valid non-empty", in: NullString{sql.NullString{String: "hello", Valid: true}}, want: `"hello"`},
		{name: "valid empty string", in: NullString{sql.NullString{String: "", Valid: true}}, want: `""`},
		{name: "valid with special chars", in: NullString{sql.NullString{String: "line\nnewline", Valid: true}}, want: `"line\nnewline"`},
		{name: "valid with unicode", in: NullString{sql.NullString{String: "caf\u00e9", Valid: true}}, want: "\"caf\u00e9\""},
		{name: "valid with quotes", in: NullString{sql.NullString{String: `say "hi"`, Valid: true}}, want: `"say \"hi\""`},
		{name: "null", in: NullString{sql.NullString{String: "", Valid: false}}, want: "null"},
		{name: "null with leftover value", in: NullString{sql.NullString{String: "leftover", Valid: false}}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.in.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNullString_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantVal   string
		wantValid bool
		wantErr   bool
	}{
		{name: "valid string", input: `"hello"`, wantVal: "hello", wantValid: true},
		{name: "valid empty string", input: `""`, wantVal: "", wantValid: true},
		{name: "valid string with escape", input: `"line\nnewline"`, wantVal: "line\nnewline", wantValid: true},
		{name: "valid unicode", input: `"caf\u00e9"`, wantVal: "caf\u00e9", wantValid: true},
		{name: "null", input: "null", wantVal: "", wantValid: false},
		{name: "error on number", input: "42", wantErr: true},
		{name: "error on boolean", input: "true", wantErr: true},
		{name: "error on object", input: `{"x":1}`, wantErr: true},
		{name: "error on array", input: `["a"]`, wantErr: true},
		{name: "error on malformed", input: `"unterminated`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var n NullString
			err := n.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("UnmarshalJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() unexpected error: %v", err)
			}
			if n.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", n.Valid, tt.wantValid)
			}
			if n.String != tt.wantVal {
				t.Errorf("String = %q, want %q", n.String, tt.wantVal)
			}
		})
	}
}

func TestNullString_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullString
	}{
		{name: "non-empty string", in: NullString{sql.NullString{String: "round trip", Valid: true}}},
		{name: "empty string", in: NullString{sql.NullString{String: "", Valid: true}}},
		{name: "null", in: NullString{sql.NullString{String: "", Valid: false}}},
		{name: "special characters", in: NullString{sql.NullString{String: "tab\there\nnewline", Valid: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			var got NullString
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if got.Valid != tt.in.Valid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.in.Valid)
			}
			if got.Valid && got.String != tt.in.String {
				t.Errorf("String = %q, want %q", got.String, tt.in.String)
			}
		})
	}
}

func TestNullString_UnmarshalJSON_ClearsValidOnNull(t *testing.T) {
	t.Parallel()
	n := NullString{sql.NullString{String: "was set", Valid: true}}
	if err := n.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}
	if n.Valid {
		t.Error("expected Valid=false after unmarshalling null")
	}
}

// ---------------------------------------------------------------------------
// NullTime
// ---------------------------------------------------------------------------

func TestNullTime_MarshalJSON(t *testing.T) {
	t.Parallel()
	utcTime := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name string
		in   NullTime
		want string
	}{
		{name: "valid UTC time", in: NullTime{sql.NullTime{Time: utcTime, Valid: true}}, want: `"2024-06-15T12:30:00Z"`},
		{name: "valid zero time", in: NullTime{sql.NullTime{Time: zeroTime, Valid: true}}, want: `"0001-01-01T00:00:00Z"`},
		{name: "null", in: NullTime{sql.NullTime{Time: time.Time{}, Valid: false}}, want: "null"},
		{name: "null with leftover value", in: NullTime{sql.NullTime{Time: utcTime, Valid: false}}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.in.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNullTime_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	utcTime := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name      string
		input     string
		wantTime  time.Time
		wantValid bool
		wantErr   bool
	}{
		{name: "valid RFC3339 UTC", input: `"2024-06-15T12:30:00Z"`, wantTime: utcTime, wantValid: true},
		{name: "valid RFC3339 with offset", input: `"2024-06-15T14:30:00+02:00"`, wantTime: utcTime, wantValid: true},
		{name: "valid zero time", input: `"0001-01-01T00:00:00Z"`, wantTime: zeroTime, wantValid: true},
		{name: "valid with nanoseconds", input: `"2024-06-15T12:30:00.123456789Z"`, wantTime: time.Date(2024, 6, 15, 12, 30, 0, 123456789, time.UTC), wantValid: true},
		{name: "null", input: "null", wantTime: time.Time{}, wantValid: false},
		{name: "error on number", input: "42", wantErr: true},
		{name: "error on boolean", input: "true", wantErr: true},
		{name: "error on invalid time string", input: `"not-a-time"`, wantErr: true},
		{name: "error on malformed JSON", input: `{bad`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var n NullTime
			err := n.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("UnmarshalJSON() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() unexpected error: %v", err)
			}
			if n.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", n.Valid, tt.wantValid)
			}
			if n.Valid && !n.Time.Equal(tt.wantTime) {
				t.Errorf("Time = %v, want %v", n.Time, tt.wantTime)
			}
		})
	}
}

func TestNullTime_RoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   NullTime
	}{
		{name: "UTC time", in: NullTime{sql.NullTime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}}},
		{name: "zero time", in: NullTime{sql.NullTime{Time: time.Time{}, Valid: true}}},
		{name: "null", in: NullTime{sql.NullTime{Time: time.Time{}, Valid: false}}},
		{name: "with nanoseconds", in: NullTime{sql.NullTime{Time: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC), Valid: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}
			var got NullTime
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}
			if got.Valid != tt.in.Valid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.in.Valid)
			}
			if got.Valid && !got.Time.Equal(tt.in.Time) {
				t.Errorf("Time = %v, want %v", got.Time, tt.in.Time)
			}
		})
	}
}

func TestNullTime_UnmarshalJSON_ClearsValidOnNull(t *testing.T) {
	t.Parallel()
	n := NullTime{sql.NullTime{Time: time.Now(), Valid: true}}
	if err := n.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("UnmarshalJSON() error: %v", err)
	}
	if n.Valid {
		t.Error("expected Valid=false after unmarshalling null")
	}
}

// ---------------------------------------------------------------------------
// Cross-type: JSON struct embedding
// ---------------------------------------------------------------------------

func TestNullTypes_InStruct(t *testing.T) {
	t.Parallel()

	type record struct {
		Count   NullInt32  `json:"count"`
		Total   NullInt64  `json:"total"`
		Label   NullString `json:"label"`
		Created NullTime   `json:"created"`
	}

	ts := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		r := record{
			Count:   NullInt32{sql.NullInt32{Int32: 5, Valid: true}},
			Total:   NullInt64{sql.NullInt64{Int64: 1000, Valid: true}},
			Label:   NullString{sql.NullString{String: "test", Valid: true}},
			Created: NullTime{sql.NullTime{Time: ts, Valid: true}},
		}
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error: %v", err)
		}
		var got record
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}
		if got.Count.Int32 != 5 {
			t.Errorf("Count.Int32 = %d, want 5", got.Count.Int32)
		}
		if got.Total.Int64 != 1000 {
			t.Errorf("Total.Int64 = %d, want 1000", got.Total.Int64)
		}
		if got.Label.String != "test" {
			t.Errorf("Label.String = %q, want %q", got.Label.String, "test")
		}
		if !got.Created.Time.Equal(ts) {
			t.Errorf("Created.Time = %v, want %v", got.Created.Time, ts)
		}
	})

	t.Run("all null", func(t *testing.T) {
		t.Parallel()
		r := record{}
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal() error: %v", err)
		}
		// All fields should be null in JSON
		want := `{"count":null,"total":null,"label":null,"created":null}`
		if string(data) != want {
			t.Errorf("Marshal() = %s, want %s", data, want)
		}
		var got record
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}
		if got.Count.Valid || got.Total.Valid || got.Label.Valid || got.Created.Valid {
			t.Error("expected all fields to be invalid after unmarshalling nulls")
		}
	})

	t.Run("mixed valid and null", func(t *testing.T) {
		t.Parallel()
		input := `{"count":10,"total":null,"label":"mixed","created":null}`
		var got record
		if err := json.Unmarshal([]byte(input), &got); err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}
		if !got.Count.Valid || got.Count.Int32 != 10 {
			t.Errorf("Count = {%d, %v}, want {10, true}", got.Count.Int32, got.Count.Valid)
		}
		if got.Total.Valid {
			t.Error("Total should be invalid (null)")
		}
		if !got.Label.Valid || got.Label.String != "mixed" {
			t.Errorf("Label = {%q, %v}, want {%q, true}", got.Label.String, got.Label.Valid, "mixed")
		}
		if got.Created.Valid {
			t.Error("Created should be invalid (null)")
		}
	})
}
