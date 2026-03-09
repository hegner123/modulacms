package router

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// ContentHealHandler handles POST /api/v1/admin/content/heal.
func ContentHealHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	apiHealContent(w, r, svc)
}

func apiHealContent(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	dryRun := r.URL.Query().Get("dry_run") == "true"

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	report, err := svc.Content.Heal(r.Context(), ac, dryRun)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	writeJSON(w, report)
}
