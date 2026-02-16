package router

import (
	"encoding/json"
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/plugin"
	"golang.org/x/time/rate"
)

func NewModulacmsMux(mgr *config.Manager, bridge *plugin.HTTPBridge) *http.ServeMux {
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
	/*
		mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, *c)
		})
	*/
	// Auth endpoints with CORS and rate limiting
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
	mux.HandleFunc("/api/v1/admin/tree/", func(w http.ResponseWriter, r *http.Request) {
		AdminTreeHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admincontentdatas", func(w http.ResponseWriter, r *http.Request) {
		AdminContentDatasHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admincontentdatas/", func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admincontentfields", func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admincontentfields/", func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admindatatypes", func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypesHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admindatatypes/", func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/adminfields", func(w http.ResponseWriter, r *http.Request) {
		AdminFieldsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/adminfields/", func(w http.ResponseWriter, r *http.Request) {
		AdminFieldHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admindatatypefields", func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeFieldsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/admindatatypefields/", func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeFieldHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/adminroutes", func(w http.ResponseWriter, r *http.Request) {
		AdminRoutesHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/adminroutes/", func(w http.ResponseWriter, r *http.Request) {
		AdminRouteHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/contentdata", func(w http.ResponseWriter, r *http.Request) {
		ContentDatasHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/contentdata/", func(w http.ResponseWriter, r *http.Request) {
		ContentDataHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/contentfields", func(w http.ResponseWriter, r *http.Request) {
		ContentFieldsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/contentfields/", func(w http.ResponseWriter, r *http.Request) {
		ContentFieldHandler(w, r, *c)
	})
	mux.HandleFunc("POST /api/v1/content/batch", func(w http.ResponseWriter, r *http.Request) {
		ContentBatchHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/datatype", func(w http.ResponseWriter, r *http.Request) {
		DatatypesHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/datatype/", func(w http.ResponseWriter, r *http.Request) {
		DatatypeHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/datatypefields", func(w http.ResponseWriter, r *http.Request) {
		DatatypeFieldsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/datatypefields/", func(w http.ResponseWriter, r *http.Request) {
		DatatypeFieldHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/fields", func(w http.ResponseWriter, r *http.Request) {
		FieldsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/fields/", func(w http.ResponseWriter, r *http.Request) {
		FieldHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/media", func(w http.ResponseWriter, r *http.Request) {
		MediasHandler(w, r, *c)
	})
	mux.HandleFunc("GET /api/v1/media/health", func(w http.ResponseWriter, r *http.Request) {
		MediaHealthHandler(w, r, *c)
	})
	mux.HandleFunc("DELETE /api/v1/media/cleanup", func(w http.ResponseWriter, r *http.Request) {
		MediaCleanupHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		MediaHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/mediadimensions", func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/mediadimensions/", func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/routes", func(w http.ResponseWriter, r *http.Request) {
		RoutesHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/routes/", func(w http.ResponseWriter, r *http.Request) {
		RouteHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
		RolesHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/roles/", func(w http.ResponseWriter, r *http.Request) {
		RoleHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		SessionsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/sessions/", func(w http.ResponseWriter, r *http.Request) {
		SessionHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/tables", func(w http.ResponseWriter, r *http.Request) {
		TablesHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/tables/", func(w http.ResponseWriter, r *http.Request) {
		TableHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		TokensHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/tokens/", func(w http.ResponseWriter, r *http.Request) {
		TokenHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/usersoauth", func(w http.ResponseWriter, r *http.Request) {
		UserOauthsHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/usersoauth/", func(w http.ResponseWriter, r *http.Request) {
		UserOauthHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		UsersHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		UserHandler(w, r, *c)
	})

	// SSH Key management endpoints (require authentication)
	mux.HandleFunc("POST /api/v1/ssh-keys", func(w http.ResponseWriter, r *http.Request) {
		AddSSHKeyHandler(w, r, *c)
	})
	mux.HandleFunc("GET /api/v1/ssh-keys", func(w http.ResponseWriter, r *http.Request) {
		ListSSHKeysHandler(w, r, *c)
	})
	mux.HandleFunc("DELETE /api/v1/ssh-keys/", func(w http.ResponseWriter, r *http.Request) {
		DeleteSSHKeyHandler(w, r, *c)
	})

	// Import endpoints - accept CMS format and transform to ModulaCMS
	mux.HandleFunc("/api/v1/import/contentful", func(w http.ResponseWriter, r *http.Request) {
		ImportContentfulHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/import/sanity", func(w http.ResponseWriter, r *http.Request) {
		ImportSanityHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/import/strapi", func(w http.ResponseWriter, r *http.Request) {
		ImportStrapiHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/import/wordpress", func(w http.ResponseWriter, r *http.Request) {
		ImportWordPressHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/import/clean", func(w http.ResponseWriter, r *http.Request) {
		ImportCleanHandler(w, r, *c)
	})
	mux.HandleFunc("/api/v1/import", func(w http.ResponseWriter, r *http.Request) {
		ImportBulkHandler(w, r, *c)
	})

	//mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	//mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)

	// Config management endpoints (admin-only)
	configAuthChain := middleware.AuthenticatedChain(mgr)
	mux.Handle("GET /api/v1/admin/config", configAuthChain(adminOnly(ConfigGetHandler(mgr))))
	mux.Handle("PATCH /api/v1/admin/config", configAuthChain(adminOnly(ConfigUpdateHandler(mgr))))
	mux.Handle("GET /api/v1/admin/config/meta", configAuthChain(adminOnly(ConfigMetaHandler())))

	// Plugin HTTP bridge routes and admin endpoints
	if bridge != nil {
		bridge.MountOn(mux)

		// Admin route management endpoints
		authChain := middleware.AuthenticatedChain(mgr)
		mux.Handle("GET /api/v1/admin/plugins/routes", authChain(pluginRoutesListHandler(bridge)))
		mux.Handle("POST /api/v1/admin/plugins/routes/approve", authChain(adminOnly(pluginRoutesApproveHandler(bridge))))
		mux.Handle("POST /api/v1/admin/plugins/routes/revoke", authChain(adminOnly(pluginRoutesRevokeHandler(bridge))))

		// Phase 4: Mount plugin management admin endpoints via bridge.
		// The adminOnly wrapper is passed as a parameter since it is defined
		// here (router-layer concern, not plugin-layer concern).
		bridge.MountAdminEndpoints(mux, authChain, adminOnly)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		SlugHandler(w, r, *c)
	})
	return mux

}

// adminOnly is a middleware wrapper that requires the authenticated user to
// have the "admin" role. Returns 403 Forbidden if the user does not have
// admin privileges. Must be used after AuthenticatedChain which ensures a
// user exists in the context.
func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := middleware.AuthenticatedUser(r.Context())
		if user == nil || user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
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
