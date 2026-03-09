package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// UserOauthsHandler handles CRUD operations that do not require a specific user ID.
func UserOauthsHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPost:
		ApiCreateUserOauth(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserOauthHandler handles CRUD operations for specific user OAuth items.
func UserOauthHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		ApiUpdateUserOauth(w, r, svc)
	case http.MethodDelete:
		ApiDeleteUserOauth(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ApiCreateUserOauth handles POST requests to create a new user OAuth connection.
func ApiCreateUserOauth(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.CreateUserOauthParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	cfg, _ := svc.Config()
	ac := middleware.AuditContextFromRequest(r, *cfg)
	created, err := svc.OAuth.CreateUserOauth(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
	return nil
}

// ApiUpdateUserOauth handles PUT requests to update an existing user OAuth connection.
func ApiUpdateUserOauth(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	var params db.UpdateUserOauthParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	cfg, _ := svc.Config()
	ac := middleware.AuditContextFromRequest(r, *cfg)
	updated, err := svc.OAuth.UpdateUserOauth(r.Context(), ac, params)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// ApiDeleteUserOauth handles DELETE requests for user OAuth connections.
func ApiDeleteUserOauth(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	q := r.URL.Query().Get("q")
	uoID := types.UserOauthID(q)
	if err := uoID.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	cfg, _ := svc.Config()
	ac := middleware.AuditContextFromRequest(r, *cfg)
	if err := svc.OAuth.DeleteUserOauth(r.Context(), ac, uoID); err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
