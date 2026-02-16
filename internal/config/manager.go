package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Manager handles configuration loading, access, and live updates.
type Manager struct {
	provider  Provider
	saver     Saver
	config    *Config
	loaded    bool
	mu        sync.RWMutex
	onChange  []func(Config)
}

// NewManager creates a new configuration manager with the specified provider.
// If the provider also implements Saver, it will be used for Save/Update operations.
func NewManager(provider Provider) *Manager {
	mgr := &Manager{
		provider: provider,
	}
	if s, ok := provider.(Saver); ok {
		mgr.saver = s
	}
	return mgr
}

// Load loads configuration from the provider.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.loadLocked()
}

// loadLocked performs the actual config loading. Caller must hold the write lock.
func (m *Manager) loadLocked() error {
	config, err := m.provider.Get()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Normalize bucket endpoint: strip scheme so BucketEndpointURL() controls it.
	// Accepts "http://host:port", "https://host:port", or bare "host:port".
	config.Bucket_Endpoint = strings.TrimPrefix(config.Bucket_Endpoint, "https://")
	config.Bucket_Endpoint = strings.TrimPrefix(config.Bucket_Endpoint, "http://")

	// Merge keybinding overrides into defaults so unspecified actions keep
	// their default bindings.
	defaults := DefaultKeyMap()
	defaults.Merge(config.KeyBindings)
	config.KeyBindings = defaults

	m.config = config
	m.loaded = true
	return nil
}

// Config returns a pointer to the loaded configuration.
// It loads the config if not already loaded.
// Callers that need a point-in-time snapshot should dereference: cfg := *ptr.
func (m *Manager) Config() (*Config, error) {
	m.mu.RLock()
	if m.loaded {
		cfg := m.config
		m.mu.RUnlock()
		return cfg, nil
	}
	m.mu.RUnlock()

	// Not loaded yet — take write lock and load.
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock.
	if !m.loaded {
		if err := m.loadLocked(); err != nil {
			return nil, err
		}
	}
	return m.config, nil
}

// Snapshot returns a value copy of the current config for safe concurrent use.
func (m *Manager) Snapshot() (Config, error) {
	cfg, err := m.Config()
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}

// Update applies a partial update to the configuration using a JSON key-value map.
// It validates the result, saves to disk if a Saver is configured, swaps the
// in-memory config, and notifies OnChange listeners.
//
// Values that are the redaction placeholder ("********") are skipped to prevent
// accidentally overwriting sensitive fields when a client round-trips redacted config.
func (m *Manager) Update(updates map[string]any) (ValidationResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		if err := m.loadLocked(); err != nil {
			return ValidationResult{}, fmt.Errorf("loading config before update: %w", err)
		}
	}

	current := *m.config

	// Marshal current config to map, merge updates, unmarshal back.
	currentBytes, err := json.Marshal(current)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("marshaling current config: %w", err)
	}

	var currentMap map[string]any
	if err := json.Unmarshal(currentBytes, &currentMap); err != nil {
		return ValidationResult{}, fmt.Errorf("unmarshaling current config to map: %w", err)
	}

	// Merge updates, skipping redacted values.
	for k, v := range updates {
		if strVal, ok := v.(string); ok && IsRedactedValue(strVal) {
			continue
		}
		currentMap[k] = v
	}

	mergedBytes, err := json.Marshal(currentMap)
	if err != nil {
		return ValidationResult{}, fmt.Errorf("marshaling merged config: %w", err)
	}

	var proposed Config
	if err := json.Unmarshal(mergedBytes, &proposed); err != nil {
		return ValidationResult{}, fmt.Errorf("unmarshaling merged config: %w", err)
	}

	result := ValidateUpdate(current, proposed)
	if !result.Valid {
		return result, fmt.Errorf("validation failed: %s", strings.Join(result.Errors, "; "))
	}

	// Normalize bucket endpoint on the proposed config.
	proposed.Bucket_Endpoint = strings.TrimPrefix(proposed.Bucket_Endpoint, "https://")
	proposed.Bucket_Endpoint = strings.TrimPrefix(proposed.Bucket_Endpoint, "http://")

	// Merge keybinding overrides.
	defaults := DefaultKeyMap()
	defaults.Merge(proposed.KeyBindings)
	proposed.KeyBindings = defaults

	// Save to disk if possible.
	if m.saver != nil {
		if err := m.saver.Save(&proposed); err != nil {
			return result, fmt.Errorf("saving config: %w", err)
		}
	}

	// Swap in-memory config.
	m.config = &proposed

	// Notify listeners outside the lock.
	listeners := make([]func(Config), len(m.onChange))
	copy(listeners, m.onChange)
	go func() {
		snap := proposed
		for _, fn := range listeners {
			fn(snap)
		}
	}()

	return result, nil
}

// Save persists the current in-memory config to disk.
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.saver == nil {
		return fmt.Errorf("no saver configured")
	}
	if !m.loaded || m.config == nil {
		return fmt.Errorf("config not loaded")
	}
	return m.saver.Save(m.config)
}

// OnChange registers a callback that fires after each successful Update().
// Callbacks receive a value copy of the new config.
func (m *Manager) OnChange(fn func(Config)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = append(m.onChange, fn)
}
