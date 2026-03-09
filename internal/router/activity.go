package router

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/service"
)

// ActivityRecentHandler handles requests for recent change events with actor info.
func ActivityRecentHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetRecentActivity(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetRecentActivity(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var limit int64 = 25
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.ParseInt(l, 10, 64)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	views, err := svc.AuditLog.GetRecentActivity(r.Context(), limit)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(views)
}
