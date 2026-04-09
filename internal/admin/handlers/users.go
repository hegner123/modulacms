package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// UsersListHandler handles GET /admin/users.
// HTMX requests receive the page content partial; full requests get the complete layout.
func UsersListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.Users.ListUsersWithRoleLabel(r.Context())
		if err != nil {
			utility.DefaultLogger.Error("failed to list users", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.UserWithRoleLabelRow, 0)
		if items != nil {
			list = *items
		}

		// Load roles for the create form dropdown
		roles, rolesErr := svc.RBAC.ListRoles(r.Context())
		if rolesErr != nil {
			utility.DefaultLogger.Error("failed to list roles", rolesErr)
			roles = &[]db.Roles{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Users"}`)
			Render(w, r, partials.UsersListContent(list, *roles, csrfToken))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.UsersListContent(list, *roles, csrfToken))
			return
		}

		layout := NewAdminData(r, "Users")
		Render(w, r, pages.UsersList(layout, list, *roles, csrfToken))
	}
}

// UserDetailHandler handles GET /admin/users/{id}.
// HTMX requests receive the detail content partial; full requests get the complete layout.
func UserDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing user ID", http.StatusBadRequest)
			return
		}

		user, err := svc.Users.GetUser(r.Context(), types.UserID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get user", err)
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// Load roles for the role dropdown
		roles, rolesErr := svc.RBAC.ListRoles(r.Context())
		if rolesErr != nil {
			utility.DefaultLogger.Error("failed to list roles", rolesErr)
			roles = &[]db.Roles{}
		}

		// Load SSH keys for this user
		var sshKeys []db.UserSshKeys
		keyList, keysErr := svc.SSHKeys.ListKeys(r.Context(), types.NullableUserID{ID: types.UserID(id), Valid: true})
		if keysErr != nil {
			utility.DefaultLogger.Error("failed to list ssh keys", keysErr)
		} else if keyList != nil {
			sshKeys = *keyList
		}

		// Load OAuth connections for this user
		var oauthConns []db.UserOauth
		oauthEntry, oauthErr := svc.Driver().GetUserOauthByUserId(types.NullableUserID{ID: types.UserID(id), Valid: true})
		if oauthErr == nil && oauthEntry != nil {
			oauthConns = []db.UserOauth{*oauthEntry}
		}

		// Check if OAuth is configured
		var oauthConfigured bool
		if cfg, cfgErr := svc.Config(); cfgErr == nil {
			oauthConfigured = cfg.Oauth_Client_Id != ""
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "User: `+user.Name+`"}`)
			Render(w, r, partials.UserDetailContent(*user, *roles, sshKeys, oauthConns, oauthConfigured, csrfToken))
			return
		}

		if IsHTMX(r) {
			Render(w, r, partials.UserDetailContent(*user, *roles, sshKeys, oauthConns, oauthConfigured, csrfToken))
			return
		}

		layout := NewAdminData(r, "User: "+user.Name)
		Render(w, r, pages.UserDetail(layout, *user, *roles, sshKeys, oauthConns, oauthConfigured, csrfToken))
	}
}

// UserCreateHandler handles POST /admin/users.
// Delegates validation, uniqueness, password hashing, and role resolution to UserService.
func UserCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		role := strings.TrimSpace(r.FormValue("role"))

		ac := middleware.AuditContextFromRequest(r, *c)

		_, err := svc.Users.CreateUser(r.Context(), ac, service.CreateUserInput{
			Username: username,
			Name:     name,
			Email:    types.Email(email),
			Password: password,
			Role:     types.RoleID(role),
			IsAdmin:  middleware.ContextIsAdmin(r.Context()),
		})
		if err != nil {
			errs := userServiceErrorToMap(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := svc.RBAC.ListRoles(r.Context())
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserForm(username, name, email, role, *roles, errs, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "user created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users")
		w.WriteHeader(http.StatusOK)
	}
}

// UserUpdateHandler handles POST /admin/users/{id}.
// Delegates validation, uniqueness, password hashing, and role gating to UserService.
func UserUpdateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing user ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		role := strings.TrimSpace(r.FormValue("role"))

		ac := middleware.AuditContextFromRequest(r, *c)

		_, err := svc.Users.UpdateUser(r.Context(), ac, service.UpdateUserInput{
			UserID:   types.UserID(id),
			Username: username,
			Name:     name,
			Email:    types.Email(email),
			Password: password,
			Role:     types.RoleID(role),
			IsAdmin:  middleware.ContextIsAdmin(r.Context()),
		})
		if err != nil {
			if service.IsNotFound(err) {
				http.Error(w, "user not found", http.StatusNotFound)
				return
			}
			errs := userServiceErrorToMap(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := svc.RBAC.ListRoles(r.Context())
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserEditForm(id, username, name, email, role, *roles, errs, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/users/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "user updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// UserDeleteHandler handles DELETE /admin/users/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func UserDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
			return
		}

		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing user ID", http.StatusBadRequest)
			return
		}

		ac := middleware.AuditContextFromRequest(r, *c)
		err := svc.Users.DeleteUser(r.Context(), ac, types.UserID(id))
		if err != nil {
			service.HandleServiceError(w, r, err)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "user deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// userServiceErrorToMap converts a user service error into a map[string]string
// suitable for form re-rendering with field-level error messages.
func userServiceErrorToMap(err error) map[string]string {
	if service.IsValidation(err) {
		return serviceValidationToMap(err)
	}

	var ce *service.ConflictError
	if errors.As(err, &ce) {
		if strings.Contains(ce.Detail, "email") {
			return map[string]string{"email": "A user with this email already exists"}
		}
		if strings.Contains(ce.Detail, "username") {
			return map[string]string{"username": "A user with this username already exists"}
		}
		return map[string]string{"_": ce.Detail}
	}

	if service.IsForbidden(err) {
		return map[string]string{"role": "Only administrators can assign roles"}
	}

	utility.DefaultLogger.Error("user service error", err)
	return map[string]string{"_": "An unexpected error occurred"}
}
