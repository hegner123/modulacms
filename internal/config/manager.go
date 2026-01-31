package config

import (
	"fmt"
	"strings"
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

	// Normalize bucket endpoint: strip scheme so BucketEndpointURL() controls it.
	// Accepts "http://host:port", "https://host:port", or bare "host:port".
	config.Bucket_Endpoint = strings.TrimPrefix(config.Bucket_Endpoint, "https://")
	config.Bucket_Endpoint = strings.TrimPrefix(config.Bucket_Endpoint, "http://")

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
