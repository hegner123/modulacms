// White-box tests for state_store.go: OAuth state parameter generation/validation
// and PKCE verifier storage.
//
// White-box access is needed because:
//   - Testing expiration requires injecting entries with past timestamps directly
//     into globalStateStore.states (no public API to create expired entries).
//   - cleanup() is unexported and needs direct verification.
//   - globalStateStore and globalVerifierStore are package-level variables that
//     must be accessed directly for setup and teardown.
package auth

import (
	"strings"
	"sync"
	"testing"
	"time"
)

// resetStores clears both global stores to prevent cross-test contamination.
// Must be called in every test that touches GenerateState, ValidateState,
// StoreVerifier, or GetVerifier.
func resetStores(t *testing.T) {
	t.Helper()
	globalStateStore.Clear()
	globalVerifierStore.Clear()
}

// ---------------------------------------------------------------------------
// GenerateState
// ---------------------------------------------------------------------------

func TestGenerateState_ReturnsNonEmptyState(t *testing.T) {
	t.Cleanup(func() { resetStores(t) })

	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() returned error: %v", err)
	}
	if state == "" {
		t.Fatal("GenerateState() returned empty state")
	}
}

func TestGenerateState_StoresStateInGlobal(t *testing.T) {
	t.Cleanup(func() { resetStores(t) })

	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() returned error: %v", err)
	}

	// The state should now exist in the global store
	globalStateStore.mu.RLock()
	_, exists := globalStateStore.states[state]
	globalStateStore.mu.RUnlock()

	if !exists {
		t.Errorf("generated state %q not found in globalStateStore", state)
	}
}

func TestGenerateState_ExpiryIsAbout20Minutes(t *testing.T) {
	t.Cleanup(func() { resetStores(t) })

	before := time.Now()
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() returned error: %v", err)
	}
	after := time.Now()

	globalStateStore.mu.RLock()
	expiry := globalStateStore.states[state]
	globalStateStore.mu.RUnlock()

	earliestExpected := before.Add(20 * time.Minute)
	latestExpected := after.Add(20 * time.Minute)

	if expiry.Before(earliestExpected) {
		t.Errorf("expiry %v is before expected earliest %v", expiry, earliestExpected)
	}
	if expiry.After(latestExpected) {
		t.Errorf("expiry %v is after expected latest %v", expiry, latestExpected)
	}
}

func TestGenerateState_ProducesUniqueStates(t *testing.T) {
	t.Cleanup(func() { resetStores(t) })

	seen := make(map[string]struct{})
	iterations := 100
	for range iterations {
		state, err := GenerateState()
		if err != nil {
			t.Fatalf("GenerateState() returned error: %v", err)
		}
		if _, exists := seen[state]; exists {
			t.Fatalf("duplicate state generated: %q", state)
		}
		seen[state] = struct{}{}
	}
}

// GenerateState produces 32 random bytes encoded as base64url.
// base64url of 32 bytes = ceil(32/3)*4 = 44 characters (with padding).
func TestGenerateState_OutputLength(t *testing.T) {
	t.Cleanup(func() { resetStores(t) })

	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() returned error: %v", err)
	}

	// 32 bytes -> base64url produces 44 characters (with = padding)
	expectedLen := 44
	if len(state) != expectedLen {
		t.Errorf("state length = %d, want %d", len(state), expectedLen)
	}
}

func TestGenerateState_IncrementsStoreSize(t *testing.T) {
	t.Cleanup(func() { resetStores(t) })

	for i := 1; i <= 5; i++ {
		_, err := GenerateState()
		if err != nil {
			t.Fatalf("GenerateState() call %d returned error: %v", i, err)
		}

		got := globalStateStore.Size()
		if got != i {
			t.Errorf("after %d GenerateState calls, Size() = %d, want %d", i, got, i)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateState
// ---------------------------------------------------------------------------

func TestValidateState(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns the state to validate
		wantErr string                     // empty means no error expected
	}{
		{
			name: "valid state succeeds",
			setup: func(t *testing.T) string {
				t.Helper()
				state, err := GenerateState()
				if err != nil {
					t.Fatalf("setup: GenerateState() failed: %v", err)
				}
				return state
			},
			wantErr: "",
		},
		{
			name: "empty state is rejected",
			setup: func(t *testing.T) string {
				t.Helper()
				return ""
			},
			wantErr: "state parameter is required",
		},
		{
			name: "unknown state is rejected",
			setup: func(t *testing.T) string {
				t.Helper()
				return "this-state-was-never-generated"
			},
			wantErr: "invalid state parameter",
		},
		{
			name: "expired state is rejected",
			setup: func(t *testing.T) string {
				t.Helper()
				// Inject a state with an expiry in the past
				expiredState := "expired-test-state"
				globalStateStore.mu.Lock()
				globalStateStore.states[expiredState] = time.Now().Add(-1 * time.Minute)
				globalStateStore.mu.Unlock()
				return expiredState
			},
			wantErr: "state parameter expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStores(t)
			t.Cleanup(func() { resetStores(t) })

			state := tt.setup(t)

			err := ValidateState(state)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// ValidateState deletes the state after successful validation (one-time use).
func TestValidateState_OneTimeUse(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() failed: %v", err)
	}

	// First validation should succeed
	if err := ValidateState(state); err != nil {
		t.Fatalf("first ValidateState() failed: %v", err)
	}

	// Second validation of the same state should fail
	err = ValidateState(state)
	if err == nil {
		t.Fatal("second ValidateState() should have failed, got nil")
	}
	if !strings.Contains(err.Error(), "invalid state parameter") {
		t.Errorf("second ValidateState() error = %q, want it to contain %q", err.Error(), "invalid state parameter")
	}
}

// ValidateState removes expired states from the store when it encounters them.
func TestValidateState_DeletesExpiredState(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	expiredState := "will-expire-soon"
	globalStateStore.mu.Lock()
	globalStateStore.states[expiredState] = time.Now().Add(-5 * time.Second)
	globalStateStore.mu.Unlock()

	// Should fail with "expired"
	err := ValidateState(expiredState)
	if err == nil {
		t.Fatal("expected error for expired state, got nil")
	}

	// The expired state should have been removed from the store
	globalStateStore.mu.RLock()
	_, exists := globalStateStore.states[expiredState]
	globalStateStore.mu.RUnlock()

	if exists {
		t.Error("expired state was not deleted from store after validation attempt")
	}
}

// ValidateState should only delete the state being validated, not others.
func TestValidateState_DoesNotAffectOtherStates(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() #1 failed: %v", err)
	}
	state2, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() #2 failed: %v", err)
	}

	// Validate state1
	if err := ValidateState(state1); err != nil {
		t.Fatalf("ValidateState(state1) failed: %v", err)
	}

	// state2 should still be valid
	if err := ValidateState(state2); err != nil {
		t.Fatalf("ValidateState(state2) failed after state1 was consumed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// StateStore.cleanup
// ---------------------------------------------------------------------------

func TestStateStore_Cleanup_RemovesExpiredEntries(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	now := time.Now()

	globalStateStore.mu.Lock()
	globalStateStore.states["expired-1"] = now.Add(-10 * time.Minute)
	globalStateStore.states["expired-2"] = now.Add(-1 * time.Second)
	globalStateStore.states["still-valid"] = now.Add(10 * time.Minute)
	globalStateStore.mu.Unlock()

	globalStateStore.cleanup()

	globalStateStore.mu.RLock()
	defer globalStateStore.mu.RUnlock()

	if _, exists := globalStateStore.states["expired-1"]; exists {
		t.Error("expired-1 should have been cleaned up")
	}
	if _, exists := globalStateStore.states["expired-2"]; exists {
		t.Error("expired-2 should have been cleaned up")
	}
	if _, exists := globalStateStore.states["still-valid"]; !exists {
		t.Error("still-valid should NOT have been cleaned up")
	}
}

func TestStateStore_Cleanup_EmptyStoreIsNoOp(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	// Should not panic on empty store
	globalStateStore.cleanup()

	if globalStateStore.Size() != 0 {
		t.Errorf("Size() = %d after cleanup of empty store, want 0", globalStateStore.Size())
	}
}

func TestStateStore_Cleanup_AllExpired(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	past := time.Now().Add(-1 * time.Hour)
	globalStateStore.mu.Lock()
	globalStateStore.states["a"] = past
	globalStateStore.states["b"] = past
	globalStateStore.states["c"] = past
	globalStateStore.mu.Unlock()

	globalStateStore.cleanup()

	if globalStateStore.Size() != 0 {
		t.Errorf("Size() = %d after cleaning all expired, want 0", globalStateStore.Size())
	}
}

func TestStateStore_Cleanup_NoneExpired(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	future := time.Now().Add(1 * time.Hour)
	globalStateStore.mu.Lock()
	globalStateStore.states["x"] = future
	globalStateStore.states["y"] = future
	globalStateStore.mu.Unlock()

	globalStateStore.cleanup()

	if globalStateStore.Size() != 2 {
		t.Errorf("Size() = %d after cleaning none expired, want 2", globalStateStore.Size())
	}
}

// ---------------------------------------------------------------------------
// StateStore.Size
// ---------------------------------------------------------------------------

func TestStateStore_Size_EmptyStore(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	if got := globalStateStore.Size(); got != 0 {
		t.Errorf("Size() = %d for empty store, want 0", got)
	}
}

func TestStateStore_Size_AfterInsertions(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	future := time.Now().Add(1 * time.Hour)
	globalStateStore.mu.Lock()
	globalStateStore.states["one"] = future
	globalStateStore.states["two"] = future
	globalStateStore.states["three"] = future
	globalStateStore.mu.Unlock()

	if got := globalStateStore.Size(); got != 3 {
		t.Errorf("Size() = %d, want 3", got)
	}
}

// ---------------------------------------------------------------------------
// StateStore.Clear
// ---------------------------------------------------------------------------

func TestStateStore_Clear(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	future := time.Now().Add(1 * time.Hour)
	globalStateStore.mu.Lock()
	globalStateStore.states["a"] = future
	globalStateStore.states["b"] = future
	globalStateStore.mu.Unlock()

	globalStateStore.Clear()

	if got := globalStateStore.Size(); got != 0 {
		t.Errorf("Size() = %d after Clear(), want 0", got)
	}
}

func TestStateStore_Clear_OnEmptyStore(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	// Should not panic
	globalStateStore.Clear()

	if got := globalStateStore.Size(); got != 0 {
		t.Errorf("Size() = %d after Clear() on empty store, want 0", got)
	}
}

// Clear should reinitialize the map, not just nil it out.
func TestStateStore_Clear_StoreIsUsableAfterClear(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	globalStateStore.Clear()

	// Should be able to add entries after clear
	globalStateStore.mu.Lock()
	globalStateStore.states["post-clear"] = time.Now().Add(10 * time.Minute)
	globalStateStore.mu.Unlock()

	if got := globalStateStore.Size(); got != 1 {
		t.Errorf("Size() = %d after adding entry post-Clear(), want 1", got)
	}
}

// ---------------------------------------------------------------------------
// StoreVerifier / GetVerifier
// ---------------------------------------------------------------------------

func TestStoreVerifier_And_GetVerifier(t *testing.T) {
	tests := []struct {
		name         string
		state        string
		verifier     string
		lookupState  string
		wantVerifier string
		wantErr      string
	}{
		{
			name:         "store and retrieve verifier",
			state:        "state-abc",
			verifier:     "verifier-xyz-123",
			lookupState:  "state-abc",
			wantVerifier: "verifier-xyz-123",
		},
		{
			name:        "lookup non-existent state",
			state:       "state-exists",
			verifier:    "some-verifier",
			lookupState: "state-does-not-exist",
			wantErr:     "verifier not found for state",
		},
		{
			name:         "empty state key",
			state:        "",
			verifier:     "verifier-for-empty-state",
			lookupState:  "",
			wantVerifier: "verifier-for-empty-state",
		},
		{
			name:         "empty verifier value",
			state:        "state-with-empty-verifier",
			verifier:     "",
			lookupState:  "state-with-empty-verifier",
			wantVerifier: "",
		},
		{
			name:         "verifier with special characters",
			state:        "state-special",
			verifier:     "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			lookupState:  "state-special",
			wantVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStores(t)
			t.Cleanup(func() { resetStores(t) })

			StoreVerifier(tt.state, tt.verifier)

			got, err := GetVerifier(tt.lookupState)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantVerifier {
				t.Errorf("GetVerifier(%q) = %q, want %q", tt.lookupState, got, tt.wantVerifier)
			}
		})
	}
}

// GetVerifier deletes the verifier after retrieval (one-time use).
func TestGetVerifier_OneTimeUse(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	StoreVerifier("my-state", "my-verifier")

	// First retrieval should succeed
	got, err := GetVerifier("my-state")
	if err != nil {
		t.Fatalf("first GetVerifier() failed: %v", err)
	}
	if got != "my-verifier" {
		t.Errorf("first GetVerifier() = %q, want %q", got, "my-verifier")
	}

	// Second retrieval should fail
	_, err = GetVerifier("my-state")
	if err == nil {
		t.Fatal("second GetVerifier() should have failed, got nil")
	}
	if !strings.Contains(err.Error(), "verifier not found for state") {
		t.Errorf("second GetVerifier() error = %q, want it to contain %q", err.Error(), "verifier not found for state")
	}
}

// StoreVerifier with the same state key overwrites the previous verifier.
func TestStoreVerifier_Overwrite(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	StoreVerifier("same-state", "verifier-v1")
	StoreVerifier("same-state", "verifier-v2")

	got, err := GetVerifier("same-state")
	if err != nil {
		t.Fatalf("GetVerifier() failed: %v", err)
	}
	if got != "verifier-v2" {
		t.Errorf("GetVerifier() = %q, want %q (latest overwrite)", got, "verifier-v2")
	}
}

// Multiple independent state-verifier pairs should not interfere with each other.
func TestStoreVerifier_MultipleIndependentPairs(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	StoreVerifier("state-1", "verifier-1")
	StoreVerifier("state-2", "verifier-2")
	StoreVerifier("state-3", "verifier-3")

	// Retrieve in reverse order to verify independence
	got3, err := GetVerifier("state-3")
	if err != nil {
		t.Fatalf("GetVerifier(state-3) failed: %v", err)
	}
	if got3 != "verifier-3" {
		t.Errorf("GetVerifier(state-3) = %q, want %q", got3, "verifier-3")
	}

	got1, err := GetVerifier("state-1")
	if err != nil {
		t.Fatalf("GetVerifier(state-1) failed: %v", err)
	}
	if got1 != "verifier-1" {
		t.Errorf("GetVerifier(state-1) = %q, want %q", got1, "verifier-1")
	}

	got2, err := GetVerifier("state-2")
	if err != nil {
		t.Fatalf("GetVerifier(state-2) failed: %v", err)
	}
	if got2 != "verifier-2" {
		t.Errorf("GetVerifier(state-2) = %q, want %q", got2, "verifier-2")
	}
}

func TestGetVerifier_EmptyStoreReturnsError(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	_, err := GetVerifier("any-state")
	if err == nil {
		t.Fatal("GetVerifier() on empty store should return error, got nil")
	}
	if !strings.Contains(err.Error(), "verifier not found for state") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "verifier not found for state")
	}
}

// GetVerifier returns an error for expired verifiers.
func TestGetVerifier_ExpiredVerifier(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	// Inject an expired verifier directly
	globalVerifierStore.mu.Lock()
	globalVerifierStore.verifiers["expired-state"] = verifierEntry{
		verifier:  "expired-verifier",
		expiresAt: time.Now().Add(-1 * time.Minute),
	}
	globalVerifierStore.mu.Unlock()

	_, err := GetVerifier("expired-state")
	if err == nil {
		t.Fatal("expected error for expired verifier, got nil")
	}
	if !strings.Contains(err.Error(), "verifier expired") {
		t.Errorf("error = %q, want it to contain 'verifier expired'", err.Error())
	}

	// The expired entry should have been deleted
	globalVerifierStore.mu.RLock()
	_, exists := globalVerifierStore.verifiers["expired-state"]
	globalVerifierStore.mu.RUnlock()
	if exists {
		t.Error("expired verifier was not deleted from store")
	}
}

// ---------------------------------------------------------------------------
// VerifierStore.Clear
// ---------------------------------------------------------------------------

func TestVerifierStore_Clear(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	StoreVerifier("state-a", "verifier-a")
	StoreVerifier("state-b", "verifier-b")

	globalVerifierStore.Clear()

	// Both should now be gone
	_, err := GetVerifier("state-a")
	if err == nil {
		t.Error("GetVerifier(state-a) should fail after Clear(), got nil error")
	}
	_, err = GetVerifier("state-b")
	if err == nil {
		t.Error("GetVerifier(state-b) should fail after Clear(), got nil error")
	}
}

func TestVerifierStore_Clear_UsableAfterClear(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	globalVerifierStore.Clear()

	// Store should be usable after clear
	StoreVerifier("post-clear-state", "post-clear-verifier")
	got, err := GetVerifier("post-clear-state")
	if err != nil {
		t.Fatalf("GetVerifier() after Clear() failed: %v", err)
	}
	if got != "post-clear-verifier" {
		t.Errorf("GetVerifier() = %q, want %q", got, "post-clear-verifier")
	}
}

// ---------------------------------------------------------------------------
// Concurrency: StateStore
// ---------------------------------------------------------------------------

// Verifies that concurrent GenerateState + ValidateState calls do not race.
// Run with: go test -race ./internal/auth/ -run TestStateStore_ConcurrentGenerateAndValidate
func TestStateStore_ConcurrentGenerateAndValidate(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	const goroutines = 50
	states := make(chan string, goroutines)
	var wg sync.WaitGroup

	// Phase 1: generate states concurrently
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			state, err := GenerateState()
			if err != nil {
				t.Errorf("GenerateState() error: %v", err)
				return
			}
			states <- state
		}()
	}
	wg.Wait()
	close(states)

	// Collect all states
	var allStates []string
	for s := range states {
		allStates = append(allStates, s)
	}

	if len(allStates) != goroutines {
		t.Fatalf("expected %d states, got %d", goroutines, len(allStates))
	}

	// Phase 2: validate all states concurrently
	var validateWg sync.WaitGroup
	validateWg.Add(len(allStates))
	for _, s := range allStates {
		go func(state string) {
			defer validateWg.Done()
			if err := ValidateState(state); err != nil {
				t.Errorf("ValidateState(%q) failed: %v", state, err)
			}
		}(s)
	}
	validateWg.Wait()

	// All states should have been consumed
	if got := globalStateStore.Size(); got != 0 {
		t.Errorf("store Size() = %d after validating all states, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Concurrency: VerifierStore
// ---------------------------------------------------------------------------

// Verifies that concurrent StoreVerifier + GetVerifier calls do not race.
// Run with: go test -race ./internal/auth/ -run TestVerifierStore_ConcurrentStoreAndGet
func TestVerifierStore_ConcurrentStoreAndGet(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	const goroutines = 50
	type pair struct {
		state    string
		verifier string
	}

	pairs := make([]pair, goroutines)
	for i := range goroutines {
		pairs[i] = pair{
			state:    strings.Repeat("s", i+1), // unique states
			verifier: strings.Repeat("v", i+1),
		}
	}

	// Phase 1: store all verifiers concurrently
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for _, p := range pairs {
		go func(p pair) {
			defer wg.Done()
			StoreVerifier(p.state, p.verifier)
		}(p)
	}
	wg.Wait()

	// Phase 2: retrieve all verifiers concurrently
	wg.Add(goroutines)
	for _, p := range pairs {
		go func(p pair) {
			defer wg.Done()
			got, err := GetVerifier(p.state)
			if err != nil {
				t.Errorf("GetVerifier(%q) error: %v", p.state, err)
				return
			}
			if got != p.verifier {
				t.Errorf("GetVerifier(%q) = %q, want %q", p.state, got, p.verifier)
			}
		}(p)
	}
	wg.Wait()
}

// ---------------------------------------------------------------------------
// Integration: GenerateState + ValidateState round trip
// ---------------------------------------------------------------------------

func TestGenerateState_ValidateState_RoundTrip(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() failed: %v", err)
	}

	if err := ValidateState(state); err != nil {
		t.Fatalf("ValidateState() failed for freshly generated state: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Integration: full OAuth flow simulation
// ---------------------------------------------------------------------------

// Simulates a realistic OAuth flow: generate state, store verifier,
// validate state, retrieve verifier.
func TestOAuthFlow_StateAndVerifier(t *testing.T) {
	resetStores(t)
	t.Cleanup(func() { resetStores(t) })

	// Step 1: Generate state (simulates redirect to OAuth provider)
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() failed: %v", err)
	}

	// Step 2: Store PKCE verifier associated with this state
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	StoreVerifier(state, verifier)

	// Step 3: Callback arrives -- validate state (simulates OAuth callback)
	if err := ValidateState(state); err != nil {
		t.Fatalf("ValidateState() failed: %v", err)
	}

	// Step 4: Retrieve verifier to exchange code for token
	got, err := GetVerifier(state)
	if err != nil {
		t.Fatalf("GetVerifier() failed: %v", err)
	}
	if got != verifier {
		t.Errorf("GetVerifier() = %q, want %q", got, verifier)
	}

	// Step 5: Both should be consumed now
	err = ValidateState(state)
	if err == nil {
		t.Error("state should be consumed after validation")
	}
	_, err = GetVerifier(state)
	if err == nil {
		t.Error("verifier should be consumed after retrieval")
	}
}
