package plugin

import (
	"errors"
	"testing"
	"time"
)

// -- CircuitState tests --

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("CircuitState(%d).String() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

// -- CircuitBreaker basic lifecycle --

func TestNewCircuitBreaker_Defaults(t *testing.T) {
	cb := NewCircuitBreaker("test", 0, 0)
	if cb.State() != CircuitClosed {
		t.Errorf("new CB should be closed, got %s", cb.State())
	}
	if cb.ConsecutiveErrors() != 0 {
		t.Errorf("new CB should have 0 errors, got %d", cb.ConsecutiveErrors())
	}
	if cb.maxFailures != 5 {
		t.Errorf("default maxFailures should be 5, got %d", cb.maxFailures)
	}
	if cb.resetInterval != 60*time.Second {
		t.Errorf("default resetInterval should be 60s, got %s", cb.resetInterval)
	}
}

func TestNewCircuitBreaker_CustomConfig(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 10*time.Second)
	if cb.maxFailures != 3 {
		t.Errorf("maxFailures = %d, want 3", cb.maxFailures)
	}
	if cb.resetInterval != 10*time.Second {
		t.Errorf("resetInterval = %s, want 10s", cb.resetInterval)
	}
}

// -- Allow() tests --

func TestCircuitBreaker_Allow_ClosedAlwaysAllows(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)
	for range 10 {
		if !cb.Allow() {
			t.Fatal("closed circuit should always allow")
		}
	}
}

func TestCircuitBreaker_Allow_OpenRejects(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 60*time.Second)

	// Trip the CB.
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after 2 failures, got %s", cb.State())
	}

	if cb.Allow() {
		t.Fatal("open circuit should reject requests")
	}
}

func TestCircuitBreaker_Allow_OpenTransitionsToHalfOpenAfterResetInterval(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1*time.Millisecond)

	// Trip immediately.
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State())
	}

	// Wait for the reset interval to elapse.
	time.Sleep(5 * time.Millisecond)

	if !cb.Allow() {
		t.Fatal("open circuit should allow after resetInterval (half-open probe)")
	}
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected half-open after reset interval probe, got %s", cb.State())
	}
}

// -- RecordSuccess() tests --

func TestCircuitBreaker_RecordSuccess_ClosesFromHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1*time.Millisecond)

	// Trip and recover.
	cb.RecordFailure()
	time.Sleep(5 * time.Millisecond)
	cb.Allow() // transitions to half-open

	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after success from half-open, got %s", cb.State())
	}
	if cb.ConsecutiveErrors() != 0 {
		t.Errorf("expected 0 consecutive errors after success, got %d", cb.ConsecutiveErrors())
	}
}

func TestCircuitBreaker_RecordSuccess_ResetsConsecutiveErrors(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)

	// Accumulate some failures (not enough to trip).
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.ConsecutiveErrors() != 2 {
		t.Fatalf("expected 2 errors, got %d", cb.ConsecutiveErrors())
	}

	cb.RecordSuccess()
	if cb.ConsecutiveErrors() != 0 {
		t.Errorf("expected 0 errors after success, got %d", cb.ConsecutiveErrors())
	}
}

// -- RecordFailure() tests --

func TestCircuitBreaker_RecordFailure_TripsAtThreshold(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 60*time.Second)

	tripped1 := cb.RecordFailure() // 1
	tripped2 := cb.RecordFailure() // 2
	tripped3 := cb.RecordFailure() // 3 -- should trip

	if tripped1 || tripped2 {
		t.Error("should not trip before reaching threshold")
	}
	if !tripped3 {
		t.Error("should trip at threshold")
	}
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after threshold failures, got %s", cb.State())
	}
}

func TestCircuitBreaker_RecordFailure_HalfOpenProbeFailReopens(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1*time.Millisecond)

	cb.RecordFailure() // trip
	time.Sleep(5 * time.Millisecond)
	cb.Allow() // half-open

	tripped := cb.RecordFailure() // probe failed -- reopen
	if !tripped {
		t.Error("probe failure should reopen circuit")
	}
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after failed probe, got %s", cb.State())
	}
}

func TestCircuitBreaker_RecordFailure_DoesNotTripWhenAlreadyOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 60*time.Second)

	cb.RecordFailure()
	cb.RecordFailure() // trips
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State())
	}

	// Additional failure while open should not re-trip.
	tripped := cb.RecordFailure()
	if tripped {
		t.Error("should not report tripped when already open")
	}
}

// -- Reset() tests --

func TestCircuitBreaker_Reset_FromOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 60*time.Second)

	cb.RecordFailure() // trip
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State())
	}

	cb.Reset("admin_user")
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after Reset, got %s", cb.State())
	}
	if cb.ConsecutiveErrors() != 0 {
		t.Errorf("expected 0 errors after Reset, got %d", cb.ConsecutiveErrors())
	}
}

func TestCircuitBreaker_Reset_FromHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1*time.Millisecond)

	cb.RecordFailure()
	time.Sleep(5 * time.Millisecond)
	cb.Allow() // half-open

	cb.Reset("admin_user")
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after Reset from half-open, got %s", cb.State())
	}
}

// -- Trip() tests --

func TestCircuitBreaker_Trip_ForcesOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)

	cb.Trip("test reason")
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after Trip, got %s", cb.State())
	}
}

func TestCircuitBreaker_Trip_SetsLastFailure(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)

	before := time.Now()
	cb.Trip("test reason")

	cb.mu.Lock()
	after := cb.lastFailure
	cb.mu.Unlock()

	if after.Before(before) {
		t.Errorf("lastFailure should be after Trip() call, got %v (Trip called at %v)", after, before)
	}
}

// -- SafeExecute tests --

func TestSafeExecute_SuccessfulExecution(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)

	result := SafeExecute(cb, func() error {
		return nil
	})

	if result.Err != nil {
		t.Errorf("expected nil error, got %v", result.Err)
	}
	if result.Panicked {
		t.Error("should not have panicked")
	}
	if cb.ConsecutiveErrors() != 0 {
		t.Errorf("expected 0 errors after success, got %d", cb.ConsecutiveErrors())
	}
}

func TestSafeExecute_ErrorExecution(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)

	errTest := errors.New("test error")
	result := SafeExecute(cb, func() error {
		return errTest
	})

	if !errors.Is(result.Err, errTest) {
		t.Errorf("expected test error, got %v", result.Err)
	}
	if result.Panicked {
		t.Error("should not have panicked")
	}
	if cb.ConsecutiveErrors() != 1 {
		t.Errorf("expected 1 error after failure, got %d", cb.ConsecutiveErrors())
	}
}

func TestSafeExecute_PanicRecovery(t *testing.T) {
	cb := NewCircuitBreaker("test", 5, 60*time.Second)

	result := SafeExecute(cb, func() error {
		panic("test panic")
	})

	if result.Err == nil {
		t.Error("expected error from panic recovery")
	}
	if !result.Panicked {
		t.Error("should report panicked")
	}
	if result.PanicVal != "test panic" {
		t.Errorf("panic value = %v, want %q", result.PanicVal, "test panic")
	}
	if cb.ConsecutiveErrors() != 1 {
		t.Errorf("expected 1 error after panic, got %d", cb.ConsecutiveErrors())
	}
}

func TestSafeExecute_CircuitOpenRejectsImmediately(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 60*time.Second)
	cb.RecordFailure() // trip

	executed := false
	result := SafeExecute(cb, func() error {
		executed = true
		return nil
	})

	if executed {
		t.Error("function should not execute when circuit is open")
	}
	if result.Err == nil {
		t.Error("expected error when circuit is open")
	}
}

func TestSafeExecute_NilCircuitBreaker(t *testing.T) {
	result := SafeExecute(nil, func() error {
		return nil
	})

	if result.Err != nil {
		t.Errorf("expected nil error with nil CB, got %v", result.Err)
	}
}

func TestSafeExecute_NilCircuitBreaker_PanicStillRecovered(t *testing.T) {
	result := SafeExecute(nil, func() error {
		panic("test panic")
	})

	if !result.Panicked {
		t.Error("panic should still be recovered with nil CB")
	}
	if result.Err == nil {
		t.Error("expected error from panic with nil CB")
	}
}

// -- Cycle tests: trip, reset, trip again --

func TestCircuitBreaker_FullCycle(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 60*time.Second)

	// Phase 1: Trip.
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State())
	}

	// Phase 2: Admin reset.
	cb.Reset("admin")
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after reset, got %s", cb.State())
	}

	// Phase 3: Successes keep it closed.
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected still closed, got %s", cb.State())
	}

	// Phase 4: Trip again.
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open again, got %s", cb.State())
	}
}

func TestCircuitBreaker_SuccessResetsErrorCounter_PreventsFalseTripAfterRecovery(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 60*time.Second)

	// Two failures, then a success.
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	// One more failure should NOT trip (counter was reset).
	tripped := cb.RecordFailure()
	if tripped {
		t.Error("should not trip: success should have reset the counter")
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed, got %s", cb.State())
	}
}
