package types

import (
	"encoding/json"
	"fmt"
)

// EmptyJSON is the canonical empty JSON object used as a default for JSON columns.
const EmptyJSON = "{}"

// Cardinality describes whether a relation targets one or many items.
type Cardinality string

// Cardinality constants for one-to-one and one-to-many relations.
const (
	CardinalityOne  Cardinality = "one"
	CardinalityMany Cardinality = "many"
)

// Validate checks whether c is a valid Cardinality value.
func (c Cardinality) Validate() error {
	switch c {
	case CardinalityOne, CardinalityMany:
		return nil
	default:
		return fmt.Errorf("Cardinality: invalid value %q (must be %q or %q)", c, CardinalityOne, CardinalityMany)
	}
}

// RelationConfig is the parsed form of a relation field's data_config JSON column.
type RelationConfig struct {
	TargetDatatypeID DatatypeID  `json:"target_datatype_id"`
	Cardinality      Cardinality `json:"cardinality"`
	MaxDepth         *int        `json:"max_depth,omitempty"`
}

// ValidationConfig is the parsed form of a field's validation_config JSON column.
type ValidationConfig struct {
	Required  bool   `json:"required,omitempty"`
	MinLength *int   `json:"min_length,omitempty"`
	MaxLength *int   `json:"max_length,omitempty"`
	Min       *int   `json:"min,omitempty"`
	Max       *int   `json:"max,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	MaxItems  *int   `json:"max_items,omitempty"`
}

// UIConfig is the parsed form of a field's ui_config JSON column.
type UIConfig struct {
	Widget      string `json:"widget,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"help_text,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
}

// ParseValidationConfig parses a JSON string into a ValidationConfig.
// Returns a zero-value config for empty or "{}" input.
func ParseValidationConfig(s string) (ValidationConfig, error) {
	if s == "" || s == EmptyJSON {
		return ValidationConfig{}, nil
	}
	var vc ValidationConfig
	if err := json.Unmarshal([]byte(s), &vc); err != nil {
		return ValidationConfig{}, fmt.Errorf("ParseValidationConfig: %w", err)
	}
	return vc, nil
}

// ParseUIConfig parses a JSON string into a UIConfig.
// Returns a zero-value config for empty or "{}" input.
func ParseUIConfig(s string) (UIConfig, error) {
	if s == "" || s == EmptyJSON {
		return UIConfig{}, nil
	}
	var uc UIConfig
	if err := json.Unmarshal([]byte(s), &uc); err != nil {
		return UIConfig{}, fmt.Errorf("ParseUIConfig: %w", err)
	}
	return uc, nil
}

// ParseRelationConfig parses a JSON string into a RelationConfig.
// Unlike the other parse functions, this returns an error for empty or "{}" input
// because relation fields require a non-empty data config.
func ParseRelationConfig(s string) (RelationConfig, error) {
	if s == "" || s == EmptyJSON {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: relation fields require a non-empty data config")
	}
	var rc RelationConfig
	if err := json.Unmarshal([]byte(s), &rc); err != nil {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: %w", err)
	}
	if rc.TargetDatatypeID.IsZero() {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: target_datatype_id is required")
	}
	if err := rc.Cardinality.Validate(); err != nil {
		return RelationConfig{}, fmt.Errorf("ParseRelationConfig: %w", err)
	}
	return rc, nil
}
