package config_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// Compile-time checks: all known Provider implementations must satisfy the
// interface. FileProvider is checked in file_provider_test.go. This file
// serves as the canonical location for any additional compile-time checks.

// stubProvider is defined in manager_test.go (same test package).
// Re-asserting it here is redundant, so we only verify that the Provider
// interface itself is usable as a function parameter type.

func TestProvider_InterfaceIsUsable(t *testing.T) {
	t.Parallel()

	// Verify a function accepting Provider can be called with a concrete
	// implementation. This is a smoke test that the interface definition
	// compiles and is usable.
	loadConfig := func(p config.Provider) (*config.Config, error) {
		return p.Get()
	}

	sp := &stubProvider{cfg: &config.Config{Environment: "provider-test"}}
	cfg, err := loadConfig(sp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Environment != "provider-test" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "provider-test")
	}
}
