package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// UsersHandler handles CRUD operations that do not require a specific user ID.
func UsersHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		ApiListUsers(w, r, c)
	case http.MethodPost:
		ApiCreateUser(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserHandler handles CRUD operations for specific user items.
func UserHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		ApiGetUser(w, r, c)
	case http.MethodPut:
		ApiUpdateUser(w, r, c)
	case http.MethodDelete:
		ApiDeleteUser(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserFullHandler handles requests for a single user with all related entities.
func UserFullHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiGetUserFull(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiGetUserFull handles GET requests for a user with all related data composed.
func apiGetUserFull(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	userID, err := types.ParseUserID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	view, err := db.AssembleUserFullView(d, userID)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
	return nil
}

// UsersFullHandler handles the composed users+role_label endpoint.
func UsersFullHandler(w http.ResponseWriter, r *http.Request, c config.Config) {
	switch r.Method {
	case http.MethodGet:
		apiListUsersWithRoleLabel(w, r, c)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// apiListUsersWithRoleLabel handles GET requests for listing users with role labels joined.
func apiListUsersWithRoleLabel(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	users, err := d.ListUsersWithRoleLabel()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
	return nil
}

// ApiGetUser handles GET requests for a single user
func ApiGetUser(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	uId, err := types.ParseUserID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	user, err := d.GetUser(uId)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
	return nil
}

// ApiListUsers handles GET requests for listing users
func ApiListUsers(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	users, err := d.ListUsers()
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
	return nil
}

// ApiCreateUser handles POST requests to create a new user
func ApiCreateUser(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var req struct {
		Username     string          `json:"username"`
		Name         string          `json:"name"`
		Email        types.Email     `json:"email"`
		Password     string          `json:"password"`
		Role         string          `json:"role"`
		DateCreated  types.Timestamp `json:"date_created"`
		DateModified types.Timestamp `json:"date_modified"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return fmt.Errorf("password is required")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return err
	}

	// Role assignment validation: non-admins cannot set role
	if req.Role != "" && !middleware.ContextIsAdmin(r.Context()) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "forbidden",
			"detail": "only administrators can assign roles",
		})
		return nil
	}
	// Default to viewer role if role is empty
	if req.Role == "" {
		viewerRole, roleErr := d.GetRoleByLabel("viewer")
		if roleErr != nil {
			utility.DefaultLogger.Error("failed to get viewer role for default assignment", roleErr)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return roleErr
		}
		req.Role = string(viewerRole.RoleID)
	}

	newUser := db.CreateUserParams{
		Username:     req.Username,
		Name:         req.Name,
		Email:        req.Email,
		Hash:         hash,
		Role:         req.Role,
		DateCreated:  req.DateCreated,
		DateModified: req.DateModified,
	}

	ac := middleware.AuditContextFromRequest(r, c)
	createdUser, err := d.CreateUser(r.Context(), ac, newUser)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
	return nil
}

// ApiUpdateUser handles PUT requests to update an existing user
func ApiUpdateUser(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	var req struct {
		Username     string          `json:"username"`
		Name         string          `json:"name"`
		Email        types.Email     `json:"email"`
		Password     string          `json:"password"`
		Role         string          `json:"role"`
		DateCreated  types.Timestamp `json:"date_created"`
		DateModified types.Timestamp `json:"date_modified"`
		UserID       types.UserID    `json:"user_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	// Always fetch the existing user for role comparison and password fallback
	existing, getErr := d.GetUser(req.UserID)
	if getErr != nil {
		utility.DefaultLogger.Error("", getErr)
		http.Error(w, "user not found", http.StatusNotFound)
		return getErr
	}

	// Role assignment validation: non-admins cannot change roles
	if req.Role != "" && req.Role != existing.Role && !middleware.ContextIsAdmin(r.Context()) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "forbidden",
			"detail": "only administrators can assign roles",
		})
		return nil
	}

	var hash string
	if req.Password != "" {
		h, hashErr := auth.HashPassword(req.Password)
		if hashErr != nil {
			utility.DefaultLogger.Error("", hashErr)
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return hashErr
		}
		hash = h
	} else {
		hash = existing.Hash
	}

	updateUser := db.UpdateUserParams{
		Username:     req.Username,
		Name:         req.Name,
		Email:        req.Email,
		Hash:         hash,
		Role:         req.Role,
		DateCreated:  req.DateCreated,
		DateModified: req.DateModified,
		UserID:       req.UserID,
	}

	ac := middleware.AuditContextFromRequest(r, c)
	updatedUser, err := d.UpdateUser(r.Context(), ac, updateUser)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
	return nil
}

// ApiDeleteUser handles DELETE requests for users
func ApiDeleteUser(w http.ResponseWriter, r *http.Request, c config.Config) error {
	d := db.ConfigDB(c)

	q := r.URL.Query().Get("q")
	uId, err := types.ParseUserID(q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	ac := middleware.AuditContextFromRequest(r, c)
	err = d.DeleteUser(r.Context(), ac, uId)
	if err != nil {
		utility.DefaultLogger.Error("", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return nil
}
