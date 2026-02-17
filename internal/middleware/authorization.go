package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const (
	permissionsKey contextKey = "permissions"
	isAdminKey     contextKey = "isAdmin"
)

// PermissionSet is a set of permission strings for O(1) lookup.
type PermissionSet map[string]struct{}

// Has returns true if the set contains the given permission.
func (ps PermissionSet) Has(perm string) bool {
	_, ok := ps[perm]
	return ok
}

// HasAny returns true if the set contains at least one of the given permissions.
func (ps PermissionSet) HasAny(perms ...string) bool {
	for _, p := range perms {
		if ps.Has(p) {
			return true
		}
	}
	return false
}

// HasAll returns true if the set contains all of the given permissions.
func (ps PermissionSet) HasAll(perms ...string) bool {
	for _, p := range perms {
		if !ps.Has(p) {
			return false
		}
	}
	return true
}

// PermissionCache holds role-to-permissions mappings in memory.
// Safe for concurrent reads. Refreshed via Load() using build-then-swap.
type PermissionCache struct {
	mu          sync.RWMutex
	cache       map[types.RoleID]PermissionSet
	adminRoleID types.RoleID
	isAdmin     map[types.RoleID]bool
	lastLoaded  time.Time
}

// NewPermissionCache creates an empty PermissionCache.
func NewPermissionCache() *PermissionCache {
	return &PermissionCache{
		cache:   make(map[types.RoleID]PermissionSet),
		isAdmin: make(map[types.RoleID]bool),
	}
}

// Load populates the cache from the database using build-then-swap.
// Readers are never blocked during the DB queries.
func (pc *PermissionCache) Load(driver db.DbDriver) error {
	// Build new cache (no lock held)
	newCache := make(map[types.RoleID]PermissionSet)
	var newAdminRoleID types.RoleID
	newIsAdmin := make(map[types.RoleID]bool)

	roles, err := driver.ListRoles()
	if err != nil {
		return fmt.Errorf("loading permission cache: %w", err)
	}
	if roles == nil {
		return fmt.Errorf("loading permission cache: ListRoles returned nil")
	}
	if len(*roles) > 1000 {
		return fmt.Errorf("refusing to load permission cache: %d roles exceeds safety limit of 1000", len(*roles))
	}

	for _, role := range *roles {
		if role.Label == "admin" {
			newAdminRoleID = role.RoleID
			newIsAdmin[role.RoleID] = true
		}
		labels, err := driver.ListPermissionLabelsByRoleID(role.RoleID)
		if err != nil {
			return fmt.Errorf("loading permissions for role %s: %w", role.RoleID, err)
		}
		ps := make(PermissionSet, len(*labels))
		for _, label := range *labels {
			ps[label] = struct{}{}
		}
		newCache[role.RoleID] = ps
	}

	// Swap under write lock (nanoseconds)
	pc.mu.Lock()
	pc.cache = newCache
	pc.adminRoleID = newAdminRoleID
	pc.isAdmin = newIsAdmin
	pc.lastLoaded = time.Now()
	pc.mu.Unlock()

	return nil
}

// PermissionsForRole returns the PermissionSet for the given role.
// Returns nil if the role is not in the cache.
func (pc *PermissionCache) PermissionsForRole(roleID types.RoleID) PermissionSet {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.cache[roleID]
}

// IsAdmin returns true if the given role is the admin role.
func (pc *PermissionCache) IsAdmin(roleID types.RoleID) bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.isAdmin[roleID]
}

// StartPeriodicRefresh starts a background goroutine that refreshes
// the cache at the given interval. Stops when ctx is cancelled.
func (pc *PermissionCache) StartPeriodicRefresh(ctx context.Context, driver db.DbDriver, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		consecutiveFailures := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := pc.Load(driver); err != nil {
					consecutiveFailures++
					if consecutiveFailures <= 3 || consecutiveFailures%30 == 0 {
						utility.DefaultLogger.Error("periodic permission cache refresh failed",
							err,
							"consecutive_failures", consecutiveFailures,
						)
					}
				} else {
					if consecutiveFailures > 0 {
						utility.DefaultLogger.Info("periodic permission cache refresh recovered",
							"after_failures", consecutiveFailures,
						)
					}
					consecutiveFailures = 0
				}
			}
		}
	}()
}

// ContextPermissions extracts the PermissionSet from the request context.
// Returns nil if no PermissionSet is present.
func ContextPermissions(ctx context.Context) PermissionSet {
	ps, _ := ctx.Value(permissionsKey).(PermissionSet)
	return ps
}

// ContextIsAdmin returns true if the request context indicates an admin user.
func ContextIsAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(isAdminKey).(bool)
	return v
}

// PermissionInjector resolves the user's role to a PermissionSet and stores
// it in context. Must run after HTTPAuthenticationMiddleware.
// Short-circuits for unauthenticated requests: if no user is in context,
// the handler chain continues immediately with no PermissionSet in context.
func PermissionInjector(pc *PermissionCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := AuthenticatedUser(r.Context())
			if user == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			roleID := types.RoleID(user.Role)
			ps := pc.PermissionsForRole(roleID)
			ctx = context.WithValue(ctx, permissionsKey, ps)
			ctx = context.WithValue(ctx, isAdminKey, pc.IsAdmin(roleID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeForbidden writes a 403 JSON response and logs the authorization failure.
func writeForbidden(w http.ResponseWriter, r *http.Request, permission string) {
	user := AuthenticatedUser(r.Context())
	userID := "anonymous"
	roleID := ""
	if user != nil {
		userID = user.UserID.String()
		roleID = user.Role
	}

	utility.DefaultLogger.Warn("authorization denied",
		nil,
		"user_id", userID,
		"role_id", roleID,
		"required_permission", permission,
		"path", r.URL.Path,
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	// Encode error is non-recoverable (client disconnected or similar);
	// the response is already partially written so no recovery is possible.
	json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
}

// RequirePermission checks for a single permission.
// Admin bypass is checked via ContextIsAdmin, not via a "*" key in PermissionSet.
// Fail-closed: returns 403 if no PermissionSet is found in context.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ContextIsAdmin(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
			ps := ContextPermissions(r.Context())
			if ps == nil || !ps.Has(permission) {
				writeForbidden(w, r, permission)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission checks for at least one permission (OR logic).
// Fail-closed: returns 403 if no PermissionSet is found in context.
func RequireAnyPermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ContextIsAdmin(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
			ps := ContextPermissions(r.Context())
			if ps == nil || !ps.HasAny(permissions...) {
				writeForbidden(w, r, fmt.Sprintf("any(%v)", permissions))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions checks for all permissions (AND logic).
// Fail-closed: returns 403 if no PermissionSet is found in context.
func RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ContextIsAdmin(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}
			ps := ContextPermissions(r.Context())
			if ps == nil || !ps.HasAll(permissions...) {
				writeForbidden(w, r, fmt.Sprintf("all(%v)", permissions))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// methodToOperation maps HTTP methods to resource operations.
var methodToOperation = map[string]string{
	http.MethodGet:    "read",
	http.MethodPost:   "create",
	http.MethodPut:    "update",
	http.MethodPatch:  "update",
	http.MethodDelete: "delete",
}

// RequireResourcePermission maps HTTP method to resource:operation automatically.
// GET -> resource:read, POST -> resource:create, PUT/PATCH -> resource:update,
// DELETE -> resource:delete. Unmapped methods are denied (403).
// Fail-closed: returns 403 if no PermissionSet is found in context.
func RequireResourcePermission(resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ContextIsAdmin(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}

			op, ok := methodToOperation[r.Method]
			if !ok {
				writeForbidden(w, r, resource+":unknown")
				return
			}

			perm := resource + ":" + op
			ps := ContextPermissions(r.Context())
			if ps == nil || !ps.Has(perm) {
				writeForbidden(w, r, perm)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidatePermissionLabel validates that a permission label matches the
// format "resource:operation" using character-by-character validation (no regex).
// Resource part: lowercase alphanumeric and underscores [a-z0-9_].
// Operation part: lowercase alphabetic [a-z].
// The literal string "*" is rejected.
func ValidatePermissionLabel(label string) error {
	if label == "" {
		return fmt.Errorf("permission label cannot be empty")
	}
	if label == "*" {
		return fmt.Errorf("permission label cannot be '*'")
	}

	// Find the colon separator
	colonIdx := -1
	for i := range len(label) {
		if label[i] == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return fmt.Errorf("permission label must contain a colon separator")
	}
	if colonIdx == 0 {
		return fmt.Errorf("permission label resource part cannot be empty")
	}
	if colonIdx == len(label)-1 {
		return fmt.Errorf("permission label operation part cannot be empty")
	}

	// Validate resource part: [a-z0-9_]
	resource := label[:colonIdx]
	for i := range len(resource) {
		c := resource[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			continue
		}
		return fmt.Errorf("permission label resource contains invalid character '%c' at position %d", c, i)
	}

	// Validate operation part: [a-z]
	operation := label[colonIdx+1:]
	for i := range len(operation) {
		c := operation[i]
		if c >= 'a' && c <= 'z' {
			continue
		}
		return fmt.Errorf("permission label operation contains invalid character '%c' at position %d", c, i)
	}

	return nil
}
