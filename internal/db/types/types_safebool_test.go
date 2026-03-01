package types

import (
	"encoding/json"
	"testing"
)

func TestNewSafeBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		sb := NewSafeBool(true)
		if !sb.Val {
			t.Error("expected Val to be true")
		}
		if !sb.Bool() {
			t.Error("expected Bool() to return true")
		}
	})

	t.Run("false", func(t *testing.T) {
		sb := NewSafeBool(false)
		if sb.Val {
			t.Error("expected Val to be false")
		}
		if sb.Bool() {
			t.Error("expected Bool() to return false")
		}
	})
}

func TestSafeBoolScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    bool
		wantErr bool
	}{
		{"bool true", true, true, false},
		{"bool false", false, false, false},
		{"int64 1", int64(1), true, false},
		{"int64 0", int64(0), false, false},
		{"int64 42", int64(42), true, false},
		{"int32 1", int32(1), true, false},
		{"int32 0", int32(0), false, false},
		{"int 1", int(1), true, false},
		{"int 0", int(0), false, false},
		{"nil", nil, false, true},
		{"string", "true", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb SafeBool
			err := sb.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil && sb.Val != tt.want {
				t.Errorf("Scan(%v) = %v, want %v", tt.input, sb.Val, tt.want)
			}
		})
	}
}

func TestSafeBoolValue(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		sb := NewSafeBool(true)
		v, err := sb.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != true {
			t.Errorf("Value() = %v, want true", v)
		}
	})

	t.Run("false", func(t *testing.T) {
		sb := NewSafeBool(false)
		v, err := sb.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != false {
			t.Errorf("Value() = %v, want false", v)
		}
	})
}

func TestSafeBoolJSON(t *testing.T) {
	t.Run("marshal true", func(t *testing.T) {
		sb := NewSafeBool(true)
		data, err := json.Marshal(sb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != "true" {
			t.Errorf("Marshal = %s, want true", string(data))
		}
	})

	t.Run("marshal false", func(t *testing.T) {
		sb := NewSafeBool(false)
		data, err := json.Marshal(sb)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(data) != "false" {
			t.Errorf("Marshal = %s, want false", string(data))
		}
	})

	t.Run("roundtrip", func(t *testing.T) {
		original := NewSafeBool(true)
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var decoded SafeBool
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if decoded.Val != original.Val {
			t.Errorf("roundtrip: got %v, want %v", decoded.Val, original.Val)
		}
	})

	t.Run("unmarshal false", func(t *testing.T) {
		var sb SafeBool
		if err := json.Unmarshal([]byte("false"), &sb); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if sb.Val {
			t.Error("expected Val to be false after unmarshal")
		}
	})
}

func TestSafeBoolString(t *testing.T) {
	tests := []struct {
		val  bool
		want string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			sb := NewSafeBool(tt.val)
			if got := sb.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
