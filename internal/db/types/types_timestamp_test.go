package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTimestamp(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ts := NewTimestamp(now)
	if !ts.Valid {
		t.Error("NewTimestamp.Valid = false")
	}
	if ts.IsZero() {
		t.Error("NewTimestamp.IsZero() = true")
	}
	// Should be stored in UTC
	if ts.Time.Location() != time.UTC {
		t.Errorf("NewTimestamp.Time.Location() = %v, want UTC", ts.Time.Location())
	}
}

func TestTimestampNow(t *testing.T) {
	t.Parallel()
	before := time.Now().Add(-time.Second)
	ts := TimestampNow()
	after := time.Now().Add(time.Second)

	if !ts.Valid {
		t.Fatal("TimestampNow().Valid = false")
	}
	if ts.Time.Before(before) || ts.Time.After(after) {
		t.Errorf("TimestampNow() = %v, want between %v and %v", ts.Time, before, after)
	}
}

func TestTimestamp_String(t *testing.T) {
	t.Parallel()
	// Null
	ts := Timestamp{Valid: false}
	if ts.String() != "null" {
		t.Errorf("null.String() = %q, want %q", ts.String(), "null")
	}

	// Valid - should be RFC3339 in UTC
	ref := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	ts = NewTimestamp(ref)
	want := "2024-06-15T12:30:00Z"
	if ts.String() != want {
		t.Errorf("String() = %q, want %q", ts.String(), want)
	}
}

func TestTimestamp_IsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ts   Timestamp
		want bool
	}{
		{name: "invalid", ts: Timestamp{Valid: false}, want: true},
		{name: "valid zero time", ts: Timestamp{Time: time.Time{}, Valid: true}, want: true},
		{name: "valid non-zero", ts: TimestampNow(), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.ts.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimestamp_Value(t *testing.T) {
	t.Parallel()
	// Null
	ts := Timestamp{Valid: false}
	v, err := ts.Value()
	if err != nil {
		t.Fatalf("null Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null Value() = %v, want nil", v)
	}

	// Valid
	ref := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	ts = NewTimestamp(ref)
	v, err = ts.Value()
	if err != nil {
		t.Fatalf("valid Value() error = %v", err)
	}
	vt, ok := v.(time.Time)
	if !ok {
		t.Fatalf("valid Value() type = %T, want time.Time", v)
	}
	if !vt.Equal(ref) {
		t.Errorf("valid Value() = %v, want %v", vt, ref)
	}
}

func TestTimestamp_Scan(t *testing.T) {
	t.Parallel()
	ref := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)
	tests := []struct {
		name    string
		input   any
		wantOk  bool
		wantErr bool
		wantUTC time.Time
	}{
		{name: "nil", input: nil, wantOk: false, wantErr: false},
		{name: "time.Time", input: ref, wantOk: true, wantErr: false, wantUTC: ref},
		{name: "empty string", input: "", wantOk: false, wantErr: false},
		// Strict formats
		{name: "RFC3339 UTC", input: "2024-06-15T12:30:45Z", wantOk: true, wantErr: false, wantUTC: ref},
		{name: "RFC3339 offset", input: "2024-06-15T14:30:45+02:00", wantOk: true, wantErr: false, wantUTC: ref},
		// Legacy formats
		{name: "MySQL datetime", input: "2024-06-15 12:30:45", wantOk: true, wantErr: false, wantUTC: ref},
		{name: "date only", input: "2024-06-15", wantOk: true, wantErr: false, wantUTC: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)},
		// []byte path
		{name: "bytes RFC3339", input: []byte("2024-06-15T12:30:45Z"), wantOk: true, wantErr: false, wantUTC: ref},
		// Errors
		{name: "unparseable", input: "not-a-date", wantOk: false, wantErr: true},
		{name: "int", input: 42, wantOk: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ts Timestamp
			err := ts.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if ts.Valid != tt.wantOk {
				t.Errorf("Scan(%v) Valid = %v, want %v", tt.input, ts.Valid, tt.wantOk)
			}
			if tt.wantOk && !ts.Time.Equal(tt.wantUTC) {
				t.Errorf("Scan(%v) Time = %v, want %v", tt.input, ts.Time, tt.wantUTC)
			}
		})
	}
}

func TestTimestamp_Scan_NonUTCConvertedToUTC(t *testing.T) {
	t.Parallel()
	// time.Time in non-UTC location
	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation error = %v", err)
	}
	local := time.Date(2024, 6, 15, 8, 30, 0, 0, eastern)
	var ts Timestamp
	if err := ts.Scan(local); err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	if ts.Time.Location() != time.UTC {
		t.Errorf("Scan(non-UTC) location = %v, want UTC", ts.Time.Location())
	}
	if !ts.Time.Equal(local.UTC()) {
		t.Errorf("Scan(non-UTC) time = %v, want %v", ts.Time, local.UTC())
	}
}

func TestTimestamp_MarshalJSON(t *testing.T) {
	t.Parallel()
	// Null
	ts := Timestamp{Valid: false}
	data, err := ts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON(null) error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON(null) = %s", data)
	}

	// Valid
	ref := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	ts = NewTimestamp(ref)
	data, err = ts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	if string(data) != `"2024-06-15T12:00:00Z"` {
		t.Errorf("MarshalJSON = %s", data)
	}
}

func TestTimestamp_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantErr bool
	}{
		{name: "null", input: "null", wantOk: false, wantErr: false},
		{name: "RFC3339 UTC", input: `"2024-06-15T12:00:00Z"`, wantOk: true, wantErr: false},
		{name: "RFC3339 offset", input: `"2024-06-15T14:00:00+02:00"`, wantOk: true, wantErr: false},
		// Strict: legacy formats are rejected
		{name: "MySQL format rejected", input: `"2024-06-15 12:00:00"`, wantOk: false, wantErr: true},
		{name: "date only rejected", input: `"2024-06-15"`, wantOk: false, wantErr: true},
		// Bad input
		{name: "not a string", input: `42`, wantOk: false, wantErr: true},
		{name: "garbage", input: `"not-a-date"`, wantOk: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var ts Timestamp
			err := json.Unmarshal([]byte(tt.input), &ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && ts.Valid != tt.wantOk {
				t.Errorf("UnmarshalJSON(%s) Valid = %v, want %v", tt.input, ts.Valid, tt.wantOk)
			}
		})
	}
}

func TestTimestamp_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	ref := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)
	ts := NewTimestamp(ref)

	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}

	var got Timestamp
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid {
		t.Fatal("round-trip Valid = false")
	}
	// Note: MarshalJSON uses RFC3339 which has second precision,
	// so sub-second precision is lost. Both should match at second precision.
	if got.Time.Unix() != ref.Unix() {
		t.Errorf("round-trip Time = %v, want %v", got.Time, ref)
	}
}

func TestTimestamp_Before(t *testing.T) {
	t.Parallel()
	earlier := NewTimestamp(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	later := NewTimestamp(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
	null := Timestamp{Valid: false}

	if !earlier.Before(later) {
		t.Error("earlier.Before(later) = false")
	}
	if later.Before(earlier) {
		t.Error("later.Before(earlier) = true")
	}
	// Null timestamps: Before always returns false
	if null.Before(later) {
		t.Error("null.Before(valid) = true")
	}
	if earlier.Before(null) {
		t.Error("valid.Before(null) = true")
	}
}

func TestTimestamp_After(t *testing.T) {
	t.Parallel()
	earlier := NewTimestamp(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	later := NewTimestamp(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
	null := Timestamp{Valid: false}

	if !later.After(earlier) {
		t.Error("later.After(earlier) = false")
	}
	if earlier.After(later) {
		t.Error("earlier.After(later) = true")
	}
	if null.After(earlier) {
		t.Error("null.After(valid) = true")
	}
}

func TestTimestamp_UTC(t *testing.T) {
	t.Parallel()
	ref := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	ts := NewTimestamp(ref)
	if !ts.UTC().Equal(ref) {
		t.Errorf("UTC() = %v, want %v", ts.UTC(), ref)
	}
}

func TestTimestamp_Add(t *testing.T) {
	t.Parallel()
	ref := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	ts := NewTimestamp(ref)

	added := ts.Add(24 * time.Hour)
	if !added.Valid {
		t.Error("Add().Valid = false")
	}
	want := ref.Add(24 * time.Hour)
	if !added.Time.Equal(want) {
		t.Errorf("Add(24h) = %v, want %v", added.Time, want)
	}

	// Null + Add = null
	null := Timestamp{Valid: false}
	result := null.Add(time.Hour)
	if result.Valid {
		t.Error("null.Add().Valid = true")
	}
}

func TestTimestamp_Sub(t *testing.T) {
	t.Parallel()
	earlier := NewTimestamp(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC))
	later := NewTimestamp(time.Date(2024, 6, 16, 12, 0, 0, 0, time.UTC))

	d := later.Sub(earlier)
	if d != 24*time.Hour {
		t.Errorf("Sub() = %v, want 24h", d)
	}

	// Null sub returns 0
	null := Timestamp{Valid: false}
	if d := null.Sub(earlier); d != 0 {
		t.Errorf("null.Sub() = %v, want 0", d)
	}
	if d := earlier.Sub(null); d != 0 {
		t.Errorf("valid.Sub(null) = %v, want 0", d)
	}
}
