package handlers

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// AuditLogHandler shows the audit log with pagination.
// Displays change events in reverse chronological order. Read-only.
func AuditLogHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		events, err := svc.AuditLog.ListChangeEvents(r.Context(), db.ListChangeEventsParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to list change events", err)
			http.Error(w, "failed to load audit log", http.StatusInternalServerError)
			return
		}

		total, countErr := svc.AuditLog.CountChangeEvents(r.Context())
		if countErr != nil {
			utility.DefaultLogger.Error("failed to count change events", countErr)
			http.Error(w, "failed to load audit log", http.StatusInternalServerError)
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

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Audit Log"}`)
			Render(w, r, pages.AuditContent(changeEvents, pg))
			return
		}

		if IsHTMX(r) {
			Render(w, r, pages.AuditTableRowsPartial(changeEvents, pg))
			return
		}

		layout := NewAdminData(r, "Audit Log")
		Render(w, r, pages.Audit(layout, changeEvents, pg))
	}
}

// AuditDetailHandler shows a single change event with full JSON values.
func AuditDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("eventID")
		if id == "" {
			http.Error(w, "missing event ID", http.StatusBadRequest)
			return
		}

		event, err := svc.AuditLog.GetChangeEvent(r.Context(), types.EventID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get change event", err)
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}

		// Resolve user name from ID
		userName := ""
		if event.UserID.Valid {
			user, userErr := svc.Users.GetUser(r.Context(), event.UserID.ID)
			if userErr == nil && user != nil {
				userName = user.Name
			}
		}

		// Fetch related events for the same record
		related, relErr := svc.AuditLog.GetChangeEventsByRecord(r.Context(), event.TableName, event.RecordID)
		var relatedEvents []db.ChangeEvent
		if relErr != nil {
			utility.DefaultLogger.Error("failed to get related events", relErr)
		} else if related != nil {
			relatedEvents = *related
		}

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Audit Event"}`)
			Render(w, r, pages.AuditDetailContent(*event, relatedEvents, userName))
			return
		}

		layout := NewAdminData(r, "Audit Event")
		Render(w, r, pages.AuditDetail(layout, *event, relatedEvents, userName))
	}
}
