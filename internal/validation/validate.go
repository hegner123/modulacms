package validation

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db/types"
)

// FieldInput holds the data needed to validate a single field value.
type FieldInput struct {
	FieldID    types.FieldID
	Label      string
	FieldType  types.FieldType
	Value      string // submitted value
	Validation string // raw JSON from fields.validation column
	Data       string // raw JSON from fields.data column
}

// ValidateField validates a single field input and returns a FieldError if
// validation fails, or nil if the value is valid.
func ValidateField(input FieldInput) *FieldError {
	// 1. Parse validation config.
	config, err := types.ParseValidationConfig(input.Validation)
	if err != nil {
		return &FieldError{
			FieldID:  input.FieldID,
			Label:    input.Label,
			Messages: []string{fmt.Sprintf("invalid validation configuration: %s", err)},
		}
	}

	// 2. Scan rules for a required rule.
	hasRequired := false
	for _, entry := range config.Rules {
		if entry.Rule != nil && entry.Rule.Op == types.RuleRequired {
			hasRequired = true
			break
		}
	}

	// 3. Empty gate: if no required rule and value is empty, pass immediately.
	if !hasRequired && input.Value == "" {
		return nil
	}

	// 4. Required check: if required rule and value is empty, fail immediately.
	if hasRequired && input.Value == "" {
		return &FieldError{
			FieldID:  input.FieldID,
			Label:    input.Label,
			Messages: []string{"is required"},
		}
	}

	// 5. Run type-specific validator.
	var messages []string
	if typeMsg := validateType(input.FieldType, input.Value, input.Data); typeMsg != "" {
		messages = append(messages, typeMsg)
	}

	// 6. Run composable rules (skips required op).
	ruleMessages := EvaluateRules(input.Value, config.Rules)
	messages = append(messages, ruleMessages...)

	if len(messages) == 0 {
		return nil
	}

	return &FieldError{
		FieldID:  input.FieldID,
		Label:    input.Label,
		Messages: messages,
	}
}

// ValidateBatch validates multiple field inputs and returns all errors collected.
func ValidateBatch(inputs []FieldInput) ValidationErrors {
	var ve ValidationErrors
	for _, input := range inputs {
		if fe := ValidateField(input); fe != nil {
			ve.Fields = append(ve.Fields, *fe)
		}
	}
	return ve
}
