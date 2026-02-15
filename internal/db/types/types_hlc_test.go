package types

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestHLCNow_ReturnsNonZero(t *testing.T) {
	t.Parallel()
	h := HLCNow()
	if h == 0 {
		t.Fatal("HLCNow() returned 0")
	}
}

func TestHLCNow_Monotonic(t *testing.T) {
	t.Parallel()
	prev := HLCNow()
	for range 1000 {
		curr := HLCNow()
		if curr <= prev {
			t.Fatalf("HLCNow() not monotonic: prev=%d, curr=%d", prev, curr)
		}
		prev = curr
	}
}

func TestHLCNow_ThreadSafety(t *testing.T) {
	t.Parallel()
	const goroutines = 20
	const perGoroutine = 200
	results := make(chan HLC, goroutines*perGoroutine)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range perGoroutine {
				results <- HLCNow()
			}
		}()
	}
	wg.Wait()
	close(results)

	// All values should be unique (monotonic across concurrent calls)
	seen := make(map[HLC]struct{}, goroutines*perGoroutine)
	for h := range results {
		if _, exists := seen[h]; exists {
			t.Fatalf("concurrent HLCNow() produced duplicate: %d", h)
		}
		seen[h] = struct{}{}
	}
}

func TestHLC_Physical(t *testing.T) {
	t.Parallel()
	before := time.Now().Add(-time.Second)
	h := HLCNow()
	after := time.Now().Add(time.Second)

	physical := h.Physical()
	if physical.Before(before) || physical.After(after) {
		t.Errorf("Physical() = %v, want between %v and %v", physical, before, after)
	}
}

func TestHLC_Logical(t *testing.T) {
	t.Parallel()
	// Generate two HLCs rapidly (likely same millisecond -> logical counter increments)
	h1 := HLCNow()
	h2 := HLCNow()

	// h2 should be greater
	if h2 <= h1 {
		t.Errorf("second HLC not greater: h1=%d, h2=%d", h1, h2)
	}

	// Logical is the lower 16 bits
	l := h1.Logical()
	// Just verify it's in valid range
	if l > 0xFFFF {
		t.Errorf("Logical() = %d, out of uint16 range", l)
	}
}

func TestHLC_String(t *testing.T) {
	t.Parallel()
	// Known value: physical=100, logical=5 -> (100 << 16) | 5
	h := HLC(100<<16 | 5)
	want := "HLC(100:5)"
	if h.String() != want {
		t.Errorf("String() = %q, want %q", h.String(), want)
	}
}

func TestHLC_BeforeAfter(t *testing.T) {
	t.Parallel()
	a := HLC(100)
	b := HLC(200)

	if !a.Before(b) {
		t.Error("100.Before(200) = false")
	}
	if a.After(b) {
		t.Error("100.After(200) = true")
	}
	if !b.After(a) {
		t.Error("200.After(100) = false")
	}
	if b.Before(a) {
		t.Error("200.Before(100) = true")
	}
	// Equal
	if a.Before(a) {
		t.Error("a.Before(a) = true")
	}
	if a.After(a) {
		t.Error("a.After(a) = true")
	}
}

func TestHLC_Value(t *testing.T) {
	t.Parallel()
	h := HLC(12345)
	v, err := h.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	i, ok := v.(int64)
	if !ok {
		t.Fatalf("Value() type = %T, want int64", v)
	}
	if i != 12345 {
		t.Errorf("Value() = %d, want 12345", i)
	}
}

func TestHLC_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		want    HLC
		wantErr bool
	}{
		{name: "nil", input: nil, want: 0, wantErr: false},
		{name: "int64", input: int64(12345), want: 12345, wantErr: false},
		{name: "int", input: int(99), want: 99, wantErr: false},
		{name: "string", input: "bad", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var h HLC
			err := h.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && h != tt.want {
				t.Errorf("Scan(%v) = %d, want %d", tt.input, h, tt.want)
			}
		})
	}
}

func TestHLC_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	h := HLCNow()
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}

	var got HLC
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if got != h {
		t.Errorf("JSON round-trip: got %d, want %d", got, h)
	}
}

func TestHLC_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()
	var h HLC
	if err := json.Unmarshal([]byte(`"not a number"`), &h); err == nil {
		t.Error("UnmarshalJSON(string) expected error")
	}
}

func TestHLCUpdate_AdvancesPastReceived(t *testing.T) {
	t.Parallel()
	// Create a received HLC far in the "future" (high physical value)
	futureMs := time.Now().Add(10 * time.Second).UnixMilli()
	received := HLC(futureMs << 16)

	updated := HLCUpdate(received)

	// Updated should be >= received (merges with local time, picks max physical)
	if updated < received {
		t.Errorf("HLCUpdate(future) = %d, want >= %d", updated, received)
	}

	// A second call with the same received should produce a strictly larger value
	// (since hlcLast is now at least received, counter increments)
	updated2 := HLCUpdate(received)
	if updated2 <= updated {
		t.Errorf("second HLCUpdate(future) = %d, want > %d", updated2, updated)
	}
}

func TestHLCUpdate_AdvancesPastLocal(t *testing.T) {
	t.Parallel()
	// Get current local HLC
	local := HLCNow()

	// Send a received that is in the past
	pastMs := time.Now().Add(-10 * time.Second).UnixMilli()
	received := HLC(pastMs << 16)

	updated := HLCUpdate(received)

	// Updated should be > local (since local was already advanced)
	if updated <= local {
		t.Errorf("HLCUpdate(past) = %d, want > local %d", updated, local)
	}
}
