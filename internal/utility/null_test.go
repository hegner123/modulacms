package utility

import (
	"database/sql"
	"testing"
	"time"
)

// ============================================================
// IsNull
// ============================================================

func TestIsNull_NullString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value sql.NullString
		want  bool
	}{
		{name: "valid string", value: sql.NullString{String: "hello", Valid: true}, want: true},
		{name: "valid empty string", value: sql.NullString{String: "", Valid: true}, want: true},
		{name: "invalid", value: sql.NullString{Valid: false}, want: false},
		{name: "zero value", value: sql.NullString{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsNull(tt.value)
			if got != tt.want {
				t.Errorf("IsNull(%+v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsNull_NullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value sql.NullInt64
		want  bool
	}{
		{name: "valid", value: sql.NullInt64{Int64: 42, Valid: true}, want: true},
		{name: "valid zero", value: sql.NullInt64{Int64: 0, Valid: true}, want: true},
		{name: "invalid", value: sql.NullInt64{Valid: false}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsNull(tt.value)
			if got != tt.want {
				t.Errorf("IsNull(%+v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsNull_NullInt32(t *testing.T) {
	t.Parallel()
	got := IsNull(sql.NullInt32{Int32: 10, Valid: true})
	if got != true {
		t.Error("IsNull(valid NullInt32) = false, want true")
	}
	got = IsNull(sql.NullInt32{Valid: false})
	if got != false {
		t.Error("IsNull(invalid NullInt32) = true, want false")
	}
}

func TestIsNull_NullInt16(t *testing.T) {
	t.Parallel()
	got := IsNull(sql.NullInt16{Int16: 5, Valid: true})
	if got != true {
		t.Error("IsNull(valid NullInt16) = false, want true")
	}
	got = IsNull(sql.NullInt16{Valid: false})
	if got != false {
		t.Error("IsNull(invalid NullInt16) = true, want false")
	}
}

func TestIsNull_NullFloat64(t *testing.T) {
	t.Parallel()
	got := IsNull(sql.NullFloat64{Float64: 3.14, Valid: true})
	if got != true {
		t.Error("IsNull(valid NullFloat64) = false, want true")
	}
	got = IsNull(sql.NullFloat64{Valid: false})
	if got != false {
		t.Error("IsNull(invalid NullFloat64) = true, want false")
	}
}

func TestIsNull_NullBool(t *testing.T) {
	t.Parallel()
	got := IsNull(sql.NullBool{Bool: true, Valid: true})
	if got != true {
		t.Error("IsNull(valid NullBool) = false, want true")
	}
	got = IsNull(sql.NullBool{Valid: false})
	if got != false {
		t.Error("IsNull(invalid NullBool) = true, want false")
	}
}

func TestIsNull_NullByte(t *testing.T) {
	t.Parallel()
	got := IsNull(sql.NullByte{Byte: 0xFF, Valid: true})
	if got != true {
		t.Error("IsNull(valid NullByte) = false, want true")
	}
	got = IsNull(sql.NullByte{Valid: false})
	if got != false {
		t.Error("IsNull(invalid NullByte) = true, want false")
	}
}

func TestIsNull_NullTime(t *testing.T) {
	t.Parallel()
	now := time.Now()
	got := IsNull(sql.NullTime{Time: now, Valid: true})
	if got != true {
		t.Error("IsNull(valid NullTime) = false, want true")
	}
	got = IsNull(sql.NullTime{Valid: false})
	if got != false {
		t.Error("IsNull(invalid NullTime) = true, want false")
	}
}

// ============================================================
// NullToString
// ============================================================

func TestNullToString_NullString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value sql.NullString
		want  string
	}{
		{name: "valid string", value: sql.NullString{String: "hello", Valid: true}, want: "hello"},
		{name: "valid empty string", value: sql.NullString{String: "", Valid: true}, want: ""},
		{name: "invalid returns null", value: sql.NullString{Valid: false}, want: "null"},
		{name: "zero value returns null", value: sql.NullString{}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullToString(tt.value)
			if got != tt.want {
				t.Errorf("NullToString(%+v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestNullToString_NullInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value sql.NullInt64
		want  string
	}{
		{name: "positive", value: sql.NullInt64{Int64: 42, Valid: true}, want: "42"},
		{name: "zero", value: sql.NullInt64{Int64: 0, Valid: true}, want: "0"},
		{name: "negative", value: sql.NullInt64{Int64: -100, Valid: true}, want: "-100"},
		{name: "invalid", value: sql.NullInt64{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullToString(tt.value)
			if got != tt.want {
				t.Errorf("NullToString(%+v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestNullToString_NullInt32(t *testing.T) {
	t.Parallel()
	got := NullToString(sql.NullInt32{Int32: 99, Valid: true})
	if got != "99" {
		t.Errorf("NullToString(NullInt32{99, true}) = %q, want %q", got, "99")
	}
	got = NullToString(sql.NullInt32{Valid: false})
	if got != "null" {
		t.Errorf("NullToString(NullInt32{false}) = %q, want %q", got, "null")
	}
}

func TestNullToString_NullInt16(t *testing.T) {
	t.Parallel()
	got := NullToString(sql.NullInt16{Int16: 7, Valid: true})
	if got != "7" {
		t.Errorf("NullToString(NullInt16{7, true}) = %q, want %q", got, "7")
	}
	got = NullToString(sql.NullInt16{Valid: false})
	if got != "null" {
		t.Errorf("NullToString(NullInt16{false}) = %q, want %q", got, "null")
	}
}

func TestNullToString_NullFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value sql.NullFloat64
		want  string
	}{
		{name: "decimal", value: sql.NullFloat64{Float64: 3.14, Valid: true}, want: "3.14"},
		{name: "zero", value: sql.NullFloat64{Float64: 0, Valid: true}, want: "0"},
		{name: "whole number", value: sql.NullFloat64{Float64: 42.0, Valid: true}, want: "42"},
		{name: "invalid", value: sql.NullFloat64{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullToString(tt.value)
			if got != tt.want {
				t.Errorf("NullToString(%+v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestNullToString_NullBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value sql.NullBool
		want  string
	}{
		{name: "true", value: sql.NullBool{Bool: true, Valid: true}, want: "true"},
		{name: "false", value: sql.NullBool{Bool: false, Valid: true}, want: "false"},
		{name: "invalid", value: sql.NullBool{Valid: false}, want: "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NullToString(tt.value)
			if got != tt.want {
				t.Errorf("NullToString(%+v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestNullToString_NullByte(t *testing.T) {
	t.Parallel()
	got := NullToString(sql.NullByte{Byte: 255, Valid: true})
	if got != "255" {
		t.Errorf("NullToString(NullByte{255, true}) = %q, want %q", got, "255")
	}
	got = NullToString(sql.NullByte{Byte: 0, Valid: true})
	if got != "0" {
		t.Errorf("NullToString(NullByte{0, true}) = %q, want %q", got, "0")
	}
	got = NullToString(sql.NullByte{Valid: false})
	if got != "null" {
		t.Errorf("NullToString(NullByte{false}) = %q, want %q", got, "null")
	}
}

func TestNullToString_NullTime(t *testing.T) {
	t.Parallel()
	ref := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
	got := NullToString(sql.NullTime{Time: ref, Valid: true})
	want := "2025-06-15 14:30:45"
	if got != want {
		t.Errorf("NullToString(NullTime{%v, true}) = %q, want %q", ref, got, want)
	}
	got = NullToString(sql.NullTime{Valid: false})
	if got != "null" {
		t.Errorf("NullToString(NullTime{false}) = %q, want %q", got, "null")
	}
}
