package db

import (
	"database/sql"
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestDBTableString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input DBTable
		want  string
	}{
		{name: "route table", input: Route, want: "routes"},
		{name: "user table", input: User, want: "users"},
		{name: "media table", input: MediaT, want: "media"},
		{name: "empty table", input: DBTable(""), want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DBTableString(tt.input)
			if got != tt.want {
				t.Errorf("DBTableString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStringDBTable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  DBTable
	}{
		{name: "routes string", input: "routes", want: Route},
		{name: "users string", input: "users", want: User},
		{name: "empty string", input: "", want: DBTable("")},
		{name: "arbitrary string", input: "nonexistent", want: DBTable("nonexistent")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringDBTable(tt.input)
			if got != tt.want {
				t.Errorf("StringDBTable(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDBTableRoundTrip(t *testing.T) {
	t.Parallel()
	// Verify that converting to string and back yields the same DBTable
	tables := []DBTable{Route, User, MediaT, Permission, Role, Session, Token, Table, Datatype, Field}
	for _, tbl := range tables {
		t.Run(string(tbl), func(t *testing.T) {
			t.Parallel()
			s := DBTableString(tbl)
			back := StringDBTable(s)
			if back != tbl {
				t.Errorf("round-trip failed: %q -> %q -> %q", tbl, s, back)
			}
		})
	}
}

func TestTimeToNullString(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	got := TimeToNullString(now)
	if !got.Valid {
		t.Fatal("expected Valid=true")
	}
	if got.String != now.String() {
		t.Errorf("got %q, want %q", got.String, now.String())
	}
}

func TestTimeToNullString_ZeroTime(t *testing.T) {
	t.Parallel()
	zero := time.Time{}
	got := TimeToNullString(zero)
	if !got.Valid {
		t.Fatal("expected Valid=true even for zero time")
	}
	if got.String != zero.String() {
		t.Errorf("got %q, want %q", got.String, zero.String())
	}
}

func TestIntToNullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input int
		want  int64
	}{
		{name: "positive", input: 42, want: 42},
		{name: "zero", input: 0, want: 0},
		{name: "negative", input: -7, want: -7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IntToNullInt64(tt.input)
			if !got.Valid {
				t.Fatal("expected Valid=true")
			}
			if got.Int64 != tt.want {
				t.Errorf("IntToNullInt64(%d).Int64 = %d, want %d", tt.input, got.Int64, tt.want)
			}
		})
	}
}

func TestInt64ToNullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input int64
	}{
		{name: "positive", input: 100},
		{name: "zero", input: 0},
		{name: "negative", input: -999},
		{name: "max int64", input: math.MaxInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Int64ToNullInt64(tt.input)
			if !got.Valid {
				t.Fatal("expected Valid=true")
			}
			if got.Int64 != tt.input {
				t.Errorf("got %d, want %d", got.Int64, tt.input)
			}
		})
	}
}

func TestInt64ToNullInt32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input int64
		want  int32
	}{
		{name: "fits in int32", input: 42, want: 42},
		{name: "zero", input: 0, want: 0},
		{name: "negative", input: -10, want: -10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Int64ToNullInt32(tt.input)
			if !got.Valid {
				t.Fatal("expected Valid=true")
			}
			if got.Int32 != tt.want {
				t.Errorf("got %d, want %d", got.Int32, tt.want)
			}
		})
	}
}

func TestNullInt32ToNullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     sql.NullInt32
		wantInt64 int64
		wantValid bool
	}{
		{name: "valid positive", input: sql.NullInt32{Int32: 42, Valid: true}, wantInt64: 42, wantValid: true},
		{name: "valid negative", input: sql.NullInt32{Int32: -5, Valid: true}, wantInt64: -5, wantValid: true},
		{name: "valid zero", input: sql.NullInt32{Int32: 0, Valid: true}, wantInt64: 0, wantValid: true},
		{name: "null", input: sql.NullInt32{Int32: 0, Valid: false}, wantInt64: 0, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullInt32ToNullInt64(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Int64 != tt.wantInt64 {
				t.Errorf("Int64 = %d, want %d", got.Int64, tt.wantInt64)
			}
		})
	}
}

func TestNullInt64ToNullInt32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     sql.NullInt64
		wantInt32 int32
		wantValid bool
	}{
		{name: "valid positive", input: sql.NullInt64{Int64: 42, Valid: true}, wantInt32: 42, wantValid: true},
		{name: "valid negative", input: sql.NullInt64{Int64: -5, Valid: true}, wantInt32: -5, wantValid: true},
		{name: "null", input: sql.NullInt64{Int64: 0, Valid: false}, wantInt32: 0, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullInt64ToNullInt32(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Int32 != tt.wantInt32 {
				t.Errorf("Int32 = %d, want %d", got.Int32, tt.wantInt32)
			}
		})
	}
}

func TestBoolToNullBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input bool
	}{
		{name: "true", input: true},
		{name: "false", input: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BoolToNullBool(tt.input)
			if !got.Valid {
				t.Fatal("expected Valid=true")
			}
			if got.Bool != tt.input {
				t.Errorf("got %v, want %v", got.Bool, tt.input)
			}
		})
	}
}

func TestTimeToNullTime(t *testing.T) {
	t.Parallel()
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := TimeToNullTime(now)
	if !got.Valid {
		t.Fatal("expected Valid=true")
	}
	if !got.Time.Equal(now) {
		t.Errorf("got %v, want %v", got.Time, now)
	}
}

func TestStringToNTime(t *testing.T) {
	t.Parallel()
	// StringToNTime parses Unix timestamps and RFC3339 strings into sql.NullTime.
	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{name: "RFC3339 string parses", input: "2024-06-15T12:30:00Z", wantValid: true},
		{name: "empty string", input: "", wantValid: false},
		{name: "garbage", input: "not-a-date", wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToNTime(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("StringToNTime(%q).Valid = %v, want %v", tt.input, got.Valid, tt.wantValid)
			}
		})
	}
}

func TestNullTimeToString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input sql.NullTime
		want  string
	}{
		// NullTime.Value() returns the time as a time.Time, which is not a string,
		// so the function returns "" for valid times too (the type assertion to string fails).
		// This tests the actual behavior of the function.
		{name: "null time", input: sql.NullTime{Valid: false}, want: ""},
		{name: "valid time returns empty because Value is time.Time not string",
			input: sql.NullTime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
			want:  ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullTimeToString(tt.input)
			if got != tt.want {
				t.Errorf("NullTimeToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringToInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "positive number", input: "42", want: 42},
		{name: "zero", input: "0", want: 0},
		{name: "negative", input: "-100", want: -100},
		{name: "empty string", input: "", want: 0},
		{name: "non-numeric", input: "abc", want: 0},
		{name: "float string", input: "3.14", want: 0},
		{name: "max int64", input: "9223372036854775807", want: math.MaxInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToInt64(tt.input)
			if got != tt.want {
				t.Errorf("StringToInt64(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestStringToBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "true", input: "true", want: true},
		{name: "false", input: "false", want: false},
		{name: "1", input: "1", want: true},
		{name: "0", input: "0", want: false},
		{name: "empty", input: "", want: false},
		{name: "garbage", input: "yes", want: false},
		{name: "True", input: "True", want: true},
		{name: "FALSE", input: "FALSE", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToBool(tt.input)
			if got != tt.want {
				t.Errorf("StringToBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "true", input: "true", want: true},
		{name: "false", input: "false", want: false},
		{name: "invalid", input: "maybe", want: false},
		{name: "empty", input: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseBool(tt.input)
			if got != tt.want {
				t.Errorf("ParseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	t.Parallel()
	t.Run("valid RFC3339", func(t *testing.T) {
		t.Parallel()
		got := ParseTime("2024-06-15T12:30:00Z")
		want := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("ParseTime() = %v, want %v", got, want)
		}
	})

	t.Run("invalid returns now-ish time", func(t *testing.T) {
		t.Parallel()
		before := time.Now().Add(-time.Second)
		got := ParseTime("not-a-time")
		after := time.Now().Add(time.Second)
		if got.Before(before) || got.After(after) {
			t.Errorf("ParseTime(invalid) should return approximately time.Now(), got %v", got)
		}
	})

	t.Run("empty string returns now-ish time", func(t *testing.T) {
		t.Parallel()
		before := time.Now().Add(-time.Second)
		got := ParseTime("")
		after := time.Now().Add(time.Second)
		if got.Before(before) || got.After(after) {
			t.Errorf("ParseTime(\"\") should return approximately time.Now(), got %v", got)
		}
	})
}

func TestAssertString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{name: "actual string", input: "hello", want: "hello"},
		{name: "empty string", input: "", want: ""},
		{name: "integer falls back to Sprint", input: 42, want: "42"},
		{name: "nil falls back to Sprint", input: nil, want: "<nil>"},
		{name: "bool falls back to Sprint", input: true, want: "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AssertString(tt.input)
			if got != tt.want {
				t.Errorf("AssertString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAssertInteger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input any
		want  int
	}{
		{name: "actual int", input: 42, want: 42},
		{name: "zero int", input: 0, want: 0},
		{name: "negative int", input: -5, want: -5},
		{name: "string returns 0", input: "42", want: 0},
		{name: "nil returns 0", input: nil, want: 0},
		{name: "int64 returns 0", input: int64(42), want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AssertInteger(tt.input)
			if got != tt.want {
				t.Errorf("AssertInteger(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestAssertInt32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input any
		want  int32
	}{
		{name: "actual int32", input: int32(42), want: 42},
		{name: "wrong type returns 0", input: int64(42), want: 0},
		{name: "nil returns 0", input: nil, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AssertInt32(tt.input)
			if got != tt.want {
				t.Errorf("AssertInt32(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestAssertInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input any
		want  int64
	}{
		{name: "actual int64", input: int64(42), want: 42},
		{name: "wrong type returns 0", input: int32(42), want: 0},
		{name: "nil returns 0", input: nil, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AssertInt64(tt.input)
			if got != tt.want {
				t.Errorf("AssertInt64(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestStringToNullString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantStr   string
		wantValid bool
	}{
		{name: "non-empty string", input: "hello", wantStr: "hello", wantValid: true},
		{name: "empty string is invalid", input: "", wantStr: "", wantValid: false},
		{name: "null literal is invalid", input: "null", wantStr: "null", wantValid: false},
		{name: "whitespace is valid", input: " ", wantStr: " ", wantValid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToNullString(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.String != tt.wantStr {
				t.Errorf("String = %q, want %q", got.String, tt.wantStr)
			}
		})
	}
}

func TestStringToNullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantInt64 int64
		wantValid bool
	}{
		{name: "valid number", input: "42", wantInt64: 42, wantValid: true},
		{name: "zero", input: "0", wantInt64: 0, wantValid: true},
		{name: "negative", input: "-10", wantInt64: -10, wantValid: true},
		{name: "empty", input: "", wantInt64: 0, wantValid: false},
		{name: "non-numeric", input: "abc", wantInt64: 0, wantValid: false},
		{name: "float", input: "3.14", wantInt64: 0, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToNullInt64(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Int64 != tt.wantInt64 {
				t.Errorf("Int64 = %d, want %d", got.Int64, tt.wantInt64)
			}
		})
	}
}

func TestStringToNullInt32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantInt32 int32
		wantValid bool
	}{
		{name: "valid number", input: "42", wantInt32: 42, wantValid: true},
		{name: "empty", input: "", wantInt32: 0, wantValid: false},
		{name: "non-numeric", input: "abc", wantInt32: 0, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToNullInt32(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Int32 != tt.wantInt32 {
				t.Errorf("Int32 = %d, want %d", got.Int32, tt.wantInt32)
			}
		})
	}
}

func TestStringToNullBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		wantBool  bool
		wantValid bool
	}{
		{name: "true", input: "true", wantBool: true, wantValid: true},
		{name: "false", input: "false", wantBool: false, wantValid: true},
		{name: "1", input: "1", wantBool: true, wantValid: true},
		{name: "0", input: "0", wantBool: false, wantValid: true},
		{name: "invalid", input: "maybe", wantBool: false, wantValid: false},
		{name: "empty", input: "", wantBool: false, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := StringToNullBool(tt.input)
			if got.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", got.Valid, tt.wantValid)
			}
			if got.Bool != tt.wantBool {
				t.Errorf("Bool = %v, want %v", got.Bool, tt.wantBool)
			}
		})
	}
}

func TestReadNullString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input sql.NullString
		want  string
	}{
		{name: "valid string", input: sql.NullString{String: "hello", Valid: true}, want: "hello"},
		{name: "empty valid string", input: sql.NullString{String: "", Valid: true}, want: ""},
		{name: "null returns literal null", input: sql.NullString{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReadNullString(tt.input)
			if got != tt.want {
				t.Errorf("ReadNullString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadNullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input sql.NullInt64
		want  string
	}{
		{name: "valid", input: sql.NullInt64{Int64: 42, Valid: true}, want: "42"},
		{name: "zero", input: sql.NullInt64{Int64: 0, Valid: true}, want: "0"},
		{name: "null", input: sql.NullInt64{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReadNullInt64(tt.input)
			if got != tt.want {
				t.Errorf("ReadNullInt64() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadNullInt32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input sql.NullInt32
		want  string
	}{
		{name: "valid", input: sql.NullInt32{Int32: 7, Valid: true}, want: "7"},
		{name: "null", input: sql.NullInt32{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReadNullInt32(tt.input)
			if got != tt.want {
				t.Errorf("ReadNullInt32() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadNullTime(t *testing.T) {
	t.Parallel()
	ts := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name  string
		input sql.NullTime
		want  string
	}{
		{name: "valid", input: sql.NullTime{Time: ts, Valid: true}, want: ts.String()},
		{name: "null", input: sql.NullTime{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReadNullTime(tt.input)
			if got != tt.want {
				t.Errorf("ReadNullTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadNullBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input sql.NullBool
		want  string
	}{
		{name: "true", input: sql.NullBool{Bool: true, Valid: true}, want: "true"},
		{name: "false", input: sql.NullBool{Bool: false, Valid: true}, want: "false"},
		{name: "null", input: sql.NullBool{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ReadNullBool(tt.input)
			if got != tt.want {
				t.Errorf("ReadNullBool() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONRawToString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input json.RawMessage
		want  string
	}{
		{name: "valid JSON object", input: json.RawMessage(`{"key":"value"}`), want: `{"key":"value"}`},
		{name: "valid JSON array", input: json.RawMessage(`[1,2,3]`), want: `[1,2,3]`},
		{name: "null JSON", input: json.RawMessage(`null`), want: `null`},
		{name: "empty string JSON", input: json.RawMessage(`""`), want: `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := JSONRawToString(tt.input)
			if got != tt.want {
				t.Errorf("JSONRawToString() = %q, want %q", got, tt.want)
			}
		})
	}
}
