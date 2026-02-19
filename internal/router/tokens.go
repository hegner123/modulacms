package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// TokensHandler handles CRUD operations that do not require a specific user ID.
func TokensHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPost:
		apiCreateToken(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// TokenHandler handles CRUD operations for specific user items.
func TokenHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetToken(w, r, c)
	case http.MethodPut:
		apiUpdateToken(w, r, c)
	case http.MethodDelete:
		apiDeleteToken(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetToken handles GET requests for a single token
func apiGetToken(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	tID := r.URL.Query().Get("q")
	if tID == "" {
		err := fmt.Errorf("missing token ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	token, err := d.GetToken(tID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
	return nil
}

// apiCreateToken handles POST requests to create a new token
func apiCreateToken(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var newToken db.CreateTokenParams
	err := json.NewDecoder(r.Body).Decode(&newToken)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdToken, err := d.CreateToken(r.Context(), ac, newToken)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdToken)
	return nil
}

// apiUpdateToken handles PUT requests to update an existing token
func apiUpdateToken(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var updateToken db.UpdateTokenParams
	err := json.NewDecoder(r.Body).Decode(&updateToken)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	ac := middleware.AuditContextFromRequest(r, c)
	_, err = d.UpdateToken(r.Context(), ac, updateToken)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	updated, err := d.GetToken(updateToken.ID)
	if err != nil {
		utility.DefaultLogger.Error("failed to fetch updated token", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
	return nil
}

// apiDeleteToken handles DELETE requests for tokens
func apiDeleteToken(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	tId := r.URL.Query().Get("q")
	if tId == "" {
		err := fmt.Errorf("missing token ID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err := d.DeleteToken(r.Context(), ac, tId)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}

