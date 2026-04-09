package handlers

import (
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// TablesListHandler handles GET /admin/tables.
// Lists all registered CMS metadata tables.
func TablesListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := svc.Driver().ListTables()
		if err != nil {
			utility.DefaultLogger.Error("failed to list tables", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		items := make([]db.Tables, 0)
		if list != nil {
			items = *list
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Tables"}`)
			RenderWithOOB(w, r, pages.TablesListContent(items, csrfToken),
				OOBSwap{TargetID: "admin-dialogs", Component: pages.TablesDialogs(csrfToken)})
			return
		}

		layout := NewAdminData(r, "Tables")
		Render(w, r, pages.TablesList(layout, items))
	}
}

// verifyPassword checks the confirm_password form field against the
// authenticated user's stored hash. Returns true if valid.
// On failure, sends an HTMX toast error and returns false.
func verifyPassword(w http.ResponseWriter, r *http.Request, svc *service.Registry) bool {
	password := r.FormValue("confirm_password")
	if password == "" {
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Password confirmation required", "type": "error"}}`)
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}

	fullUser, err := svc.Driver().GetUser(user.UserID)
	if err != nil || fullUser == nil {
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to verify identity", "type": "error"}}`)
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}

	if !auth.CheckPasswordHash(password, fullUser.Hash) {
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Incorrect password", "type": "error"}}`)
		w.WriteHeader(http.StatusForbidden)
		return false
	}

	return true
}

// TableCreateHandler handles POST /admin/tables.
// Requires password re-authentication before creating a table.
func TableCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		if !verifyPassword(w, r, svc) {
			return
		}

		label := strings.TrimSpace(r.FormValue("label"))
		if label == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Label is required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, createErr := svc.Driver().CreateTable(r.Context(), ac, db.CreateTableParams{
			Label: label,
		})
		if createErr != nil {
			utility.DefaultLogger.Error("failed to create table", createErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to create table", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Table created", "type": "success"}}`)
		renderTablesTableRows(w, r, svc)
	}
}

// TableUpdateHandler handles POST /admin/tables/update.
// Requires password re-authentication before updating a table.
func TableUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		if !verifyPassword(w, r, svc) {
			return
		}

		id := strings.TrimSpace(r.FormValue("table_id"))
		label := strings.TrimSpace(r.FormValue("label"))

		if id == "" || label == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Table ID and label are required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		_, updateErr := svc.Driver().UpdateTable(r.Context(), ac, db.UpdateTableParams{
			ID:    id,
			Label: label,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error("failed to update table", updateErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to update table", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Table updated", "type": "success"}}`)
		renderTablesTableRows(w, r, svc)
	}
}

// TableDeleteHandler handles POST /admin/tables/delete.
// Requires password re-authentication before deleting a table.
func TableDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		if !verifyPassword(w, r, svc) {
			return
		}

		id := strings.TrimSpace(r.FormValue("table_id"))
		if id == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Table ID is required", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		ac, acErr := svc.AuditCtx(r.Context())
		if acErr != nil {
			utility.DefaultLogger.Error("failed to build audit context", acErr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		deleteErr := svc.Driver().DeleteTable(r.Context(), ac, id)
		if deleteErr != nil {
			utility.DefaultLogger.Error("failed to delete table", deleteErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "failed to delete table", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Table deleted", "type": "success"}}`)
		renderTablesTableRows(w, r, svc)
	}
}

// renderTablesTableRows reloads and renders the table body after a mutation.
func renderTablesTableRows(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	list, listErr := svc.Driver().ListTables()
	if listErr != nil {
		utility.DefaultLogger.Error("failed to list tables after mutation", listErr)
		http.Error(w, "failed to reload tables", http.StatusInternalServerError)
		return
	}

	items := make([]db.Tables, 0)
	if list != nil {
		items = *list
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, pages.TablesTableRows(items, csrfToken))
}
