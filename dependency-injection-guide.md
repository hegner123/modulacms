# Dependency Injection Guide for ModulaCMS Config Package

This guide demonstrates how to refactor the config package to implement dependency injection, following Go best practices.

## Current Issues

The current implementation has several issues:

1. Global variables (`Env`, `config`, `file`, `err`) make testing difficult
2. Direct dependency on the file system
3. Hard-coded configuration file path
4. No option for different configuration sources (env vars, flags)
5. No proper error handling strategy

## Proposed Solution: Dependency Injection

### Step 1: Create a Configuration Provider Interface

```go
// internal/config/provider.go
package config

// Provider defines an interface for retrieving configuration
type Provider interface {
    // Get returns the application configuration
    Get() (*Config, error)
}
```

### Step 2: Implement File-Based Provider

```go
// internal/config/file_provider.go
package config

import (
    "encoding/json"
    "fmt"
    "os"
    "io"
)

// FileProvider loads configuration from a JSON file
type FileProvider struct {
    path string
}

// NewFileProvider creates a new file-based configuration provider
func NewFileProvider(path string) *FileProvider {
    if path == "" {
        path = "config.json"
    }
    return &FileProvider{path: path}
}

// Get implements the Provider interface
func (fp *FileProvider) Get() (*Config, error) {
    file, err := os.Open(fp.path)
    if err != nil {
        return nil, fmt.Errorf("opening config file: %w", err)
    }
    defer file.Close()

    bytes, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("reading config file: %w", err)
    }

    var config Config
    if err := json.Unmarshal(bytes, &config); err != nil {
        return nil, fmt.Errorf("parsing config JSON: %w", err)
    }

    return &config, nil
}
```

### Step 3: Create a Manager Service for Config

```go
// internal/config/manager.go
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
```

### Step 4: Setting up in main.go

```go
// cmd/main.go (partial)
package main

import (
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/utility"
)

func main() {
    // Parse flags
    configPath := flag.String("config", "config.json", "Path to configuration file")
    verbose := flag.Bool("v", false, "Enable verbose mode")
    flag.Parse()

    // Create config provider and manager
    configProvider := config.NewFileProvider(*configPath)
    configManager := config.NewManager(configProvider)
    
    // Load config
    if err := configManager.Load(); err != nil {
        utility.DefaultLogger.Fatal("Failed to load configuration", err)
    }
    
    // Get the config
    cfg, _ := configManager.Config()

    // Use cfg instead of global Env
    // ...
}
```

### Step 5: Create Environment Variable Provider (Optional)

```go
// internal/config/env_provider.go
package config

import (
    "os"
    "strings"
)

// EnvProvider loads configuration from environment variables
type EnvProvider struct {
    prefix string
}

// NewEnvProvider creates a new environment variable configuration provider
func NewEnvProvider(prefix string) *EnvProvider {
    return &EnvProvider{prefix: prefix}
}

// Get implements the Provider interface
func (ep *EnvProvider) Get() (*Config, error) {
    config := &Config{}
    
    // Map environment variables to config fields
    config.Environment = os.Getenv(ep.prefix + "ENVIRONMENT")
    config.Port = os.Getenv(ep.prefix + "PORT")
    config.SSL_Port = os.Getenv(ep.prefix + "SSL_PORT")
    // ... map other fields
    
    // For array values, use comma separation
    if corsOrigins := os.Getenv(ep.prefix + "CORS_ORIGINS"); corsOrigins != "" {
        config.Cors_Origins = strings.Split(corsOrigins, ",")
    }
    
    return config, nil
}
```

## Benefits of This Approach

1. **Testability**: You can now inject mock providers for testing
2. **Flexibility**: Easy to add new configuration sources
3. **Maintainability**: Clear separation of concerns
4. **Thread Safety**: Proper mutex usage for concurrent access
5. **Error Handling**: Improved error propagation

## Example: Writing Tests

```go
// internal/config/manager_test.go
package config_test

import (
    "testing"
    
    "github.com/hegner123/modulacms/internal/config"
)

// MockProvider implements Provider for testing
type MockProvider struct {
    config *config.Config
    err    error
}

func (mp *MockProvider) Get() (*config.Config, error) {
    return mp.config, mp.err
}

func TestManagerLoad(t *testing.T) {
    // Create mock config
    mockConfig := &config.Config{
        Environment: "test",
        Port:        "8080",
    }
    
    // Create mock provider
    provider := &MockProvider{config: mockConfig}
    
    // Create manager with mock provider
    manager := config.NewManager(provider)
    
    // Load config
    err := manager.Load()
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    // Get loaded config
    cfg, err := manager.Config()
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    // Assert values
    if cfg.Environment != "test" {
        t.Errorf("Expected Environment to be 'test', got '%s'", cfg.Environment)
    }
    
    if cfg.Port != "8080" {
        t.Errorf("Expected Port to be '8080', got '%s'", cfg.Port)
    }
}
```

## Implementation Steps

1. Create the new files in the config package
2. Refactor main.go to use the new approach
3. Update any packages that directly import and use config.Env
4. Write tests to verify functionality
5. Gradually migrate all code to use the injected configuration

## Packages That Need Updates

Any package that currently uses `config.Env` directly should be refactored to:

1. Accept a `*config.Config` parameter in its constructor/functions
2. Or accept a `config.Manager` in its constructor

This makes dependencies explicit and improves testability across the entire codebase.
