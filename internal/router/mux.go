package router

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/admin"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/plugin"
	"golang.org/x/time/rate"
)

func NewModulacmsMux(mgr *config.Manager, bridge *plugin.HTTPBridge, driver db.DbDriver, pc *middleware.PermissionCache) *http.ServeMux {
	mux := http.NewServeMux()

	c, err := mgr.Config()
	if err != nil {
		// Without config the router cannot function. Return a mux with a
		// single catch-all that returns 500.
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "configuration unavailable", http.StatusInternalServerError)
		})
		return mux
	}

	// Create rate limiter for auth endpoints: 10 requests per minute per IP
	authLimiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10)

	// Create CORS middleware
	corsMiddleware := middleware.CorsMiddleware(c)

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	// Auth endpoints with CORS and rate limiting (PUBLIC - no auth/permission required)
	mux.Handle("POST /api/v1/auth/login", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, *c)
	}))))
	mux.Handle("POST /api/v1/auth/logout", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, *c)
	}))))
	mux.Handle("GET /api/v1/auth/me", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MeHandler(w, r, *c)
	}))))
	mux.Handle("POST /api/v1/auth/register", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, *c)
	}))))
	mux.Handle("POST /api/v1/auth/reset", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ResetPasswordHandler(w, r, *c)
	}))))

	// OAuth endpoints with CORS and rate limiting (PUBLIC - no auth required)
	mux.Handle("GET /api/v1/auth/oauth/login", corsMiddleware(authLimiter.Middleware(OauthInitiateHandler(*c))))
	mux.Handle("GET /api/v1/auth/oauth/callback", corsMiddleware(authLimiter.Middleware(OauthCallbackHandler(*c))))

	// Admin tree
	mux.Handle("/api/v1/admin/tree/", middleware.RequireResourcePermission("admin_tree")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminTreeHandler(w, r, *c)
	})))

	// Admin content data
	mux.Handle("/api/v1/admincontentdatas", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDatasHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/admincontentdatas/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataHandler(w, r, *c)
	})))

	// Admin content fields
	mux.Handle("/api/v1/admincontentfields", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/admincontentfields/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldHandler(w, r, *c)
	})))

	// Admin datatypes
	mux.Handle("/api/v1/admindatatypes", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypesHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/admindatatypes/", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeHandler(w, r, *c)
	})))

	// Admin fields
	mux.Handle("/api/v1/adminfields", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminFieldsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/adminfields/", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminFieldHandler(w, r, *c)
	})))

	// Admin datatype fields
	mux.Handle("/api/v1/admindatatypefields", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeFieldsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/admindatatypefields/", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeFieldHandler(w, r, *c)
	})))

	// Admin routes
	mux.Handle("/api/v1/adminroutes", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminRoutesHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/adminroutes/", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminRouteHandler(w, r, *c)
	})))

	// Content data
	mux.Handle("/api/v1/contentdata", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDatasHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/contentdata/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataHandler(w, r, *c)
	})))

	// Content fields
	mux.Handle("/api/v1/contentfields", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentFieldsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/contentfields/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentFieldHandler(w, r, *c)
	})))

	// Content batch
	mux.Handle("POST /api/v1/content/batch", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentBatchHandler(w, r, *c)
	})))

	// Content data reorder
	mux.Handle("POST /api/v1/contentdata/reorder", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataReorderHandler(w, r, *c)
	})))

	// Admin content data reorder
	mux.Handle("POST /api/v1/admincontentdatas/reorder", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataReorderHandler(w, r, *c)
	})))

	// Datatypes
	mux.Handle("/api/v1/datatype", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypesHandler(w, r, *c)
	})))
	mux.Handle("GET /api/v1/datatype/full", middleware.RequirePermission("datatypes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeFullHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/datatype/", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeHandler(w, r, *c)
	})))

	// Datatype fields
	mux.Handle("/api/v1/datatypefields", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeFieldsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/datatypefields/", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeFieldHandler(w, r, *c)
	})))

	// Fields
	mux.Handle("/api/v1/fields", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/fields/", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldHandler(w, r, *c)
	})))

	// Media
	mux.Handle("/api/v1/media", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediasHandler(w, r, *c)
	})))
	mux.Handle("GET /api/v1/media/health", middleware.RequirePermission("media:admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaHealthHandler(w, r, *c)
	})))
	mux.Handle("DELETE /api/v1/media/cleanup", middleware.RequirePermission("media:admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaCleanupHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/media/", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaHandler(w, r, *c)
	})))

	// Media dimensions
	mux.Handle("/api/v1/mediadimensions", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/mediadimensions/", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionHandler(w, r, *c)
	})))

	// Routes
	mux.Handle("/api/v1/routes", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RoutesHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/routes/", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RouteHandler(w, r, *c)
	})))

	// Roles (handlers receive pc for cache refresh)
	mux.Handle("/api/v1/roles", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolesHandler(w, r, *c, pc)
	})))
	mux.Handle("/api/v1/roles/", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RoleHandler(w, r, *c, pc)
	})))

	// Permissions (handlers receive pc for cache refresh)
	mux.Handle("/api/v1/permissions", middleware.RequireResourcePermission("permissions")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PermissionsHandler(w, r, *c, pc)
	})))
	mux.Handle("/api/v1/permissions/", middleware.RequireResourcePermission("permissions")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PermissionHandler(w, r, *c, pc)
	})))

	// Sessions
	mux.Handle("/api/v1/sessions", middleware.RequirePermission("sessions:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SessionsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/sessions/", middleware.RequireResourcePermission("sessions")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SessionHandler(w, r, *c)
	})))

	// Tables
	mux.Handle("/api/v1/tables", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TablesHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/tables/", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TableHandler(w, r, *c)
	})))

	// Tokens
	mux.Handle("/api/v1/tokens", middleware.RequireResourcePermission("tokens")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TokensHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/tokens/", middleware.RequireResourcePermission("tokens")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TokenHandler(w, r, *c)
	})))

	// Users OAuth
	mux.Handle("/api/v1/usersoauth", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserOauthsHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/usersoauth/", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserOauthHandler(w, r, *c)
	})))

	// Users
	mux.Handle("/api/v1/users", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UsersHandler(w, r, *c)
	})))
	mux.Handle("GET /api/v1/users/full", middleware.RequirePermission("users:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UsersFullHandler(w, r, *c)
	})))
	mux.Handle("GET /api/v1/users/full/", middleware.RequirePermission("users:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserFullHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/users/", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserHandler(w, r, *c)
	})))

	// SSH Key management endpoints
	mux.Handle("POST /api/v1/ssh-keys", middleware.RequirePermission("ssh_keys:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddSSHKeyHandler(w, r, *c)
	})))
	mux.Handle("GET /api/v1/ssh-keys", middleware.RequirePermission("ssh_keys:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ListSSHKeysHandler(w, r, *c)
	})))
	mux.Handle("DELETE /api/v1/ssh-keys/", middleware.RequirePermission("ssh_keys:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteSSHKeyHandler(w, r, *c)
	})))

	// Import endpoints
	mux.Handle("/api/v1/import/contentful", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportContentfulHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/import/sanity", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportSanityHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/import/strapi", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportStrapiHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/import/wordpress", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportWordPressHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/import/clean", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportCleanHandler(w, r, *c)
	})))
	mux.Handle("/api/v1/import", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportBulkHandler(w, r, *c)
	})))

	// Role-permissions junction table CRUD
	mux.Handle("/api/v1/role-permissions", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolePermissionsHandler(w, r, *c, pc)
	})))
	mux.Handle("/api/v1/role-permissions/", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolePermissionHandler(w, r, *c, pc)
	})))
	mux.Handle("GET /api/v1/role-permissions/role/", middleware.RequirePermission("roles:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolePermissionsByRoleHandler(w, r, *c)
	})))

	// Config management endpoints (permission-gated)
	configAuthChain := middleware.AuthenticatedChain(mgr)
	mux.Handle("GET /api/v1/admin/config", configAuthChain(middleware.RequirePermission("config:read")(ConfigGetHandler(mgr))))
	mux.Handle("PATCH /api/v1/admin/config", configAuthChain(middleware.RequirePermission("config:update")(ConfigUpdateHandler(mgr))))
	mux.Handle("GET /api/v1/admin/config/meta", configAuthChain(middleware.RequirePermission("config:read")(ConfigMetaHandler())))

	// Plugin HTTP bridge routes and admin endpoints
	if bridge != nil {
		bridge.MountOn(mux)

		// Admin route management endpoints
		authChain := middleware.AuthenticatedChain(mgr)
		mux.Handle("GET /api/v1/admin/plugins/routes", authChain(middleware.RequirePermission("plugins:read")(pluginRoutesListHandler(bridge))))
		mux.Handle("POST /api/v1/admin/plugins/routes/approve", authChain(middleware.RequirePermission("plugins:admin")(pluginRoutesApproveHandler(bridge))))
		mux.Handle("POST /api/v1/admin/plugins/routes/revoke", authChain(middleware.RequirePermission("plugins:admin")(pluginRoutesRevokeHandler(bridge))))

		// Phase 4: Mount plugin management admin endpoints via bridge.
		bridge.MountAdminEndpoints(mux, authChain,
			middleware.RequirePermission("plugins:read"),
			middleware.RequirePermission("plugins:admin"),
		)
	}

	// Content delivery via slug
	mux.HandleFunc("/api/v1/content/", func(w http.ResponseWriter, r *http.Request) {
		SlugHandler(w, r, *c)
	})

	// Embedded admin panel (SPA)
	distFS, err := fs.Sub(admin.DistFS, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distFS))
		mux.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
			// Strip the /admin prefix for filesystem lookups
			path := strings.TrimPrefix(r.URL.Path, "/admin")
			if path == "" || path == "/" {
				path = "/index.html"
			}
			// Try to open the file; if it doesn't exist, serve index.html for SPA routing
			f, openErr := distFS.Open(strings.TrimPrefix(path, "/"))
			if openErr != nil {
				r.URL.Path = "/index.html"
				fileServer.ServeHTTP(w, r)
				return
			}
			f.Close()
			// File exists — let the file server handle it (with /admin/ stripped)
			r.URL.Path = path
			fileServer.ServeHTTP(w, r)
		})
	}

	// Root redirects to admin panel
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusFound)
	})
	return mux

}

// pluginRoutesListHandler returns all registered plugin routes with approval status.
func pluginRoutesListHandler(bridge *plugin.HTTPBridge) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		routes := bridge.ListRoutes()

		type routeJSON struct {
			Plugin   string `json:"plugin"`
			Method   string `json:"method"`
			Path     string `json:"path"`
			Public   bool   `json:"public"`
			Approved bool   `json:"approved"`
		}

		result := make([]routeJSON, 0, len(routes))
		for _, route := range routes {
			result = append(result, routeJSON{
				Plugin:   route.PluginName,
				Method:   route.Method,
				Path:     route.Path,
				Public:   route.Public,
				Approved: route.Approved,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"routes": result})
	})
}

// pluginRoutesApproveHandler approves one or more plugin routes.
func pluginRoutesApproveHandler(bridge *plugin.HTTPBridge) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body size limit: 1 MB. Must be applied before json.NewDecoder.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req struct {
			Routes []struct {
				Plugin string `json:"plugin"`
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"routes"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		approvedBy := ""
		if user != nil {
			approvedBy = user.Username
		}

		var errs []string
		for _, route := range req.Routes {
			if err := bridge.ApproveRoute(r.Context(), route.Plugin, route.Method, route.Path, approvedBy); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if len(errs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			// Encode error is non-recoverable (client disconnected or similar);
			// the response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{"errors": errs})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
}

// pluginRoutesRevokeHandler revokes approval for one or more plugin routes.
func pluginRoutesRevokeHandler(bridge *plugin.HTTPBridge) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Body size limit: 1 MB. Must be applied before json.NewDecoder.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req struct {
			Routes []struct {
				Plugin string `json:"plugin"`
				Method string `json:"method"`
				Path   string `json:"path"`
			} `json:"routes"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var errs []string
		for _, route := range req.Routes {
			if err := bridge.RevokeRoute(r.Context(), route.Plugin, route.Method, route.Path); err != nil {
				errs = append(errs, err.Error())
			}
		}

		if len(errs) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			// Encode error is non-recoverable (client disconnected or similar);
			// the response is already partially written so no recovery is possible.
			json.NewEncoder(w).Encode(map[string]any{"errors": errs})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Encode error is non-recoverable (client disconnected or similar);
		// the response is already partially written so no recovery is possible.
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
}
