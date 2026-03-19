package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

///////////////////////////////
// REQUEST TYPES
///////////////////////////////

// ValidationCreateRequest is the JSON body for POST /api/v1/validations.
type ValidationCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      string `json:"config"`
}

// ValidationUpdateRequest is the JSON body for PUT /api/v1/validations/{id}.
type ValidationUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      string `json:"config"`
}

///////////////////////////////
// PUBLIC VALIDATION HANDLERS
///////////////////////////////

// ValidationListHandler handles GET /api/v1/validations.
func ValidationListHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	list, err := svc.Validations.ListValidations()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

// ValidationGetHandler handles GET /api/v1/validations/{id}.
func ValidationGetHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractValidationID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v, getErr := svc.Validations.GetValidation(id)
	if getErr != nil {
		service.HandleServiceError(w, r, getErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

// ValidationCreateHandler handles POST /api/v1/validations.
func ValidationCreateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ValidationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	created, err := svc.Validations.CreateValidation(r.Context(), ac, db.CreateValidationParams{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		AuthorID:    types.NullableUserID{ID: user.UserID, Valid: true},
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// ValidationUpdateHandler handles PUT /api/v1/validations/{id}.
func ValidationUpdateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractValidationID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ValidationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	updated, updateErr := svc.Validations.UpdateValidation(r.Context(), ac, db.UpdateValidationParams{
		ValidationID: id,
		Name:         req.Name,
		Description:  req.Description,
		Config:       req.Config,
	})
	if updateErr != nil {
		service.HandleServiceError(w, r, updateErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// ValidationDeleteHandler handles DELETE /api/v1/validations/{id}.
func ValidationDeleteHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractValidationID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	if deleteErr := svc.Validations.DeleteValidation(r.Context(), ac, id); deleteErr != nil {
		service.HandleServiceError(w, r, deleteErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// ValidationSearchHandler handles GET /api/v1/validations/search?name=...
func ValidationSearchHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name query parameter is required", http.StatusBadRequest)
		return
	}

	results, err := svc.Validations.ListValidationsByName(name)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

///////////////////////////////
// ADMIN VALIDATION HANDLERS
///////////////////////////////

// AdminValidationListHandler handles GET /api/v1/admin/validations.
func AdminValidationListHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	list, err := svc.Validations.ListAdminValidations()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

// AdminValidationGetHandler handles GET /api/v1/admin/validations/{id}.
func AdminValidationGetHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractAdminValidationID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v, getErr := svc.Validations.GetAdminValidation(id)
	if getErr != nil {
		service.HandleServiceError(w, r, getErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

// AdminValidationCreateHandler handles POST /api/v1/admin/validations.
func AdminValidationCreateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ValidationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	created, err := svc.Validations.CreateAdminValidation(r.Context(), ac, db.CreateAdminValidationParams{
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		AuthorID:    types.NullableUserID{ID: user.UserID, Valid: true},
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// AdminValidationUpdateHandler handles PUT /api/v1/admin/validations/{id}.
func AdminValidationUpdateHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractAdminValidationID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ValidationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	updated, updateErr := svc.Validations.UpdateAdminValidation(r.Context(), ac, db.UpdateAdminValidationParams{
		AdminValidationID: id,
		Name:              req.Name,
		Description:       req.Description,
		Config:            req.Config,
	})
	if updateErr != nil {
		service.HandleServiceError(w, r, updateErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

// AdminValidationDeleteHandler handles DELETE /api/v1/admin/validations/{id}.
func AdminValidationDeleteHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	id, err := extractAdminValidationID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *cfg)

	if deleteErr := svc.Validations.DeleteAdminValidation(r.Context(), ac, id); deleteErr != nil {
		service.HandleServiceError(w, r, deleteErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// AdminValidationSearchHandler handles GET /api/v1/admin/validations/search?name=...
func AdminValidationSearchHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name query parameter is required", http.StatusBadRequest)
		return
	}

	results, err := svc.Validations.ListAdminValidationsByName(name)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

///////////////////////////////
// HELPERS
///////////////////////////////

// extractValidationID extracts the validation ID from the URL path.
func extractValidationID(r *http.Request) (types.ValidationID, error) {
	path := r.URL.Path
	idx := strings.Index(path, "validations/")
	if idx == -1 {
		return "", fmt.Errorf("validation ID not found in path")
	}
	rest := path[idx+len("validations/"):]
	if slashIdx := strings.IndexByte(rest, '/'); slashIdx != -1 {
		rest = rest[:slashIdx]
	}
	id := types.ValidationID(rest)
	if err := id.Validate(); err != nil {
		return "", fmt.Errorf("invalid validation_id: %w", err)
	}
	return id, nil
}

// extractAdminValidationID extracts the admin validation ID from the URL path.
func extractAdminValidationID(r *http.Request) (types.AdminValidationID, error) {
	path := r.URL.Path
	idx := strings.Index(path, "validations/")
	if idx == -1 {
		return "", fmt.Errorf("admin validation ID not found in path")
	}
	rest := path[idx+len("validations/"):]
	if slashIdx := strings.IndexByte(rest, '/'); slashIdx != -1 {
		rest = rest[:slashIdx]
	}
	id := types.AdminValidationID(rest)
	if err := id.Validate(); err != nil {
		return "", fmt.Errorf("invalid admin_validation_id: %w", err)
	}
	return id, nil
}
