package config

// Provider defines an interface for retrieving configuration
type Provider interface {
	// Get returns the application configuration
	Get() (*Config, error)
}
