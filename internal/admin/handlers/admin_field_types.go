package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
)

// AdminFieldTypesListHandler handles GET /admin/admin-schema/field-types.
// Read-only list of available field types. No CRUD operations.
// Reuses the same allFieldTypes slice defined in field_types.go.
func AdminFieldTypesListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		layout := NewAdminData(r, "Admin Field Types")
		RenderNav(w, r, "Admin Field Types",
			pages.AdminFieldTypesListContent(allFieldTypes),
			pages.AdminFieldTypesList(layout, allFieldTypes))
	}
}
