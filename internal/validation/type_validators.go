package validation

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/hegner123/modulacms/internal/db/types"
)

// selectOption mirrors the format used by fields.data JSON for select options.
type selectOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// validateType runs the type-specific validation for the given field type and value.
// Returns an error message string, or empty string if valid.
// Unknown field types return empty string (skip type validation).
func validateType(ft types.FieldType, value string, data string) string {
	switch ft {
	case types.FieldTypeText, types.FieldTypeTextarea, types.FieldTypeRichText:
		// No type validation for text types.
		return ""

	case types.FieldTypeNumber:
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "must be a valid number"
		}
		return ""

	case types.FieldTypeIDRef:
		id := types.ContentID(value)
		if err := id.Validate(); err != nil {
			return "must be a valid content reference (ULID)"
		}
		return ""

	case types.FieldTypeDate:
		_, err := time.Parse("2006-01-02", value)
		if err != nil {
			return "must be a valid date (YYYY-MM-DD)"
		}
		return ""

	case types.FieldTypeDatetime:
		if _, err := time.Parse(time.RFC3339, value); err == nil {
			return ""
		}
		if _, err := time.Parse("2006-01-02T15:04:05", value); err == nil {
			return ""
		}
		if _, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
			return ""
		}
		return "must be a valid datetime"

	case types.FieldTypeBoolean:
		if value == "true" || value == "false" || value == "1" || value == "0" {
			return ""
		}
		return "must be a boolean (true, false, 1, or 0)"

	case types.FieldTypeSelect:
		return validateSelect(value, data)

	case types.FieldTypeEmail:
		e := types.Email(value)
		if err := e.Validate(); err != nil {
			return "must be a valid email address"
		}
		return ""

	case types.FieldTypeURL:
		u := types.URL(value)
		if err := u.Validate(); err != nil {
			return "must be a valid URL"
		}
		return ""

	case types.FieldTypeSlug:
		s := types.Slug(value)
		if err := s.Validate(); err != nil {
			return "must be a valid slug"
		}
		return ""

	case types.FieldTypeMedia:
		id := types.MediaID(value)
		if err := id.Validate(); err != nil {
			return "must be a valid media reference (ULID)"
		}
		return ""

	case types.FieldTypeJSON:
		if !json.Valid([]byte(value)) {
			return "must be valid JSON"
		}
		return ""

	default:
		// Unknown field type: skip type validation, composable rules still run.
		return ""
	}
}

// validateSelect checks that value is one of the allowed options defined in the
// field's data JSON column.
func validateSelect(value string, data string) string {
	if data == "" || data == "{}" {
		// No options defined; cannot validate membership.
		return "no select options configured"
	}

	var options []selectOption
	if err := json.Unmarshal([]byte(data), &options); err != nil {
		return "invalid select options configuration"
	}

	for _, opt := range options {
		if opt.Value == value {
			return ""
		}
	}
	return "must be one of the allowed options"
}
