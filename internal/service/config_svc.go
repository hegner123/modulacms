package service

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
)

// ConfigService wraps config.Manager with redaction and category filtering.
type ConfigService struct {
	mgr *config.Manager
}

// NewConfigService creates a ConfigService.
func NewConfigService(mgr *config.Manager) *ConfigService {
	return &ConfigService{mgr: mgr}
}

// ConfigValidationError holds multiple validation errors from a config update.
// The handler type-asserts on this to return the errors list to the client.
type ConfigValidationError struct {
	Errors []string
}

func (e *ConfigValidationError) Error() string {
	return "config validation failed: " + strings.Join(e.Errors, "; ")
}

// ConfigUpdateResult contains the outcome of a successful config update.
type ConfigUpdateResult struct {
	Config          config.Config
	RestartRequired []string
	Warnings        []string
}

// GetConfig returns the full configuration as redacted JSON bytes.
func (s *ConfigService) GetConfig() ([]byte, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return config.RedactedJSON(*cfg)
}

// GetConfigByCategory returns config fields filtered by category as a structured map.
// Returns a ValidationError if the category is unknown.
func (s *ConfigService) GetConfigByCategory(category string) (map[string]any, error) {
	fields := config.FieldsByCategory(config.FieldCategory(category))
	if len(fields) == 0 {
		return nil, NewValidationError("category", "unknown category")
	}

	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	redacted := config.RedactedConfig(*cfg)
	fieldValues := make(map[string]string)
	for _, f := range fields {
		fieldValues[f.JSONKey] = config.ConfigFieldString(redacted, f.JSONKey)
	}

	return map[string]any{
		"category": category,
		"fields":   fieldValues,
	}, nil
}

// UpdateConfig applies partial updates and returns the updated (redacted) config.
// On validation failure, returns a *ConfigValidationError containing the error list.
// On empty input, returns a *ValidationError.
func (s *ConfigService) UpdateConfig(updates map[string]any) (*ConfigUpdateResult, error) {
	if len(updates) == 0 {
		return nil, NewValidationError("updates", "no updates provided")
	}

	vr, err := s.mgr.Update(updates)
	if err != nil && !vr.Valid {
		return nil, &ConfigValidationError{Errors: vr.Errors}
	}
	if err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}

	cfg, cfgErr := s.mgr.Config()
	if cfgErr != nil {
		return nil, fmt.Errorf("load updated config: %w", cfgErr)
	}

	redacted := config.RedactedConfig(*cfg)
	return &ConfigUpdateResult{
		Config:          redacted,
		RestartRequired: vr.RestartRequired,
		Warnings:        vr.Warnings,
	}, nil
}

// GetFieldMetadata returns the field registry and category list for the config schema UI.
func (s *ConfigService) GetFieldMetadata() ([]config.FieldMeta, []config.FieldCategory) {
	return config.FieldRegistry, config.AllCategories()
}
