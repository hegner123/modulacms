package config_test

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/hegner123/modulacms/internal/config"
)

// stubProvider is a minimal test double satisfying config.Provider.
type stubProvider struct {
	cfg *config.Config
	err error
}

func (s *stubProvider) Get() (*config.Config, error) {
	if s.err != nil {
		return nil, s.err
	}
	// Return a copy so tests that mutate the result do not affect the stub.
	cp := *s.cfg
	return &cp, nil
}

// Compile-time check that stubProvider satisfies Provider.
var _ config.Provider = (*stubProvider)(nil)

// ---------------------------------------------------------------------------
// NewManager
// ---------------------------------------------------------------------------

func TestNewManager_ReturnsNonNil(t *testing.T) {
	t.Parallel()

	m := config.NewManager(&stubProvider{cfg: &config.Config{}})
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

// ---------------------------------------------------------------------------
// Manager.Load
// ---------------------------------------------------------------------------

func TestManager_Load_Success(t *testing.T) {
	t.Parallel()

	sp := &stubProvider{cfg: &config.Config{
		Environment: "testing",
		Port:        ":9090",
	}}

	m := config.NewManager(sp)
	if err := m.Load(); err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	cfg, err := m.Config()
	if err != nil {
		t.Fatalf("Config() unexpected error: %v", err)
	}
	if cfg.Environment != "testing" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "testing")
	}
	if cfg.Port != ":9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, ":9090")
	}
}

func TestManager_Load_ProviderError(t *testing.T) {
	t.Parallel()

	providerErr := errors.New("disk on fire")
	sp := &stubProvider{err: providerErr}

	m := config.NewManager(sp)
	err := m.Load()

	if err == nil {
		t.Fatal("Load() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "loading configuration") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "loading configuration")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("expected error to wrap provider error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Manager.Load -- bucket endpoint normalization
// ---------------------------------------------------------------------------

func TestManager_Load_NormalizesBucketEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "strips https prefix",
			endpoint: "https://s3.example.com",
			want:     "s3.example.com",
		},
		{
			name:     "strips http prefix",
			endpoint: "http://localhost:9000",
			want:     "localhost:9000",
		},
		{
			name:     "bare host is untouched",
			endpoint: "s3.example.com",
			want:     "s3.example.com",
		},
		{
			name:     "empty endpoint stays empty",
			endpoint: "",
			want:     "",
		},
		{
			name:     "host with port but no scheme is untouched",
			endpoint: "minio:9000",
			want:     "minio:9000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sp := &stubProvider{cfg: &config.Config{
				Bucket_Endpoint: tt.endpoint,
			}}

			m := config.NewManager(sp)
			if err := m.Load(); err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}

			cfg, err := m.Config()
			if err != nil {
				t.Fatalf("Config() unexpected error: %v", err)
			}
			if cfg.Bucket_Endpoint != tt.want {
				t.Errorf("Bucket_Endpoint = %q, want %q", cfg.Bucket_Endpoint, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Manager.Load -- keybinding merge
// ---------------------------------------------------------------------------

func TestManager_Load_MergesKeyBindingsIntoDefaults(t *testing.T) {
	t.Parallel()

	// Provider returns a config that overrides only ActionQuit.
	sp := &stubProvider{cfg: &config.Config{
		KeyBindings: config.KeyMap{
			config.ActionQuit: {"x"},
		},
	}}

	m := config.NewManager(sp)
	if err := m.Load(); err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	cfg, err := m.Config()
	if err != nil {
		t.Fatalf("Config() unexpected error: %v", err)
	}

	// ActionQuit should be overridden to just "x"
	if !cfg.KeyBindings.Matches("x", config.ActionQuit) {
		t.Error("expected ActionQuit to match 'x' after override")
	}
	if cfg.KeyBindings.Matches("q", config.ActionQuit) {
		t.Error("expected ActionQuit to NOT match 'q' after override replaced all bindings")
	}

	// ActionSelect should retain its default bindings (was not overridden)
	if !cfg.KeyBindings.Matches("enter", config.ActionSelect) {
		t.Error("expected ActionSelect to still match 'enter' (default binding)")
	}
}

func TestManager_Load_EmptyKeyBindingsGetsDefaults(t *testing.T) {
	t.Parallel()

	// Provider returns a config with nil KeyBindings -- Load should populate
	// with full defaults.
	sp := &stubProvider{cfg: &config.Config{}}

	m := config.NewManager(sp)
	if err := m.Load(); err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	cfg, err := m.Config()
	if err != nil {
		t.Fatalf("Config() unexpected error: %v", err)
	}

	if cfg.KeyBindings == nil {
		t.Fatal("KeyBindings is nil after Load with empty provider config")
	}

	// Spot-check that defaults are present
	if !cfg.KeyBindings.Matches("q", config.ActionQuit) {
		t.Error("expected default binding 'q' -> ActionQuit")
	}
	if !cfg.KeyBindings.Matches("enter", config.ActionSelect) {
		t.Error("expected default binding 'enter' -> ActionSelect")
	}
}

// ---------------------------------------------------------------------------
// Manager.Config -- lazy loading
// ---------------------------------------------------------------------------

func TestManager_Config_LazyLoadsOnFirstCall(t *testing.T) {
	t.Parallel()

	sp := &stubProvider{cfg: &config.Config{
		Environment: "lazy-test",
	}}

	m := config.NewManager(sp)

	// Config() should trigger Load() internally
	cfg, err := m.Config()
	if err != nil {
		t.Fatalf("Config() unexpected error: %v", err)
	}
	if cfg.Environment != "lazy-test" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "lazy-test")
	}
}

func TestManager_Config_LazyLoadError(t *testing.T) {
	t.Parallel()

	sp := &stubProvider{err: errors.New("cannot reach provider")}

	m := config.NewManager(sp)

	_, err := m.Config()
	if err == nil {
		t.Fatal("Config() expected error from failed lazy load, got nil")
	}
	if !strings.Contains(err.Error(), "loading configuration") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "loading configuration")
	}
}

func TestManager_Config_ReturnsLoadedAfterExplicitLoad(t *testing.T) {
	t.Parallel()

	sp := &stubProvider{cfg: &config.Config{
		Environment: "explicit-load",
	}}

	m := config.NewManager(sp)
	if err := m.Load(); err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	cfg, err := m.Config()
	if err != nil {
		t.Fatalf("Config() unexpected error: %v", err)
	}
	if cfg.Environment != "explicit-load" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "explicit-load")
	}
}

func TestManager_Config_ReturnsSamePointerAfterLoad(t *testing.T) {
	t.Parallel()

	sp := &stubProvider{cfg: &config.Config{
		Environment: "pointer-test",
	}}

	m := config.NewManager(sp)
	if err := m.Load(); err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	cfg1, _ := m.Config()
	cfg2, _ := m.Config()

	// After Load, repeated Config() calls should return the same pointer.
	if cfg1 != cfg2 {
		t.Error("Config() returned different pointers after Load -- expected same cached config")
	}
}

// ---------------------------------------------------------------------------
// Manager.Load -- concurrent safety
// ---------------------------------------------------------------------------

func TestManager_Load_ConcurrentCalls(t *testing.T) {
	t.Parallel()

	sp := &stubProvider{cfg: &config.Config{
		Environment: "concurrent-test",
	}}

	m := config.NewManager(sp)

	// Fire multiple concurrent Load calls to verify the mutex does not deadlock.
	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.Load(); err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent Load() error: %v", err)
	}

	cfg, err := m.Config()
	if err != nil {
		t.Fatalf("Config() unexpected error after concurrent loads: %v", err)
	}
	if cfg.Environment != "concurrent-test" {
		t.Errorf("Environment = %q, want %q", cfg.Environment, "concurrent-test")
	}
}
