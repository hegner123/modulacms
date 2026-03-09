package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// SessionsHandler handles CRUD operations that do not require a specific user ID.
func SessionsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPost:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// SessionHandler handles CRUD operations for specific user items.
func SessionHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		apiUpdateSession(w, r, svc)
	case http.MethodDelete:
		apiDeleteSession(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiUpdateSession handles PUT requests to update an existing session
func apiUpdateSession(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.UpdateSessionParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	cfg, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	updated, err := svc.Sessions.UpdateSession(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteSession handles DELETE requests for sessions
func apiDeleteSession(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	sID := types.SessionID(q)
	if err := sID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	cfg, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	if err := svc.Sessions.DeleteSession(r.Context(), ac, sID); err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
