package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
)

// allFieldTypes is the canonical list of supported field types with descriptions.
var allFieldTypes = []pages.FieldTypeInfo{
	{Value: "text", Label: "Text", Description: "Single-line plain text input"},
	{Value: "textarea", Label: "Textarea", Description: "Multi-line plain text input"},
	{Value: "richtext", Label: "Rich Text", Description: "Rich text editor with formatting support"},
	{Value: "number", Label: "Number", Description: "Numeric input (integer or decimal)"},
	{Value: "boolean", Label: "Boolean", Description: "True/false toggle switch"},
	{Value: "date", Label: "Date", Description: "Date picker (date only, no time)"},
	{Value: "datetime", Label: "Datetime", Description: "Date and time picker"},
	{Value: "select", Label: "Select", Description: "Dropdown selection from predefined options"},
	{Value: "media", Label: "Media", Description: "Media file reference (image, video, document)"},
	{Value: "relation", Label: "Relation", Description: "Reference to another content entry"},
	{Value: "json", Label: "JSON", Description: "Freeform JSON data"},
	{Value: "slug", Label: "Slug", Description: "URL-friendly identifier auto-generated from text"},
	{Value: "email", Label: "Email", Description: "Email address with format validation"},
	{Value: "url", Label: "URL", Description: "Web URL with format validation"},
}

// FieldTypesListHandler handles GET /admin/schema/field-types.
// Read-only list of available field types. No CRUD operations.
func FieldTypesListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		layout := NewAdminData(r, "Field Types")
		Render(w, r, pages.FieldTypesList(layout, allFieldTypes))
	}
}
