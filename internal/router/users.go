package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
)

// UsersHandler handles CRUD operations that do not require a specific user ID.
func UsersHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListUsers(w, r, svc)
	case http.MethodPost:
		apiCreateUser(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserHandler handles CRUD operations for specific user items.
func UserHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetUser(w, r, svc)
	case http.MethodPut:
		apiUpdateUser(w, r, svc)
	case http.MethodDelete:
		apiDeleteUser(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserFullHandler handles requests for a single user with all related entities.
func UserFullHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiGetUserFull(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UsersFullHandler handles the composed users+role_label endpoint.
func UsersFullHandler(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	switch r.Method {
	case http.MethodGet:
		apiListUsersWithRoleLabel(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func apiGetUser(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	userID, err := types.ParseUserID(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := svc.Users.GetUser(r.Context(), userID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func apiListUsers(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	// Support optional email query parameter to return a single user
	emailParam := r.URL.Query().Get("email")
	if emailParam != "" {
		user, err := svc.Users.GetUserByEmail(r.Context(), types.Email(emailParam))
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
		return
	}

	users, err := svc.Users.ListUsers(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func apiListUsersWithRoleLabel(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	users, err := svc.Users.ListUsersWithRoleLabel(r.Context())
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func apiGetUserFull(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	userID, err := types.ParseUserID(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	view, err := svc.Users.GetUserFull(r.Context(), userID)
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(view)
}

func apiCreateUser(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		Username string       `json:"username"`
		Name     string       `json:"name"`
		Email    types.Email  `json:"email"`
		Password string       `json:"password"`
		Role     types.RoleID `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	created, err := svc.Users.CreateUser(r.Context(), ac, service.CreateUserInput{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		IsAdmin:  middleware.ContextIsAdmin(r.Context()),
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func apiUpdateUser(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	var req struct {
		Username string       `json:"username"`
		Name     string       `json:"name"`
		Email    types.Email  `json:"email"`
		Password string       `json:"password"`
		Role     types.RoleID `json:"role"`
		UserID   types.UserID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, err := svc.Config()
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	updated, err := svc.Users.UpdateUser(r.Context(), ac, service.UpdateUserInput{
		UserID:   req.UserID,
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		IsAdmin:  middleware.ContextIsAdmin(r.Context()),
	})
	if err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func apiDeleteUser(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	q := r.URL.Query().Get("q")
	userID, err := types.ParseUserID(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c, cfgErr := svc.Config()
	if cfgErr != nil {
		service.HandleServiceError(w, r, cfgErr)
		return
	}
	ac := middleware.AuditContextFromRequest(r, *c)

	if err := svc.Users.DeleteUser(r.Context(), ac, userID); err != nil {
		service.HandleServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
