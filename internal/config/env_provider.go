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

	// Bucket settings
	if forcePathStyle := os.Getenv(ep.prefix + "BUCKET_FORCE_PATH_STYLE"); forcePathStyle != "" {
		config.Bucket_Force_Path_Style = forcePathStyle == "true"
	}

	// Update settings
	if autoUpdate := os.Getenv(ep.prefix + "UPDATE_AUTO_ENABLED"); autoUpdate != "" {
		config.Update_Auto_Enabled = autoUpdate == "true"
	}
	config.Update_Check_Interval = os.Getenv(ep.prefix + "UPDATE_CHECK_INTERVAL")
	config.Update_Channel = os.Getenv(ep.prefix + "UPDATE_CHANNEL")
	if notifyOnly := os.Getenv(ep.prefix + "UPDATE_NOTIFY_ONLY"); notifyOnly != "" {
		config.Update_Notify_Only = notifyOnly == "true"
	}

	return config, nil
}
