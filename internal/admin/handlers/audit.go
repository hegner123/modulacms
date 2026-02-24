package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// AuditLogHandler shows the audit log with pagination.
// Displays change events in reverse chronological order. Read-only.
func AuditLogHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		events, err := driver.ListChangeEvents(db.ListChangeEventsParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to list change events", err)
			http.Error(w, "Failed to load audit log", http.StatusInternalServerError)
			return
		}

		total, countErr := driver.CountChangeEvents()
		if countErr != nil {
			utility.DefaultLogger.Error("failed to count change events", countErr)
			http.Error(w, "Failed to load audit log", http.StatusInternalServerError)
			return
		}

		var changeEvents []db.ChangeEvent
		if events != nil {
			changeEvents = *events
		}

		pd := NewPaginationData(*total, limit, offset, "#audit-table-body", "/admin/audit")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		if IsHTMX(r) {
			Render(w, r, pages.AuditTableRowsPartial(changeEvents, pg))
			return
		}

		layout := NewAdminData(r, "Audit Log")
		Render(w, r, pages.Audit(layout, changeEvents, pg))
	}
}
