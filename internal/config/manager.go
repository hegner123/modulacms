package config

import (
	"fmt"
	"sync"
)

// Manager handles configuration loading and access
type Manager struct {
	provider  Provider
	config    *Config
	loaded    bool
	loadMutex sync.Mutex
}

// NewManager creates a new configuration manager with the specified provider
func NewManager(provider Provider) *Manager {
	return &Manager{
		provider: provider,
	}
}

// Load loads configuration from the provider
func (m *Manager) Load() error {
	m.loadMutex.Lock()
	defer m.loadMutex.Unlock()

	config, err := m.provider.Get()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	m.config = config
	m.loaded = true
	return nil
}

// Config returns the loaded configuration
// It loads the config if not already loaded
func (m *Manager) Config() (*Config, error) {
	if !m.loaded {
		if err := m.Load(); err != nil {
			return nil, err
		}
	}
	return m.config, nil
}
