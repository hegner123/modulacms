package tui

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
)

// FieldProperty represents a single key-value property of a field for display.
type FieldProperty struct {
	Key      string // display label: "Label", "Type", etc.
	Value    string // display value
	Field    string // struct field name for dispatch: "Label", "Type", "SortOrder", etc.
	Editable bool   // false for read-only props (ID, dates)
}

// FieldPropertiesFromField builds the property list for a regular field.
func FieldPropertiesFromField(f db.Fields) []FieldProperty {
	return []FieldProperty{
		{Key: "Label", Value: f.Label, Field: "Label", Editable: true},
		{Key: "Name", Value: f.Name, Field: "Name", Editable: true},
		{Key: "Type", Value: string(f.Type), Field: "Type", Editable: true},
		{Key: "Sort Order", Value: fmt.Sprintf("%d", f.SortOrder), Field: "SortOrder", Editable: true},
		{Key: "Data", Value: compactJSON(f.Data), Field: "Data", Editable: true},
		{Key: "Validation ID", Value: f.ValidationID.String(), Field: "ValidationID", Editable: true},
		{Key: "UI Config", Value: compactJSON(f.UIConfig), Field: "UIConfig", Editable: true},
		{Key: "Translatable", Value: fmt.Sprintf("%t", f.Translatable), Field: "Translatable", Editable: true},
		{Key: "Roles", Value: nullableStringDisplay(f.Roles.String, f.Roles.Valid), Field: "Roles", Editable: true},
		{Key: "ID", Value: string(f.FieldID), Field: "FieldID", Editable: false},
		{Key: "Author", Value: authorDisplay(f.AuthorID.ID, f.AuthorID.Valid), Field: "AuthorID", Editable: false},
		{Key: "Created", Value: f.DateCreated.String(), Field: "DateCreated", Editable: false},
		{Key: "Modified", Value: f.DateModified.String(), Field: "DateModified", Editable: false},
	}
}

// FieldPropertiesFromAdminField builds the property list for an admin field.
func FieldPropertiesFromAdminField(f db.AdminFields) []FieldProperty {
	return []FieldProperty{
		{Key: "Label", Value: f.Label, Field: "Label", Editable: true},
		{Key: "Name", Value: f.Name, Field: "Name", Editable: true},
		{Key: "Type", Value: string(f.Type), Field: "Type", Editable: true},
		{Key: "Sort Order", Value: fmt.Sprintf("%d", f.SortOrder), Field: "SortOrder", Editable: true},
		{Key: "Data", Value: compactJSON(f.Data), Field: "Data", Editable: true},
		{Key: "Validation ID", Value: f.ValidationID.String(), Field: "ValidationID", Editable: true},
		{Key: "UI Config", Value: compactJSON(f.UIConfig), Field: "UIConfig", Editable: true},
		{Key: "Translatable", Value: fmt.Sprintf("%t", f.Translatable), Field: "Translatable", Editable: true},
		{Key: "Roles", Value: nullableStringDisplay(f.Roles.String, f.Roles.Valid), Field: "Roles", Editable: true},
		{Key: "ID", Value: string(f.AdminFieldID), Field: "AdminFieldID", Editable: false},
		{Key: "Author", Value: authorDisplay(f.AuthorID.ID, f.AuthorID.Valid), Field: "AuthorID", Editable: false},
		{Key: "Created", Value: f.DateCreated.String(), Field: "DateCreated", Editable: false},
		{Key: "Modified", Value: f.DateModified.String(), Field: "DateModified", Editable: false},
	}
}

// compactJSON returns a short display string for a JSON field.
func compactJSON(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || trimmed == "{}" || trimmed == "null" {
		return "(none)"
	}
	if len(trimmed) > 40 {
		return trimmed[:37] + "..."
	}
	return trimmed
}

func nullableStringDisplay(s string, valid bool) string {
	if !valid || s == "" {
		return "(none)"
	}
	return s
}

func authorDisplay[T ~string](id T, valid bool) string {
	if !valid || string(id) == "" {
		return "(none)"
	}
	return string(id)
}
