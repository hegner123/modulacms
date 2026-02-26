package utility

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// ============================================================
// DbDriverType constants
// ============================================================

func TestDbDriverType_Values(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  DbDriverType
		want string
	}{
		{name: "sqlite", got: DbSqlite, want: "sqlite"},
		{name: "mysql", got: DbMysql, want: "mysql"},
		{name: "postgres", got: DbPsql, want: "postgres"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("DbDriverType constant %s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// ============================================================
// TimestampI / TimestampS / TimestampReadable
// ============================================================

func TestTimestampI(t *testing.T) {
	t.Parallel()
	before := time.Now().Unix()
	got := TimestampI()
	after := time.Now().Unix()
	if got < before || got > after {
		t.Errorf("TimestampI() = %d, expected between %d and %d", got, before, after)
	}
}

func TestTimestampS(t *testing.T) {
	t.Parallel()
	before := time.Now().Unix()
	got := TimestampS()
	after := time.Now().Unix()

	parsed, err := strconv.ParseInt(got, 10, 64)
	if err != nil {
		t.Fatalf("TimestampS() returned non-integer string %q: %v", got, err)
	}
	if parsed < before || parsed > after {
		t.Errorf("TimestampS() parsed = %d, expected between %d and %d", parsed, before, after)
	}
}

func TestTimestampReadable(t *testing.T) {
	t.Parallel()
	got := TimestampReadable()
	// Should be valid RFC3339
	parsed, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatalf("TimestampReadable() = %q, not valid RFC3339: %v", got, err)
	}
	// Should be within a second of now
	diff := time.Since(parsed)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("TimestampReadable() parsed time is %v away from now", diff)
	}
}

// ============================================================
// FormatTimestampForDB
// ============================================================

func TestFormatTimestampForDB(t *testing.T) {
	t.Parallel()

	// Fixed reference time
	ref := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name       string
		driver     DbDriverType
		want       string
		wantLayout string // verify round-trip
	}{
		{
			name:       "sqlite uses RFC3339",
			driver:     DbSqlite,
			want:       ref.Format(time.RFC3339),
			wantLayout: time.RFC3339,
		},
		{
			name:       "mysql uses datetime format",
			driver:     DbMysql,
			want:       "2025-06-15 14:30:45",
			wantLayout: "2006-01-02 15:04:05",
		},
		{
			name:       "postgres uses datetime format",
			driver:     DbPsql,
			want:       "2025-06-15 14:30:45",
			wantLayout: "2006-01-02 15:04:05",
		},
		{
			name:       "unknown driver defaults to RFC3339",
			driver:     DbDriverType("unknown"),
			want:       ref.Format(time.RFC3339),
			wantLayout: time.RFC3339,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatTimestampForDB(ref, tt.driver)
			if got != tt.want {
				t.Errorf("FormatTimestampForDB(%v, %q) = %q, want %q", ref, tt.driver, got, tt.want)
			}
			// Round-trip check
			_, err := time.Parse(tt.wantLayout, got)
			if err != nil {
				t.Errorf("output %q does not round-trip with layout %q: %v", got, tt.wantLayout, err)
			}
		})
	}
}

// ============================================================
// FormatTimestampForDriverString
// ============================================================

func TestFormatTimestampForDriverString(t *testing.T) {
	t.Parallel()

	ref := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name   string
		driver string
		want   string
	}{
		{name: "sqlite", driver: "sqlite", want: ref.Format(time.RFC3339)},
		{name: "mysql", driver: "mysql", want: "2025-01-02 03:04:05"},
		{name: "postgres", driver: "postgres", want: "2025-01-02 03:04:05"},
		{name: "unknown defaults to RFC3339", driver: "badger", want: ref.Format(time.RFC3339)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatTimestampForDriverString(ref, tt.driver)
			if got != tt.want {
				t.Errorf("FormatTimestampForDriverString(%v, %q) = %q, want %q", ref, tt.driver, got, tt.want)
			}
		})
	}
}

// ============================================================
// CurrentTimestampForDB / CurrentTimestampForDriverString
// ============================================================

func TestCurrentTimestampForDB(t *testing.T) {
	t.Parallel()
	// Just ensure it returns a non-empty string and parses for each driver
	tests := []struct {
		name   string
		driver DbDriverType
		layout string
	}{
		{name: "sqlite", driver: DbSqlite, layout: time.RFC3339},
		{name: "mysql", driver: DbMysql, layout: "2006-01-02 15:04:05"},
		{name: "postgres", driver: DbPsql, layout: "2006-01-02 15:04:05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CurrentTimestampForDB(tt.driver)
			if got == "" {
				t.Fatal("CurrentTimestampForDB returned empty string")
			}
			if _, err := time.Parse(tt.layout, got); err != nil {
				t.Errorf("CurrentTimestampForDB(%q) = %q, does not parse with %q: %v", tt.driver, got, tt.layout, err)
			}
		})
	}
}

func TestCurrentTimestampForDriverString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		driver string
		layout string
	}{
		{name: "sqlite", driver: "sqlite", layout: time.RFC3339},
		{name: "mysql", driver: "mysql", layout: "2006-01-02 15:04:05"},
		{name: "postgres", driver: "postgres", layout: "2006-01-02 15:04:05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CurrentTimestampForDriverString(tt.driver)
			if got == "" {
				t.Fatal("CurrentTimestampForDriverString returned empty string")
			}
			if _, err := time.Parse(tt.layout, got); err != nil {
				t.Errorf("CurrentTimestampForDriverString(%q) = %q, does not parse: %v", tt.driver, got, err)
			}
		})
	}
}

// ============================================================
// ParseDBTimestamp
// ============================================================

func TestParseDBTimestamp(t *testing.T) {
	t.Parallel()

	ref := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp string
		driver    DbDriverType
		wantTime  time.Time
		wantErr   bool
	}{
		{
			name:      "sqlite RFC3339",
			timestamp: ref.Format(time.RFC3339),
			driver:    DbSqlite,
			wantTime:  ref,
		},
		{
			name:      "sqlite Unix timestamp",
			timestamp: strconv.FormatInt(ref.Unix(), 10),
			driver:    DbSqlite,
			wantTime:  time.Unix(ref.Unix(), 0),
		},
		{
			name:      "mysql datetime",
			timestamp: "2025-06-15 14:30:45",
			driver:    DbMysql,
			wantTime:  ref,
		},
		{
			name:      "postgres datetime",
			timestamp: "2025-06-15 14:30:45",
			driver:    DbPsql,
			wantTime:  ref,
		},
		{
			name:      "unknown driver uses RFC3339",
			timestamp: ref.Format(time.RFC3339),
			driver:    DbDriverType("unknown"),
			wantTime:  ref,
		},
		{
			name:      "sqlite invalid string",
			timestamp: "not-a-timestamp",
			driver:    DbSqlite,
			wantErr:   true,
		},
		{
			name:      "mysql invalid string",
			timestamp: "not-a-timestamp",
			driver:    DbMysql,
			wantErr:   true,
		},
		{
			name:      "postgres invalid string",
			timestamp: "not-a-timestamp",
			driver:    DbPsql,
			wantErr:   true,
		},
		{
			name:      "unknown driver invalid string",
			timestamp: "garbage",
			driver:    DbDriverType("other"),
			wantErr:   true,
		},
		{
			name:      "empty string",
			timestamp: "",
			driver:    DbSqlite,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDBTimestamp(tt.timestamp, tt.driver)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result: %v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil time pointer, got nil")
			}
			if !got.Equal(tt.wantTime) {
				t.Errorf("got %v, want %v", got, tt.wantTime)
			}
		})
	}
}

// ============================================================
// ParseDBTimestampString
// ============================================================

func TestParseDBTimestampString(t *testing.T) {
	t.Parallel()

	ref := time.Date(2025, 3, 10, 8, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp string
		driver    string
		wantTime  time.Time
		wantErr   bool
	}{
		{
			name:      "sqlite RFC3339",
			timestamp: ref.Format(time.RFC3339),
			driver:    "sqlite",
			wantTime:  ref,
		},
		{
			name:      "mysql datetime",
			timestamp: "2025-03-10 08:00:00",
			driver:    "mysql",
			wantTime:  ref,
		},
		{
			name:      "postgres datetime",
			timestamp: "2025-03-10 08:00:00",
			driver:    "postgres",
			wantTime:  ref,
		},
		{
			// Unknown driver string falls back to sqlite behavior (DbSqlite)
			name:      "unknown driver falls back to sqlite",
			timestamp: ref.Format(time.RFC3339),
			driver:    "cockroach",
			wantTime:  ref,
		},
		{
			name:      "invalid timestamp returns error",
			timestamp: "xyz",
			driver:    "mysql",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDBTimestampString(tt.timestamp, tt.driver)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result: %v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil time pointer, got nil")
			}
			if !got.Equal(tt.wantTime) {
				t.Errorf("got %v, want %v", got, tt.wantTime)
			}
		})
	}
}

// ============================================================
// ParseTimeReadable
// ============================================================

func TestParseTimeReadable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantNil bool
	}{
		{name: "valid RFC3339", input: "2025-06-15T14:30:45Z", wantNil: false},
		{name: "valid with timezone offset", input: "2025-06-15T14:30:45+05:00", wantNil: false},
		{name: "empty string", input: "", wantNil: true},
		{name: "invalid format", input: "June 15 2025", wantNil: true},
		{name: "date only", input: "2025-06-15", wantNil: true},
		{name: "unix timestamp string", input: "1718461845", wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseTimeReadable(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseTimeReadable(%q) = %v, want nil", tt.input, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ParseTimeReadable(%q) = nil, want non-nil", tt.input)
			}
		})
	}

	t.Run("correct value", func(t *testing.T) {
		t.Parallel()
		got := ParseTimeReadable("2025-06-15T14:30:45Z")
		want := time.Date(2025, 6, 15, 14, 30, 45, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("ParseTimeReadable value = %v, want %v", got, want)
		}
	})
}

// ============================================================
// TokenExpiredTime
// ============================================================

func TestTokenExpiredTime(t *testing.T) {
	t.Parallel()

	before := time.Now().Add(168 * time.Hour).Unix()
	gotStr, gotInt := TokenExpiredTime()
	after := time.Now().Add(168 * time.Hour).Unix()

	// Integer should be ~168 hours from now
	if gotInt < before || gotInt > after {
		t.Errorf("TokenExpiredTime() int = %d, expected between %d and %d", gotInt, before, after)
	}

	// String should be the same value
	parsed, err := strconv.ParseInt(gotStr, 10, 64)
	if err != nil {
		t.Fatalf("TokenExpiredTime() string %q is not a valid integer: %v", gotStr, err)
	}
	if parsed != gotInt {
		t.Errorf("TokenExpiredTime() string=%d, int=%d -- should be equal", parsed, gotInt)
	}
}

// ============================================================
// TimestampLessThan
// ============================================================

func TestTimestampLessThan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    string
		want bool
	}{
		{
			name: "past timestamp is less than now",
			a:    strconv.FormatInt(time.Now().Add(-1*time.Hour).Unix(), 10),
			want: true,
		},
		{
			name: "future timestamp is not less than now",
			a:    strconv.FormatInt(time.Now().Add(1*time.Hour).Unix(), 10),
			want: false,
		},
		{
			name: "invalid string returns false",
			a:    "not-a-number",
			want: false,
		},
		{
			name: "empty string returns false",
			a:    "",
			want: false,
		},
		{
			name: "zero timestamp (1970) is less than now",
			a:    "0",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := TimestampLessThan(tt.a)
			if got != tt.want {
				t.Errorf("TimestampLessThan(%q) = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}

// ============================================================
// Round-trip: FormatTimestampForDB -> ParseDBTimestamp
// ============================================================

func TestTimestamp_RoundTrip(t *testing.T) {
	t.Parallel()

	ref := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	drivers := []DbDriverType{DbSqlite, DbMysql, DbPsql}

	for _, driver := range drivers {
		t.Run(string(driver), func(t *testing.T) {
			t.Parallel()
			formatted := FormatTimestampForDB(ref, driver)
			parsed, err := ParseDBTimestamp(formatted, driver)
			if err != nil {
				t.Fatalf("round-trip parse failed: %v", err)
			}
			if !parsed.Equal(ref) {
				t.Errorf("round-trip: formatted %q, parsed %v, want %v", formatted, parsed, ref)
			}
		})
	}
}

// ============================================================
// Round-trip: FormatTimestampForDriverString -> ParseDBTimestampString
// ============================================================

func TestTimestampString_RoundTrip(t *testing.T) {
	t.Parallel()

	ref := time.Date(2025, 7, 4, 12, 0, 0, 0, time.UTC)
	drivers := []string{"sqlite", "mysql", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			t.Parallel()
			formatted := FormatTimestampForDriverString(ref, driver)
			if !strings.Contains(formatted, "2025") {
				t.Fatalf("formatted timestamp %q does not contain year", formatted)
			}
			parsed, err := ParseDBTimestampString(formatted, driver)
			if err != nil {
				t.Fatalf("round-trip parse failed: %v", err)
			}
			if !parsed.Equal(ref) {
				t.Errorf("round-trip: formatted %q, parsed %v, want %v", formatted, parsed, ref)
			}
		})
	}
}
