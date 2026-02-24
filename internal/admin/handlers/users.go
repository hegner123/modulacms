package handlers

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/utility"
)

// UsersListHandler handles GET /admin/users.
// Lists all users. HTMX requests receive partial table rows only.
func UsersListHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := driver.ListUsersWithRoleLabel()
		if err != nil {
			utility.DefaultLogger.Error("failed to list users", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		list := make([]db.UserWithRoleLabelRow, 0)
		if items != nil {
			list = *items
		}

		if IsHTMX(r) {
			Render(w, r, partials.UsersTableRows(list))
			return
		}

		// Load roles for the create form dropdown
		roles, rolesErr := driver.ListRoles()
		if rolesErr != nil {
			utility.DefaultLogger.Error("failed to list roles", rolesErr)
			roles = &[]db.Roles{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "Users")
		Render(w, r, pages.UsersList(layout, list, *roles, csrfToken))
	}
}

// UserDetailHandler handles GET /admin/users/{id}.
// Shows user detail with edit form and role assignment.
func UserDetailHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		user, err := driver.GetUser(types.UserID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to get user", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Load roles for the role dropdown
		roles, rolesErr := driver.ListRoles()
		if rolesErr != nil {
			utility.DefaultLogger.Error("failed to list roles", rolesErr)
			roles = &[]db.Roles{}
		}

		csrfToken := CSRFTokenFromContext(r.Context())
		layout := NewAdminData(r, "User: "+user.Name)
		Render(w, r, pages.UserDetail(layout, *user, *roles, csrfToken))
	}
}

// UserCreateHandler handles POST /admin/users.
// Validates username, name, email (required + unique), password (min 8 chars).
// Hashes password with auth.HashPassword before storing.
func UserCreateHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		role := strings.TrimSpace(r.FormValue("role"))

		errs := make(map[string]string)
		if username == "" {
			errs["username"] = "Username is required"
		}
		if name == "" {
			errs["name"] = "Name is required"
		}
		if email == "" {
			errs["email"] = "Email is required"
		}
		if password == "" {
			errs["password"] = "Password is required"
		} else if len(password) < 8 {
			errs["password"] = "Password must be at least 8 characters"
		}
		if role == "" {
			role = "viewer"
		}

		// Check email uniqueness
		if email != "" {
			existing, _ := driver.GetUserByEmail(types.Email(email))
			if existing != nil {
				errs["email"] = "A user with this email already exists"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := driver.ListRoles()
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserForm(username, name, email, role, *roles, errs, csrfToken))
			return
		}

		hash, hashErr := auth.HashPassword(password)
		if hashErr != nil {
			utility.DefaultLogger.Error("failed to hash password", hashErr)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := driver.ListRoles()
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserForm(username, name, email, role, *roles, map[string]string{"_": "Failed to process password"}, csrfToken))
			return
		}

		currentUser := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), currentUser.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		now := types.NewTimestamp(time.Now())
		_, err := driver.CreateUser(r.Context(), ac, db.CreateUserParams{
			Username:     username,
			Name:         name,
			Email:        types.Email(email),
			Hash:         hash,
			Role:         role,
			DateCreated:  now,
			DateModified: now,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to create user", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := driver.ListRoles()
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserForm(username, name, email, role, *roles, map[string]string{"_": "Failed to create user"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "User created", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users")
		w.WriteHeader(http.StatusOK)
	}
}

// UserUpdateHandler handles POST /admin/users/{id}.
// Can update name, email, role. Password is updated only if provided (not required).
func UserUpdateHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		existing, err := driver.GetUser(types.UserID(id))
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		name := strings.TrimSpace(r.FormValue("name"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		role := strings.TrimSpace(r.FormValue("role"))

		errs := make(map[string]string)
		if username == "" {
			errs["username"] = "Username is required"
		}
		if name == "" {
			errs["name"] = "Name is required"
		}
		if email == "" {
			errs["email"] = "Email is required"
		}
		if password != "" && len(password) < 8 {
			errs["password"] = "Password must be at least 8 characters"
		}
		if role == "" {
			role = existing.Role
		}

		// Check email uniqueness (only if changed)
		if email != "" && types.Email(email) != existing.Email {
			other, _ := driver.GetUserByEmail(types.Email(email))
			if other != nil {
				errs["email"] = "A user with this email already exists"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := driver.ListRoles()
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserEditForm(id, username, name, email, role, *roles, errs, csrfToken))
			return
		}

		// Use existing hash unless a new password is provided
		hash := existing.Hash
		if password != "" {
			newHash, hashErr := auth.HashPassword(password)
			if hashErr != nil {
				utility.DefaultLogger.Error("failed to hash password", hashErr)
				w.WriteHeader(http.StatusInternalServerError)
				csrfToken := CSRFTokenFromContext(r.Context())
				roles, rolesErr := driver.ListRoles()
				if rolesErr != nil {
					roles = &[]db.Roles{}
				}
				Render(w, r, partials.UserEditForm(id, username, name, email, role, *roles, map[string]string{"_": "Failed to process password"}, csrfToken))
				return
			}
			hash = newHash
		}

		currentUser := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), currentUser.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		_, err = driver.UpdateUser(r.Context(), ac, db.UpdateUserParams{
			UserID:       types.UserID(id),
			Username:     username,
			Name:         name,
			Email:        types.Email(email),
			Hash:         hash,
			Role:         role,
			DateCreated:  existing.DateCreated,
			DateModified: types.NewTimestamp(time.Now()),
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to update user", err)
			w.WriteHeader(http.StatusInternalServerError)
			csrfToken := CSRFTokenFromContext(r.Context())
			roles, rolesErr := driver.ListRoles()
			if rolesErr != nil {
				roles = &[]db.Roles{}
			}
			Render(w, r, partials.UserEditForm(id, username, name, email, role, *roles, map[string]string{"_": "Failed to update user"}, csrfToken))
			return
		}

		if !IsHTMX(r) {
			http.Redirect(w, r, "/admin/users/"+id, http.StatusSeeOther)
			return
		}
		w.Header().Set("HX-Trigger", `{"showToast": {"message": "User updated", "type": "success"}}`)
		w.Header().Set("HX-Redirect", "/admin/users/"+id)
		w.WriteHeader(http.StatusOK)
	}
}

// UserDeleteHandler handles DELETE /admin/users/{id}.
// HTMX-only endpoint. Non-HTMX requests receive 405.
func UserDeleteHandler(driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		currentUser := middleware.AuthenticatedUser(r.Context())
		ip, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr != nil {
			ip = r.RemoteAddr
		}
		ac := audited.Ctx(types.NodeID("0"), currentUser.UserID, middleware.RequestIDFromContext(r.Context()), ip)

		err := driver.DeleteUser(r.Context(), ac, types.UserID(id))
		if err != nil {
			utility.DefaultLogger.Error("failed to delete user", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete user", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "User deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}
