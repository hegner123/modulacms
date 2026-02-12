package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// StateStore manages OAuth state parameters for CSRF protection.
// States are stored in-memory with automatic cleanup of expired entries.
type StateStore struct {
	mu             sync.RWMutex
	states         map[string]time.Time
	cleanupCounter int
}

// globalStateStore is the package-level state store instance
var globalStateStore = &StateStore{
	states: make(map[string]time.Time),
}

// cleanupThreshold is the number of GenerateState calls between cleanup sweeps.
const cleanupThreshold = 100

// GenerateState creates a new cryptographically secure state parameter.
// The state is valid for 20 minutes and can only be used once.
// Returns the state string or an error if random generation fails.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(b)

	globalStateStore.mu.Lock()
	globalStateStore.states[state] = time.Now().Add(20 * time.Minute)
	globalStateStore.cleanupCounter++
	if globalStateStore.cleanupCounter >= cleanupThreshold {
		globalStateStore.cleanupCounter = 0
		globalStateStore.cleanupLocked()
	}
	globalStateStore.mu.Unlock()

	return state, nil
}

// ValidateState verifies that a state parameter is valid and not expired.
// States can only be used once - they are deleted after successful validation.
// Returns an error if the state is invalid, expired, or already used.
func ValidateState(state string) error {
	if state == "" {
		return fmt.Errorf("state parameter is required")
	}

	globalStateStore.mu.Lock()
	defer globalStateStore.mu.Unlock()

	expiry, exists := globalStateStore.states[state]
	if !exists {
		return fmt.Errorf("invalid state parameter")
	}

	if time.Now().After(expiry) {
		delete(globalStateStore.states, state)
		return fmt.Errorf("state parameter expired")
	}

	// One-time use: delete after validation
	delete(globalStateStore.states, state)
	return nil
}

// cleanup removes expired states from the store.
// This acquires its own lock and is safe to call externally (e.g. from tests).
func (s *StateStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
}

// cleanupLocked removes expired states. Caller must hold s.mu.
func (s *StateStore) cleanupLocked() {
	now := time.Now()
	for state, expiry := range s.states {
		if now.After(expiry) {
			delete(s.states, state)
		}
	}
}

// Size returns the current number of active states.
// This is primarily useful for testing and monitoring.
func (s *StateStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.states)
}

// Clear removes all states from the store.
// This should only be used in testing.
func (s *StateStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states = make(map[string]time.Time)
	s.cleanupCounter = 0
}

// verifierEntry holds a PKCE verifier with its expiration time.
type verifierEntry struct {
	verifier  string
	expiresAt time.Time
}

// verifierTTL is how long a verifier remains valid.
const verifierTTL = 20 * time.Minute

// VerifierStore manages PKCE verifiers for OAuth flows.
// Verifiers are stored in-memory and associated with state parameters.
type VerifierStore struct {
	mu             sync.RWMutex
	verifiers      map[string]verifierEntry // state -> verifierEntry
	cleanupCounter int
}

// globalVerifierStore is the package-level verifier store instance
var globalVerifierStore = &VerifierStore{
	verifiers: make(map[string]verifierEntry),
}

// StoreVerifier associates a PKCE verifier with a state parameter.
// The verifier can be retrieved later using the state as a key.
func StoreVerifier(state, verifier string) {
	globalVerifierStore.mu.Lock()
	defer globalVerifierStore.mu.Unlock()
	globalVerifierStore.verifiers[state] = verifierEntry{
		verifier:  verifier,
		expiresAt: time.Now().Add(verifierTTL),
	}
	globalVerifierStore.cleanupCounter++
	if globalVerifierStore.cleanupCounter >= cleanupThreshold {
		globalVerifierStore.cleanupCounter = 0
		globalVerifierStore.cleanupLocked()
	}
}

// GetVerifier retrieves the PKCE verifier associated with a state parameter.
// The verifier is deleted after retrieval (one-time use).
// Returns an error if the state is not found or the verifier has expired.
func GetVerifier(state string) (string, error) {
	globalVerifierStore.mu.Lock()
	defer globalVerifierStore.mu.Unlock()

	entry, exists := globalVerifierStore.verifiers[state]
	if !exists {
		return "", fmt.Errorf("verifier not found for state")
	}

	// One-time use: always delete after retrieval
	delete(globalVerifierStore.verifiers, state)

	if time.Now().After(entry.expiresAt) {
		return "", fmt.Errorf("verifier expired for state")
	}

	return entry.verifier, nil
}

// cleanupLocked removes expired verifiers. Caller must hold v.mu.
func (v *VerifierStore) cleanupLocked() {
	now := time.Now()
	for state, entry := range v.verifiers {
		if now.After(entry.expiresAt) {
			delete(v.verifiers, state)
		}
	}
}

// ClearVerifiers removes all verifiers from the store.
// This should only be used in testing.
func (v *VerifierStore) Clear() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.verifiers = make(map[string]verifierEntry)
	v.cleanupCounter = 0
}
