package modula

import (
	"encoding/json"
	"testing"
	"time"
)

func TestID_String(t *testing.T) {
	tests := []struct {
		name string
		id   string
		got  string
	}{
		{
			name: "ContentID",
			id:   "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			got:  ContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV").String(),
		},
		{
			name: "UserID",
			id:   "01BX5ZZKBKACTAV9WEVGEMMVRY",
			got:  UserID("01BX5ZZKBKACTAV9WEVGEMMVRY").String(),
		},
		{
			name: "DatatypeID",
			id:   "01CQXX7RN0W8TYBFGY1E1HQNVS",
			got:  DatatypeID("01CQXX7RN0W8TYBFGY1E1HQNVS").String(),
		},
		{
			name: "empty ContentID",
			id:   "",
			got:  ContentID("").String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.id {
				t.Errorf("String() = %q, want %q", tt.got, tt.id)
			}
		})
	}
}

func TestID_IsZero(t *testing.T) {
	tests := []struct {
		name string
		zero bool
		got  bool
	}{
		{
			name: "empty ContentID is zero",
			zero: true,
			got:  ContentID("").IsZero(),
		},
		{
			name: "non-empty ContentID is not zero",
			zero: false,
			got:  ContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV").IsZero(),
		},
		{
			name: "empty UserID is zero",
			zero: true,
			got:  UserID("").IsZero(),
		},
		{
			name: "non-empty UserID is not zero",
			zero: false,
			got:  UserID("01BX5ZZKBKACTAV9WEVGEMMVRY").IsZero(),
		},
		{
			name: "empty FieldID is zero",
			zero: true,
			got:  FieldID("").IsZero(),
		},
		{
			name: "non-empty FieldID is not zero",
			zero: false,
			got:  FieldID("01CQXX7RN0W8TYBFGY1E1HQNVS").IsZero(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.zero {
				t.Errorf("IsZero() = %v, want %v", tt.got, tt.zero)
			}
		})
	}
}

func TestTimestamp_Time(t *testing.T) {
	tests := []struct {
		name    string
		ts      Timestamp
		wantErr bool
		want    time.Time
	}{
		{
			name:    "valid RFC3339 timestamp",
			ts:      Timestamp("2024-01-15T10:30:00Z"),
			wantErr: false,
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:    "valid timestamp with timezone offset",
			ts:      Timestamp("2024-06-01T15:00:00+05:00"),
			wantErr: false,
			want:    time.Date(2024, 6, 1, 15, 0, 0, 0, time.FixedZone("", 5*60*60)),
		},
		{
			name:    "invalid timestamp",
			ts:      Timestamp("not-a-timestamp"),
			wantErr: true,
		},
		{
			name:    "empty timestamp",
			ts:      Timestamp(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ts.Time()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("Time() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimestamp_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "UTC time",
			input: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:  "epoch",
			input: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "future time",
			input: time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimestamp(tt.input)
			got, err := ts.Time()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Compare truncated to seconds since RFC3339 does not include sub-second precision
			if !got.Equal(tt.input.Truncate(time.Second)) {
				t.Errorf("round-trip failed: got %v, want %v", got, tt.input.Truncate(time.Second))
			}
		})
	}
}

func TestTimestampNow(t *testing.T) {
	ts := TimestampNow()
	if ts.IsZero() {
		t.Error("TimestampNow() is zero, want non-zero")
	}
	if ts.String() == "" {
		t.Error("TimestampNow().String() is empty, want non-empty")
	}
	_, err := ts.Time()
	if err != nil {
		t.Fatalf("TimestampNow() produced unparseable timestamp: %v", err)
	}
}

func TestID_JSONRoundTrip(t *testing.T) {
	t.Run("ContentID marshal and unmarshal", func(t *testing.T) {
		original := ContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV")

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		want := `"01ARZ3NDEKTSV4RRFFQ69G5FAV"`
		if string(data) != want {
			t.Errorf("Marshal = %s, want %s", string(data), want)
		}

		var decoded ContentID
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if decoded != original {
			t.Errorf("Unmarshal = %q, want %q", decoded, original)
		}
	})

	t.Run("UserID marshal and unmarshal", func(t *testing.T) {
		original := UserID("01BX5ZZKBKACTAV9WEVGEMMVRY")

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var decoded UserID
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if decoded != original {
			t.Errorf("Unmarshal = %q, want %q", decoded, original)
		}
	})

	t.Run("empty ContentID round trip", func(t *testing.T) {
		original := ContentID("")

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var decoded ContentID
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if decoded != original {
			t.Errorf("Unmarshal = %q, want %q", decoded, original)
		}
	})

	t.Run("ContentID in struct", func(t *testing.T) {
		type wrapper struct {
			ID ContentID `json:"id"`
		}
		original := wrapper{ID: ContentID("01ARZ3NDEKTSV4RRFFQ69G5FAV")}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var decoded wrapper
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if decoded.ID != original.ID {
			t.Errorf("Unmarshal ID = %q, want %q", decoded.ID, original.ID)
		}
	})
}
