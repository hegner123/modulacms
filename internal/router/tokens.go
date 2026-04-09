package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// tokenCreateRequest is the API request body for creating a token.
// The raw token value is generated server-side and returned in the response.
type tokenCreateRequest struct {
	UserID    types.NullableUserID `json:"user_id"`
	TokenType string               `json:"token_type"`
	Label     string               `json:"label"`
	ExpiresAt types.Timestamp      `json:"expires_at"`
}

// tokenCreateResponse wraps a created token with the raw value shown once.
type tokenCreateResponse struct {
	db.Tokens
	RawToken string `json:"raw_token"`
}

// TokensHandler handles CRUD operations that do not require a specific user ID.
func TokensHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPost:
		apiCreateToken(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TokenHandler handles CRUD operations for specific user items.
func TokenHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetToken(w, r, svc)
	case http.MethodPut:
		apiUpdateToken(w, r, svc)
	case http.MethodDelete:
		apiDeleteToken(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetToken handles GET requests for a single token.
func apiGetToken(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	tID := r.URL.Query().Get("q")
	if tID == "" {
		err := fmt.Errorf("missing token ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	token, err := svc.Tokens.GetToken(r.Context(), tID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
	return nil
}

// apiCreateToken handles POST requests to create a new token.
// The raw token value is generated server-side and returned once in the
// response as "raw_token". Only the SHA-256 hash is stored in the database.
func apiCreateToken(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return cfgErr
	}

	var req tokenCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	input := service.CreateTokenInput{
		UserID:    req.UserID,
		TokenType: req.TokenType,
		Label:     req.Label,
		Expiry:    req.ExpiresAt,
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	result, err := svc.Tokens.CreateToken(r.Context(), ac, input)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tokenCreateResponse{
		Tokens:   *result.Token,
		RawToken: result.RawToken,
	})
	return nil
}

// apiUpdateToken handles PUT requests to update an existing token.
func apiUpdateToken(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return cfgErr
	}

	var updateToken db.UpdateTokenParams
	if err := json.NewDecoder(r.Body).Decode(&updateToken); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	updated, err := svc.Tokens.UpdateToken(r.Context(), ac, updateToken)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteToken handles DELETE requests for tokens.
func apiDeleteToken(w http.ResponseWriter, r *http.Request, svc *service.Registry) error {
	cfg, cfgErr := svc.Config()
	if cfgErr != nil {
		http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		return cfgErr
	}

	tId := r.URL.Query().Get("q")
	if tId == "" {
		err := fmt.Errorf("missing token ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, *cfg)
	if err := svc.Tokens.DeleteToken(r.Context(), ac, tId); err != nil {
		service.HandleServiceError(w, r, err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
