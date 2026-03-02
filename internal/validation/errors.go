package validation

import (
	"strings"

	"github.com/hegner123/modulacms/internal/db/types"
)

// FieldError holds validation error messages for a single field.
type FieldError struct {
	FieldID  types.FieldID `json:"field_id"`
	Label    string        `json:"label"`
	Messages []string      `json:"messages"`
}

// Error implements the error interface by joining all messages with "; ".
func (fe *FieldError) Error() string {
	return fe.Label + ": " + strings.Join(fe.Messages, "; ")
}

// ValidationErrors collects validation errors across multiple fields.
type ValidationErrors struct {
	Fields []FieldError `json:"fields"`
}

// Error implements the error interface by joining all field errors with "; ".
func (ve *ValidationErrors) Error() string {
	if len(ve.Fields) == 0 {
		return "validation passed"
	}
	parts := make([]string, len(ve.Fields))
	for i := range ve.Fields {
		parts[i] = ve.Fields[i].Error()
	}
	return strings.Join(parts, "; ")
}

// HasErrors returns true if there is at least one field error.
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Fields) > 0
}

// ForField returns the FieldError for the given field ID, or nil if none exists.
func (ve *ValidationErrors) ForField(id types.FieldID) *FieldError {
	for i := range ve.Fields {
		if ve.Fields[i].FieldID == id {
			return &ve.Fields[i]
		}
	}
	return nil
}

// ClearField removes the FieldError for the given field ID.
func (ve *ValidationErrors) ClearField(id types.FieldID) {
	for i := range ve.Fields {
		if ve.Fields[i].FieldID == id {
			ve.Fields = append(ve.Fields[:i], ve.Fields[i+1:]...)
			return
		}
	}
}
