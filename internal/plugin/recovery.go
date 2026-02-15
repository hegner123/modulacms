package plugin

import (
	"fmt"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	// CircuitClosed is the normal operating state -- requests are allowed through.
	CircuitClosed CircuitState = iota

	// CircuitOpen means the circuit breaker has tripped -- requests are rejected.
	// After resetInterval elapses since the last failure, the state transitions
	// to CircuitHalfOpen on the next Allow() call.
	CircuitOpen

	// CircuitHalfOpen allows one probe request through. If it succeeds, the
	// circuit closes. If it fails, the circuit reopens.
	CircuitHalfOpen
)

// String returns the human-readable name of the circuit state.
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// CircuitBreaker implements a plugin-level circuit breaker that tracks
// consecutive failures from HTTP handler execution and manager operations
// (reload, init). Hook execution failures are tracked by the separate
// hook-level circuit breaker in hook_engine.go and do NOT feed into this CB.
//
// This prevents a buggy hook from disabling working HTTP routes.
type CircuitBreaker struct {
	mu                sync.Mutex
	state             CircuitState
	consecutiveErrors int
	maxFailures       int           // config Plugin_Max_Failures, default 5
	resetInterval     time.Duration // config Plugin_Reset_Interval, default 60s
	lastFailure       time.Time
	pluginName        string
}

// NewCircuitBreaker creates a new circuit breaker for the named plugin.
func NewCircuitBreaker(pluginName string, maxFailures int, resetInterval time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetInterval <= 0 {
		resetInterval = 60 * time.Second
	}
	return &CircuitBreaker{
		state:         CircuitClosed,
		maxFailures:   maxFailures,
		resetInterval: resetInterval,
		pluginName:    pluginName,
	}
}

// Allow checks whether a request should be allowed through the circuit breaker.
//
//   - CircuitClosed: always allows
//   - CircuitOpen: allows if resetInterval has elapsed since lastFailure (transitions to HalfOpen)
//   - CircuitHalfOpen: allows (probe request)
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) >= cb.resetInterval {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful execution. Resets the consecutive error
// counter and transitions to CircuitClosed if currently in HalfOpen.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveErrors = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed execution. Returns true if the circuit breaker
// just tripped (transitioned from Closed/HalfOpen to Open).
func (cb *CircuitBreaker) RecordFailure() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveErrors++
	cb.lastFailure = time.Now()

	if cb.state == CircuitHalfOpen {
		// Probe failed -- reopen.
		cb.state = CircuitOpen
		utility.DefaultLogger.Warn(
			fmt.Sprintf("plugin %q: circuit breaker re-opened (half-open probe failed, consecutive errors: %d)",
				cb.pluginName, cb.consecutiveErrors),
			nil,
		)
		return true
	}

	if cb.consecutiveErrors >= cb.maxFailures && cb.state == CircuitClosed {
		cb.state = CircuitOpen
		utility.DefaultLogger.Warn(
			fmt.Sprintf("plugin %q: circuit breaker tripped (consecutive errors: %d, threshold: %d)",
				cb.pluginName, cb.consecutiveErrors, cb.maxFailures),
			nil,
		)

		// Metric: record circuit breaker trip (rare event -- negligible overhead).
		RecordCircuitBreakerTrip(cb.pluginName)

		return true
	}

	return false
}

// State returns the current circuit breaker state. Thread-safe.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ConsecutiveErrors returns the current consecutive error count. Thread-safe.
func (cb *CircuitBreaker) ConsecutiveErrors() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.consecutiveErrors
}

// Reset is an admin-initiated force reset that transitions the circuit breaker
// back to CircuitClosed regardless of current state.
//
// Security fix S5: emits slog.Warn audit event with the admin user who performed
// the reset, the plugin name, the prior CB state, and the consecutive failure
// count at the time of reset. The adminUser string is extracted from the request
// context by the caller (EnablePlugin handler) and passed in directly -- Reset
// does not accept context.Context because it is a pure state transition with no I/O.
func (cb *CircuitBreaker) Reset(adminUser string) {
	cb.mu.Lock()
	priorState := cb.state
	priorErrors := cb.consecutiveErrors
	cb.state = CircuitClosed
	cb.consecutiveErrors = 0
	cb.mu.Unlock()

	// S5: Audit event for admin-initiated reset.
	utility.DefaultLogger.Warn(
		fmt.Sprintf("plugin %q: circuit breaker reset by admin", cb.pluginName),
		nil,
		"admin_user", adminUser,
		"prior_state", priorState.String(),
		"prior_consecutive_errors", priorErrors,
	)
}

// Trip forces the circuit breaker into the Open state. Used when external
// conditions (e.g., drain timeout) indicate the plugin has systemic issues.
func (cb *CircuitBreaker) Trip(reason string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitOpen
	cb.lastFailure = time.Now()

	utility.DefaultLogger.Warn(
		fmt.Sprintf("plugin %q: circuit breaker force-tripped: %s", cb.pluginName, reason),
		nil,
	)

	RecordCircuitBreakerTrip(cb.pluginName)
}

// SafeExecuteResult holds the result of a SafeExecute call.
type SafeExecuteResult struct {
	Err      error
	Panicked bool
	PanicVal any
}

// SafeExecute runs fn with panic recovery and circuit breaker integration.
//
// Flow:
//  1. Check cb.Allow() -- if circuit is open, return error immediately.
//  2. Run fn() with defer/recover for panic safety.
//  3. Record success or failure on the circuit breaker.
//  4. On panic: wrap the panic value in an error and record failure.
//
// If cb is nil, fn is executed without circuit breaker checks (useful for
// tests or when the circuit breaker is not yet initialized).
func SafeExecute(cb *CircuitBreaker, fn func() error) SafeExecuteResult {
	if cb != nil && !cb.Allow() {
		return SafeExecuteResult{
			Err: fmt.Errorf("circuit breaker open for plugin %q", cb.pluginName),
		}
	}

	var result SafeExecuteResult

	func() {
		defer func() {
			if r := recover(); r != nil {
				result.Panicked = true
				result.PanicVal = r
				result.Err = fmt.Errorf("plugin panic: %v", r)
			}
		}()

		result.Err = fn()
	}()

	if cb != nil {
		if result.Err != nil {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}

	return result
}
