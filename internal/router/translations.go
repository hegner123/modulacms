package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// translationRequest is the JSON body for creating a translation.
type translationRequest struct {
	Locale string `json:"locale"`
}

// TranslationHandler dispatches translation operations for content data.
func TranslationHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodPost:
		apiCreateTranslation(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminTranslationHandler dispatches translation operations for admin content data.
func AdminTranslationHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodPost:
		apiCreateAdminTranslation(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiCreateTranslation creates locale-specific content field rows for a content data node.
func apiCreateTranslation(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	// Parse content data ID from path.
	rawID := r.PathValue("id")
	contentDataID := types.ContentID(rawID)
	if err := contentDataID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid content data ID", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req translationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.DefaultLogger.Error("failed to decode translation request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	result, err := svc.Locales.CreateTranslation(r.Context(), ac, contentDataID, req.Locale, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// apiCreateAdminTranslation creates locale-specific admin content field rows for an admin content data node.
func apiCreateAdminTranslation(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	// Parse admin content data ID from path.
	rawID := r.PathValue("id")
	adminContentDataID := types.AdminContentID(rawID)
	if err := adminContentDataID.Validate(); err != nil {
		utility.DefaultLogger.Error("invalid admin content data ID", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req translationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utility.DefaultLogger.Error("failed to decode translation request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := middleware.AuthenticatedUser(r.Context())
	if user == nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	result, err := svc.Locales.CreateAdminTranslation(r.Context(), ac, adminContentDataID, req.Locale, user.UserID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}
