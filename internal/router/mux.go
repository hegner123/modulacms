package router

import (
	"encoding/json"
	"net/http"

	htmxadmin "github.com/hegner123/modulacms/internal/admin"
	adminhandlers "github.com/hegner123/modulacms/internal/admin/handlers"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/deploy"
	"github.com/hegner123/modulacms/internal/email"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/search"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/time/rate"
)

func NewModulaMux(mgr *config.Manager, bridge *plugin.HTTPBridge, driver db.DbDriver, pc *middleware.PermissionCache, emailSvc *email.Service, dispatcher publishing.WebhookDispatcher, svc *service.Registry, searchSvc *search.Service) *http.ServeMux {
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
		LoginHandler(w, r, svc)
	}))))
	mux.Handle("POST /api/v1/auth/logout", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, svc)
	}))))
	mux.Handle("GET /api/v1/auth/me", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MeHandler(w, r, svc)
	}))))
	mux.Handle("POST /api/v1/auth/register", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, svc)
	}))))
	mux.Handle("POST /api/v1/auth/reset", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ResetPasswordHandler(w, r, svc)
	}))))
	mux.Handle("POST /api/v1/auth/request-password-reset", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RequestPasswordResetHandler(w, r, svc)
	}))))
	mux.Handle("POST /api/v1/auth/confirm-password-reset", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ConfirmPasswordResetHandler(w, r, svc)
	}))))

	// OAuth endpoints with CORS and rate limiting (PUBLIC - no auth required)
	mux.Handle("GET /api/v1/auth/oauth/login", corsMiddleware(authLimiter.Middleware(OauthInitiateHandler(svc))))
	mux.Handle("GET /api/v1/auth/oauth/callback", corsMiddleware(authLimiter.Middleware(OauthCallbackHandler(svc))))

	// Health check (PUBLIC - no auth required)
	var pluginHealthFn PluginHealthChecker
	if bridge != nil {
		pluginHealthFn = func() PluginHealthResult {
			s := bridge.PluginHealth()
			return PluginHealthResult{
				Healthy:             s.Healthy,
				FailedPlugins:       s.FailedPlugins,
				OpenCircuitBreakers: s.OpenCircuitBreakers,
			}
		}
	}
	mux.Handle("GET /api/v1/health", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HealthHandler(w, r, svc, pluginHealthFn)
	})))

	// Admin tree
	mux.Handle("/api/v1/admin/tree/", middleware.RequireResourcePermission("admin_tree")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminTreeHandler(w, r, svc)
	})))

	// Admin content data
	mux.Handle("/api/v1/admincontentdatas", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDatasHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/admincontentdatas/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admincontentdatas/full", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataFullHandler(w, r, svc)
	})))

	// Admin content fields
	mux.Handle("/api/v1/admincontentfields", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/admincontentfields/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldHandler(w, r, svc)
	})))

	// Admin datatypes
	mux.Handle("/api/v1/admindatatypes", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypesHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admindatatypes/max-sort-order", middleware.RequirePermission("datatypes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeMaxSortOrderHandler(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/admindatatypes/{id}/sort-order", middleware.RequirePermission("datatypes:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeSortOrderHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/admindatatypes/", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admindatatypes/full", middleware.RequirePermission("datatypes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeFullHandler(w, r, svc)
	})))

	// Admin fields
	mux.Handle("/api/v1/adminfields", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminFieldsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/adminfields/", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminFieldHandler(w, r, svc)
	})))

	// Admin routes
	mux.Handle("/api/v1/adminroutes", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminRoutesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/adminroutes/", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminRouteHandler(w, r, svc)
	})))

	// Content data
	mux.Handle("/api/v1/contentdata", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDatasHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/contentdata/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/contentdata/full", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataFullHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/contentdata/by-route", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataByRouteHandler(w, r, svc)
	})))

	// Content fields
	mux.Handle("/api/v1/contentfields", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentFieldsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/contentfields/", middleware.RequireResourcePermission("content")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentFieldHandler(w, r, svc)
	})))

	// Content create with fields (composite)
	mux.Handle("POST /api/v1/content/create", middleware.RequirePermission("content:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentCreateHandler(w, r, svc)
	})))

	// Content batch
	mux.Handle("POST /api/v1/content/batch", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentBatchHandler(w, r, svc)
	})))

	// Content tree save (bulk pointer updates + deletes)
	mux.Handle("POST /api/v1/content/tree", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentTreeSaveHandler(w, r, svc)
	})))

	// Content tree get (read tree by route ID)
	mux.Handle("GET /api/v1/content/tree/{routeID}", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentTreeGetHandler(w, r, svc)
	})))

	// Content publish / unpublish / schedule
	mux.Handle("POST /api/v1/content/publish", middleware.RequirePermission("content:publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PublishHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/content/unpublish", middleware.RequirePermission("content:publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UnpublishHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/content/schedule", middleware.RequirePermission("content:publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ScheduleHandler(w, r, svc)
	})))

	// Content versions list (filtered by content_id)
	mux.Handle("GET /api/v1/contentversions", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentVersionsListHandler(w, r, svc)
	})))

	// Content version management
	mux.Handle("GET /api/v1/content/versions", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ListVersionsHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/content/versions/", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetVersionHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/content/versions", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CreateManualVersionHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/content/versions/", middleware.RequirePermission("content:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteVersionHandler(w, r, svc)
	})))

	// Content restore from version
	mux.Handle("POST /api/v1/content/restore", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RestoreVersionHandler(w, r, svc)
	})))

	// Admin content publish / unpublish / schedule
	mux.Handle("POST /api/v1/admin/content/publish", middleware.RequirePermission("content:publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminPublishHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/content/unpublish", middleware.RequirePermission("content:publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminUnpublishHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/content/schedule", middleware.RequirePermission("content:publish")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminScheduleHandler(w, r, svc)
	})))

	// Admin content version management
	mux.Handle("GET /api/v1/admin/content/versions", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminListVersionsHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admin/content/versions/", middleware.RequirePermission("content:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminGetVersionHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/content/versions", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminCreateManualVersionHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/admin/content/versions/", middleware.RequirePermission("content:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminDeleteVersionHandler(w, r, svc)
	})))

	// Admin content restore from version
	mux.Handle("POST /api/v1/admin/content/restore", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminRestoreVersionHandler(w, r, svc)
	})))

	// Content data reorder
	mux.Handle("POST /api/v1/contentdata/reorder", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataReorderHandler(w, r, svc)
	})))

	// Admin content data reorder
	mux.Handle("POST /api/v1/admincontentdatas/reorder", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataReorderHandler(w, r, svc)
	})))

	// Content tree heal (admin repair of malformed IDs)
	mux.Handle("POST /api/v1/admin/content/heal", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentHealHandler(w, r, svc)
	})))

	// Content data move (cross-parent)
	mux.Handle("POST /api/v1/contentdata/move", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ContentDataMoveHandler(w, r, svc)
	})))

	// Admin content data move (cross-parent)
	mux.Handle("POST /api/v1/admincontentdatas/move", middleware.RequirePermission("content:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataMoveHandler(w, r, svc)
	})))

	// Datatypes
	mux.Handle("/api/v1/datatype", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypesHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/datatype/full", middleware.RequirePermission("datatypes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeFullHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/datatype/max-sort-order", middleware.RequirePermission("datatypes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeMaxSortOrderHandler(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/datatype/{id}/sort-order", middleware.RequirePermission("datatypes:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeSortOrderHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/datatype/full/list", middleware.RequirePermission("datatypes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypesFullListHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/datatype/", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DatatypeHandler(w, r, svc)
	})))

	// Fields
	mux.Handle("/api/v1/fields", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldsHandler(w, r, svc)
	})))

	// Field sort order update
	mux.Handle("PUT /api/v1/fields/{id}/sort-order", middleware.RequirePermission("fields:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldSortOrderHandler(w, r, svc)
	})))

	// Field max sort order query
	mux.Handle("GET /api/v1/fields/max-sort-order", middleware.RequirePermission("fields:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldMaxSortOrderHandler(w, r, svc)
	})))

	mux.Handle("/api/v1/fields/", middleware.RequireResourcePermission("fields")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldHandler(w, r, svc)
	})))

	// Field types
	mux.Handle("/api/v1/fieldtypes", middleware.RequireResourcePermission("field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldTypesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/fieldtypes/", middleware.RequireResourcePermission("field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		FieldTypeHandler(w, r, svc)
	})))

	// Admin field types
	mux.Handle("/api/v1/adminfieldtypes", middleware.RequireResourcePermission("admin_field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminFieldTypesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/adminfieldtypes/", middleware.RequireResourcePermission("admin_field_types")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminFieldTypeHandler(w, r, svc)
	})))

	// Media
	mux.Handle("GET /api/v1/media/{id}/download", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiDownloadMedia(w, r, svc)
	})))
	mux.Handle("/api/v1/media", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediasHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/media/full", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaFullHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/media/references", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaReferencesHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/media/health", middleware.RequirePermission("media:admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaHealthHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/media/cleanup", middleware.RequirePermission("media:admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaCleanupHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/media/move", middleware.RequirePermission("media:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiBatchMoveMedia(w, r, svc)
	})))
	mux.Handle("/api/v1/media/", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaHandler(w, r, svc)
	})))

	// Media folders (tree must be registered before {id} to avoid path parameter matching "tree")
	mux.Handle("GET /api/v1/media-folders", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiListMediaFolders(w, r, svc)
	})))
	mux.Handle("GET /api/v1/media-folders/tree", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiMediaFolderTree(w, r, svc)
	})))
	mux.Handle("POST /api/v1/media-folders", middleware.RequirePermission("media:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCreateMediaFolder(w, r, svc)
	})))
	mux.Handle("GET /api/v1/media-folders/{id}", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiGetMediaFolder(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/media-folders/{id}", middleware.RequirePermission("media:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiUpdateMediaFolder(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/media-folders/{id}", middleware.RequirePermission("media:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiDeleteMediaFolder(w, r, svc)
	})))
	mux.Handle("GET /api/v1/media-folders/{id}/media", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiMediaFolderMedia(w, r, svc)
	})))

	// Media dimensions
	mux.Handle("/api/v1/mediadimensions", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/mediadimensions/", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionHandler(w, r, svc)
	})))

	// Admin media
	mux.Handle("GET /api/v1/adminmedia/{id}/download", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiDownloadAdminMedia(w, r, svc)
	})))
	mux.Handle("POST /api/v1/adminmedia/move", middleware.RequirePermission("media:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiBatchMoveAdminMedia(w, r, svc)
	})))
	mux.Handle("/api/v1/adminmedia", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminMediasHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/adminmedia/", middleware.RequireResourcePermission("media")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminMediaHandler(w, r, svc)
	})))

	// Admin media folders (tree must be registered before {id} to avoid path parameter matching "tree")
	mux.Handle("GET /api/v1/adminmedia-folders", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiListAdminMediaFolders(w, r, svc)
	})))
	mux.Handle("GET /api/v1/adminmedia-folders/tree", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiAdminMediaFolderTree(w, r, svc)
	})))
	mux.Handle("POST /api/v1/adminmedia-folders", middleware.RequirePermission("media:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCreateAdminMediaFolder(w, r, svc)
	})))
	mux.Handle("GET /api/v1/adminmedia-folders/{id}", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiGetAdminMediaFolder(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/adminmedia-folders/{id}", middleware.RequirePermission("media:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiUpdateAdminMediaFolder(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/adminmedia-folders/{id}", middleware.RequirePermission("media:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiDeleteAdminMediaFolder(w, r, svc)
	})))
	mux.Handle("GET /api/v1/adminmedia-folders/{id}/media", middleware.RequirePermission("media:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiAdminMediaFolderMedia(w, r, svc)
	})))

	// Routes
	mux.Handle("/api/v1/routes", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RoutesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/routes/", middleware.RequireResourcePermission("routes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RouteHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/routes/full", middleware.RequirePermission("routes:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RouteFullHandler(w, r, svc)
	})))

	// Roles
	mux.Handle("/api/v1/roles", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/roles/", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RoleHandler(w, r, svc)
	})))

	// Permissions
	mux.Handle("/api/v1/permissions", middleware.RequireResourcePermission("permissions")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PermissionsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/permissions/", middleware.RequireResourcePermission("permissions")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		PermissionHandler(w, r, svc)
	})))

	// Sessions
	mux.Handle("/api/v1/sessions", middleware.RequirePermission("sessions:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SessionsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/sessions/", middleware.RequireResourcePermission("sessions")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SessionHandler(w, r, svc)
	})))

	// Tables
	mux.Handle("/api/v1/tables", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TablesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/tables/", middleware.RequireResourcePermission("datatypes")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TableHandler(w, r, svc)
	})))

	// Tokens
	mux.Handle("/api/v1/tokens", middleware.RequireResourcePermission("tokens")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TokensHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/tokens/", middleware.RequireResourcePermission("tokens")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TokenHandler(w, r, svc)
	})))

	// Users OAuth
	mux.Handle("/api/v1/usersoauth", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserOauthsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/usersoauth/", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserOauthHandler(w, r, svc)
	})))

	// User reassign-delete (composite)
	mux.Handle("POST /api/v1/users/reassign-delete", middleware.RequirePermission("users:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserReassignDeleteHandler(w, r, svc)
	})))

	// Users
	mux.Handle("/api/v1/users", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UsersHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/users/full", middleware.RequirePermission("users:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UsersFullHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/users/full/", middleware.RequirePermission("users:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserFullHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/users/", middleware.RequireResourcePermission("users")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/users/sessions", middleware.RequirePermission("sessions:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		UserSessionHandler(w, r, svc)
	})))

	// SSH Key management endpoints
	mux.Handle("POST /api/v1/ssh-keys", middleware.RequirePermission("ssh_keys:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddSSHKeyHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/ssh-keys", middleware.RequirePermission("ssh_keys:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ListSSHKeysHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/ssh-keys/", middleware.RequirePermission("ssh_keys:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		DeleteSSHKeyHandler(w, r, svc)
	})))

	// Activity feed
	mux.Handle("GET /api/v1/activity/recent", middleware.RequirePermission("audit:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ActivityRecentHandler(w, r, svc)
	})))

	// Import endpoints
	mux.Handle("/api/v1/import/contentful", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportContentfulHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/import/sanity", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportSanityHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/import/strapi", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportStrapiHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/import/wordpress", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportWordPressHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/import/clean", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportCleanHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/import", middleware.RequirePermission("import:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ImportBulkHandler(w, r, svc)
	})))

	// Role-permissions junction table CRUD
	mux.Handle("/api/v1/role-permissions", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolePermissionsHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/role-permissions/", middleware.RequireResourcePermission("roles")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolePermissionHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/role-permissions/role/", middleware.RequirePermission("roles:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RolePermissionsByRoleHandler(w, r, svc)
	})))

	// Deploy sync endpoints
	mux.Handle("GET /api/v1/deploy/health", middleware.RequirePermission("deploy:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deploy.DeployHealthHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/deploy/export", middleware.RequirePermission("deploy:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deploy.DeployExportHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/deploy/import", middleware.RequirePermission("deploy:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deploy.DeployImportHandler(w, r, svc)
	})))

	// Config management endpoints (permission-gated)
	configAuthChain := middleware.AuthenticatedChain(mgr)
	mux.Handle("GET /api/v1/admin/config", configAuthChain(middleware.RequirePermission("config:read")(ConfigGetHandler(svc))))
	mux.Handle("PATCH /api/v1/admin/config", configAuthChain(middleware.RequirePermission("config:update")(ConfigUpdateHandler(svc))))
	mux.Handle("GET /api/v1/admin/config/meta", configAuthChain(middleware.RequirePermission("config:read")(ConfigMetaHandler())))
	mux.Handle("GET /api/v1/admin/config/search-index", configAuthChain(middleware.RequirePermission("config:read")(ConfigSearchIndexHandler())))

	// Metrics endpoint (admin, config:read permission)
	mux.Handle("GET /api/v1/admin/metrics", configAuthChain(middleware.RequirePermission("config:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		snapshot := utility.GlobalMetrics.GetSnapshot()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(snapshot)
	}))))

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

	// Locales — public endpoint (enabled locales only, no auth)
	mux.Handle("GET /api/v1/locales", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LocalesPublicHandler(w, r, svc)
	})))

	// Locales — admin CRUD (requires locale:* permissions)
	mux.Handle("/api/v1/admin/locales", middleware.RequireResourcePermission("locale")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LocalesHandler(w, r, svc)
	})))
	mux.Handle("/api/v1/admin/locales/", middleware.RequireResourcePermission("locale")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LocaleHandler(w, r, svc)
	})))

	// Webhooks — admin CRUD (requires webhook:* permissions)
	mux.Handle("GET /api/v1/admin/webhooks", middleware.RequirePermission("webhook:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookListHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/webhooks", middleware.RequirePermission("webhook:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookCreateHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admin/webhooks/{id}", middleware.RequirePermission("webhook:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookGetHandler(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/admin/webhooks/{id}", middleware.RequirePermission("webhook:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookUpdateHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/admin/webhooks/{id}", middleware.RequirePermission("webhook:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookDeleteHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/webhooks/{id}/test", middleware.RequirePermission("webhook:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookTestHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admin/webhooks/{id}/deliveries", middleware.RequirePermission("webhook:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookDeliveryListHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/webhooks/deliveries/{id}/retry", middleware.RequirePermission("webhook:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WebhookDeliveryRetryHandler(w, r, svc)
	})))

	// Validations — public validation config CRUD
	mux.Handle("GET /api/v1/validations", middleware.RequirePermission("validations:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ValidationListHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/validations", middleware.RequirePermission("validations:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ValidationCreateHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/validations/search", middleware.RequirePermission("validations:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ValidationSearchHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/validations/{id}", middleware.RequirePermission("validations:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ValidationGetHandler(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/validations/{id}", middleware.RequirePermission("validations:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ValidationUpdateHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/validations/{id}", middleware.RequirePermission("validations:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ValidationDeleteHandler(w, r, svc)
	})))

	// Admin Validations — admin validation config CRUD
	mux.Handle("GET /api/v1/admin/validations", middleware.RequirePermission("admin_validations:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminValidationListHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/validations", middleware.RequirePermission("admin_validations:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminValidationCreateHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admin/validations/search", middleware.RequirePermission("admin_validations:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminValidationSearchHandler(w, r, svc)
	})))
	mux.Handle("GET /api/v1/admin/validations/{id}", middleware.RequirePermission("admin_validations:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminValidationGetHandler(w, r, svc)
	})))
	mux.Handle("PUT /api/v1/admin/validations/{id}", middleware.RequirePermission("admin_validations:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminValidationUpdateHandler(w, r, svc)
	})))
	mux.Handle("DELETE /api/v1/admin/validations/{id}", middleware.RequirePermission("admin_validations:delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminValidationDeleteHandler(w, r, svc)
	})))

	// Translations — create locale translations for content data
	mux.Handle("POST /api/v1/admin/contentdata/{id}/translations", middleware.RequirePermission("content:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TranslationHandler(w, r, svc)
	})))
	mux.Handle("POST /api/v1/admin/admincontentdata/{id}/translations", middleware.RequirePermission("content:create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AdminTranslationHandler(w, r, svc)
	})))

	// Content query by datatype (PUBLIC - no auth required)
	mux.Handle("GET /api/v1/query/{datatype}", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		QueryHandler(w, r, svc)
	})))

	// Global content delivery (PUBLIC - no auth required)
	mux.Handle("GET /api/v1/globals", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GlobalsHandler(w, r, svc)
	})))

	// Content delivery via slug
	mux.HandleFunc("/api/v1/content/", func(w http.ResponseWriter, r *http.Request) {
		SlugHandler(w, r, svc)
	})

	// Search endpoint (PUBLIC — no auth, indexes published content only)
	if searchSvc != nil {
		mux.HandleFunc("GET /api/v1/search", func(w http.ResponseWriter, r *http.Request) {
			SearchHandler(w, r, searchSvc)
		})

		// Admin rebuild endpoint
		mux.Handle("POST /api/v1/admin/search/rebuild", middleware.RequirePermission("search:update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SearchRebuildHandler(w, r, searchSvc)
		})))
	}

	// HTMX admin panel
	registerAdminRoutes(mux, mgr, driver, pc, emailSvc, dispatcher, svc)

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

// registerAdminRoutes registers all HTMX-based admin panel routes.
func registerAdminRoutes(mux *http.ServeMux, mgr *config.Manager, driver db.DbDriver, pc *middleware.PermissionCache, emailSvc *email.Service, dispatcher publishing.WebhookDispatcher, svc *service.Registry) {
	// Static assets (no auth, no CSRF)
	staticFS, staticErr := htmxadmin.StaticFS()
	if staticErr == nil {
		mux.Handle("GET /admin/static/", http.StripPrefix("/admin/static/", htmxadmin.CacheControl(http.FileServer(staticFS))))
	}

	// Auth pages (no session auth required)
	loginCSRF := htmxadmin.CSRFMiddleware()
	loginLimiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10) // 10 attempts/min per IP
	mux.Handle("GET /admin/login", loginCSRF(http.HandlerFunc(adminhandlers.LoginPageHandler())))
	mux.Handle("POST /admin/login", loginLimiter.Middleware(loginCSRF(http.HandlerFunc(adminhandlers.LoginSubmitHandler(mgr)))))
	mux.HandleFunc("POST /admin/logout", adminhandlers.LogoutHandler(mgr))
	mux.Handle("GET /admin/forgot-password", loginCSRF(http.HandlerFunc(adminhandlers.ForgotPasswordPageHandler())))
	mux.Handle("POST /admin/forgot-password", loginLimiter.Middleware(loginCSRF(http.HandlerFunc(adminhandlers.ForgotPasswordSubmitHandler(mgr, emailSvc, driver)))))
	mux.Handle("GET /admin/reset-password", loginCSRF(http.HandlerFunc(adminhandlers.ResetPasswordPageHandler(driver))))
	mux.Handle("POST /admin/reset-password", loginLimiter.Middleware(loginCSRF(http.HandlerFunc(adminhandlers.ResetPasswordSubmitHandler(mgr, driver)))))

	adminAuth := htmxadmin.AdminAuthMiddleware(mgr)
	csrf := htmxadmin.CSRFMiddleware()

	// mutating wraps auth + CSRF + permission for POST/DELETE handlers
	mutating := func(permission string, h http.HandlerFunc) http.Handler {
		return adminAuth(csrf(middleware.RequirePermission(permission)(http.HandlerFunc(h))))
	}
	// viewing wraps auth + CSRF + permission for GET handlers.
	// CSRF is needed so the csrf_token cookie and meta tag are refreshed on
	// every page load — without it, POST requests from the page (HTMX, XHR)
	// would lack a valid CSRF cookie after the session cookie expires.
	viewing := func(resource string, h http.HandlerFunc) http.Handler {
		return adminAuth(csrf(middleware.RequirePermission(resource + ":read")(http.HandlerFunc(h))))
	}

	// Dashboard (requires auth but no specific permission)
	mux.Handle("GET /admin/{$}", adminAuth(csrf(http.HandlerFunc(adminhandlers.DashboardHandler(driver)))))

	// Content
	mux.Handle("GET /admin/content", viewing("content", adminhandlers.ContentListHandler(driver, mgr)))
	mux.Handle("GET /admin/content/{id}", viewing("content", adminhandlers.ContentEditHandler(driver, mgr)))
	mux.Handle("POST /admin/content", mutating("content:create", adminhandlers.ContentCreateHandler(driver, mgr)))
	mux.Handle("POST /admin/content/{id}", mutating("content:update", adminhandlers.ContentUpdateHandler(driver, mgr)))
	mux.Handle("DELETE /admin/content/{id}", mutating("content:delete", adminhandlers.ContentDeleteHandler(driver, mgr)))
	mux.Handle("POST /admin/content/reorder", mutating("content:update", adminhandlers.ContentReorderHandler(driver, mgr)))
	mux.Handle("POST /admin/content/move", mutating("content:update", adminhandlers.ContentMoveHandler(driver, mgr)))
	mux.Handle("POST /admin/content/tree", mutating("content:update", adminhandlers.ContentTreeSaveHandler(driver, mgr)))

	// Content publish / unpublish / versions / restore
	mux.Handle("POST /admin/content/{id}/publish", mutating("content:publish", adminhandlers.ContentPublishHandler(driver, mgr, dispatcher)))
	mux.Handle("POST /admin/content/{id}/unpublish", mutating("content:publish", adminhandlers.ContentUnpublishHandler(driver, mgr, dispatcher)))
	mux.Handle("GET /admin/content/{id}/versions", viewing("content", adminhandlers.ContentVersionsHandler(driver)))
	mux.Handle("POST /admin/content/{id}/versions", mutating("content:update", adminhandlers.ContentCreateVersionHandler(driver, mgr)))
	mux.Handle("POST /admin/content/{id}/restore", mutating("content:update", adminhandlers.ContentRestoreVersionHandler(driver, mgr)))
	mux.Handle("GET /admin/content/{id}/versions/compare", viewing("content", adminhandlers.ContentVersionCompareHandler(driver)))

	// Admin API — config endpoints
	mux.Handle("GET /admin/api/config/richtext-toolbar", viewing("config", adminhandlers.RichtextToolbarHandler(mgr)))

	// datatypes (JSON API for block editor)
	mux.Handle("GET /admin/api/datatypes", viewing("datatypes", adminhandlers.DatatypesJSONHandler(svc)))
	mux.Handle("GET /admin/api/datatypes/{id}/fields", viewing("datatypes", adminhandlers.DatatypeFieldsJSONHandler(svc)))

	// datatypes
	mux.Handle("GET /admin/datatypes", viewing("datatypes", adminhandlers.DatatypesListHandler(svc)))
	mux.Handle("GET /admin/datatypes/{id}", viewing("datatypes", adminhandlers.DatatypeDetailHandler(svc)))
	mux.Handle("POST /admin/datatypes", mutating("datatypes:create", adminhandlers.DatatypeCreateHandler(svc)))
	mux.Handle("POST /admin/datatypes/{id}", mutating("datatypes:update", adminhandlers.DatatypeUpdateHandler(svc)))
	mux.Handle("DELETE /admin/datatypes/{id}", mutating("datatypes:delete", adminhandlers.DatatypeDeleteHandler(svc)))

	// fields
	mux.Handle("GET /admin/fields", viewing("fields", adminhandlers.FieldsListHandler(svc)))
	mux.Handle("GET /admin/fields/new", viewing("fields", adminhandlers.FieldCreatePageHandler(svc)))
	mux.Handle("POST /admin/fields", mutating("fields:create", adminhandlers.FieldCreateHandler(svc)))
	mux.Handle("GET /admin/fields/{id}", viewing("fields", adminhandlers.FieldDetailHandler(svc)))
	mux.Handle("POST /admin/fields/{id}", mutating("fields:update", adminhandlers.FieldUpdateHandler(svc)))
	mux.Handle("DELETE /admin/fields/{id}", mutating("fields:delete", adminhandlers.FieldDeleteHandler(svc)))

	// field types
	mux.Handle("GET /admin/field-types", viewing("field_types", adminhandlers.FieldTypesListHandler(svc)))
	mux.Handle("GET /admin/field-types/{id}", viewing("field_types", adminhandlers.FieldTypeDetailHandler(svc)))
	mux.Handle("POST /admin/field-types", mutating("field_types:create", adminhandlers.FieldTypeCreateHandler(svc)))
	mux.Handle("POST /admin/field-types/{id}", mutating("field_types:update", adminhandlers.FieldTypeUpdateHandler(svc)))
	mux.Handle("DELETE /admin/field-types/{id}", mutating("field_types:delete", adminhandlers.FieldTypeDeleteHandler(svc)))

	// validations
	mux.Handle("GET /admin/validations", viewing("validations", adminhandlers.ValidationsListHandler(svc)))
	mux.Handle("GET /admin/validations/{id}", viewing("validations", adminhandlers.ValidationDetailHandler(svc)))
	mux.Handle("POST /admin/validations", mutating("validations:create", adminhandlers.ValidationCreateHandler(svc)))
	mux.Handle("POST /admin/validations/{id}", mutating("validations:update", adminhandlers.ValidationUpdateHandler(svc)))
	mux.Handle("DELETE /admin/validations/{id}", mutating("validations:delete", adminhandlers.ValidationDeleteHandler(svc)))

	// datatype field creation
	mux.Handle("POST /admin/datatypes/{id}/fields", mutating("fields:create", adminhandlers.DatatypeCreateFieldHandler(svc)))

	// Media
	mux.Handle("GET /admin/media", viewing("media", adminhandlers.MediaListHandler(svc)))
	mux.Handle("GET /admin/media/{id}", viewing("media", adminhandlers.MediaDetailHandler(svc)))
	mux.Handle("POST /admin/media", mutating("media:create", adminhandlers.MediaUploadHandler(svc)))
	mux.Handle("POST /admin/media/{id}", mutating("media:update", adminhandlers.MediaUpdateHandler(svc)))
	mux.Handle("DELETE /admin/media/{id}", mutating("media:delete", adminhandlers.MediaDeleteHandler(svc)))
	mux.Handle("POST /admin/media/bulk-delete", mutating("media:delete", adminhandlers.MediaBulkDeleteHandler(svc)))

	// Media folders
	mux.Handle("POST /admin/media-folders", mutating("media:create", adminhandlers.MediaFolderCreateHandler(svc)))
	mux.Handle("POST /admin/media-folders/{id}", mutating("media:update", adminhandlers.MediaFolderUpdateHandler(svc)))
	mux.Handle("DELETE /admin/media-folders/{id}", mutating("media:delete", adminhandlers.MediaFolderDeleteHandler(svc)))
	mux.Handle("POST /admin/media/move/{id}", mutating("media:update", adminhandlers.MediaMoveToFolderHandler(svc)))

	// Admin Media
	mux.Handle("GET /admin/admin-media", viewing("media", adminhandlers.AdminMediaListHandler(svc)))
	mux.Handle("GET /admin/admin-media/{id}", viewing("media", adminhandlers.AdminMediaDetailHandler(svc)))
	mux.Handle("POST /admin/admin-media", mutating("media:create", adminhandlers.AdminMediaUploadHandler(svc)))
	mux.Handle("POST /admin/admin-media/{id}", mutating("media:update", adminhandlers.AdminMediaUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-media/{id}", mutating("media:delete", adminhandlers.AdminMediaDeleteHandler(svc)))
	mux.Handle("POST /admin/admin-media/bulk-delete", mutating("media:delete", adminhandlers.AdminMediaBulkDeleteHandler(svc)))

	// Admin Media folders
	mux.Handle("POST /admin/admin-media-folders", mutating("media:create", adminhandlers.AdminMediaFolderCreateHandler(svc)))
	mux.Handle("POST /admin/admin-media-folders/{id}", mutating("media:update", adminhandlers.AdminMediaFolderUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-media-folders/{id}", mutating("media:delete", adminhandlers.AdminMediaFolderDeleteHandler(svc)))
	mux.Handle("POST /admin/admin-media/move/{id}", mutating("media:update", adminhandlers.AdminMediaMoveToFolderHandler(svc)))

	// Media dimensions
	mux.Handle("GET /admin/media/dimensions", viewing("media", adminhandlers.MediaDimensionsListHandler(svc)))
	mux.Handle("POST /admin/media/dimensions", mutating("media:create", adminhandlers.MediaDimensionCreateHandler(svc)))
	mux.Handle("POST /admin/media/dimensions/{id}", mutating("media:update", adminhandlers.MediaDimensionUpdateHandler(svc)))
	mux.Handle("DELETE /admin/media/dimensions/{id}", mutating("media:delete", adminhandlers.MediaDimensionDeleteHandler(svc)))

	// Routes
	mux.Handle("GET /admin/routes", viewing("routes", adminhandlers.RoutesListHandler(svc)))
	mux.Handle("GET /admin/routes/admin", viewing("routes", adminhandlers.AdminRoutesListHandler(svc)))
	mux.Handle("GET /admin/routes/{id}", viewing("routes", adminhandlers.RouteDetailHandler(svc)))
	mux.Handle("POST /admin/routes", mutating("routes:create", adminhandlers.RouteCreateHandler(svc)))
	mux.Handle("POST /admin/routes/{id}", mutating("routes:update", adminhandlers.RouteUpdateHandler(svc)))
	mux.Handle("DELETE /admin/routes/{id}", mutating("routes:delete", adminhandlers.RouteDeleteHandler(svc)))

	// Admin Datatypes
	mux.Handle("GET /admin/admin-datatypes", viewing("datatypes", adminhandlers.AdminDatatypesListHandler(svc)))
	mux.Handle("GET /admin/admin-datatypes/{id}", viewing("datatypes", adminhandlers.AdminDatatypeDetailHandler(svc)))
	mux.Handle("POST /admin/admin-datatypes", mutating("datatypes:create", adminhandlers.AdminDatatypeCreateHandler(svc)))
	mux.Handle("POST /admin/admin-datatypes/{id}", mutating("datatypes:update", adminhandlers.AdminDatatypeUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-datatypes/{id}", mutating("datatypes:delete", adminhandlers.AdminDatatypeDeleteHandler(svc)))
	mux.Handle("POST /admin/admin-datatypes/{id}/fields", mutating("fields:create", adminhandlers.AdminDatatypeCreateFieldHandler(svc)))

	// Admin Fields (detail/update/delete only — fields are created per admin datatype)
	mux.Handle("GET /admin/admin-fields/{id}", viewing("fields", adminhandlers.AdminFieldDetailHandler(svc)))
	mux.Handle("POST /admin/admin-fields/{id}", mutating("fields:update", adminhandlers.AdminFieldUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-fields/{id}", mutating("fields:delete", adminhandlers.AdminFieldDeleteHandler(svc)))

	// Admin Field Types
	mux.Handle("GET /admin/admin-field-types", viewing("field_types", adminhandlers.AdminFieldTypesListHandler(svc)))
	mux.Handle("GET /admin/admin-field-types/{id}", viewing("field_types", adminhandlers.AdminFieldTypeDetailHandler(svc)))
	mux.Handle("POST /admin/admin-field-types", mutating("field_types:create", adminhandlers.AdminFieldTypeCreateHandler(svc)))
	mux.Handle("POST /admin/admin-field-types/{id}", mutating("field_types:update", adminhandlers.AdminFieldTypeUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-field-types/{id}", mutating("field_types:delete", adminhandlers.AdminFieldTypeDeleteHandler(svc)))

	// Admin Validations
	mux.Handle("GET /admin/admin-validations", viewing("admin_validations", adminhandlers.AdminValidationsListHandler(svc)))
	mux.Handle("GET /admin/admin-validations/{id}", viewing("admin_validations", adminhandlers.AdminValidationDetailHandler(svc)))
	mux.Handle("POST /admin/admin-validations", mutating("admin_validations:create", adminhandlers.AdminValidationCreateHandler(svc)))
	mux.Handle("POST /admin/admin-validations/{id}", mutating("admin_validations:update", adminhandlers.AdminValidationUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-validations/{id}", mutating("admin_validations:delete", adminhandlers.AdminValidationDeleteHandler(svc)))

	// Admin Routes
	mux.Handle("GET /admin/admin-routes", viewing("routes", adminhandlers.AdminSchemaRoutesListHandler(svc)))
	mux.Handle("POST /admin/admin-routes", mutating("routes:create", adminhandlers.AdminRouteCreateHandler(svc)))
	mux.Handle("POST /admin/admin-routes/{id}", mutating("routes:update", adminhandlers.AdminRouteUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-routes/{id}", mutating("routes:delete", adminhandlers.AdminRouteDeleteHandler(svc)))

	// Admin Content
	mux.Handle("GET /admin/admin-content", viewing("content", adminhandlers.AdminContentListHandler(svc)))
	mux.Handle("GET /admin/admin-content/{id}", viewing("content", adminhandlers.AdminContentEditHandler(svc)))
	mux.Handle("POST /admin/admin-content", mutating("content:create", adminhandlers.AdminContentCreateHandler(svc)))
	mux.Handle("POST /admin/admin-content/{id}", mutating("content:update", adminhandlers.AdminContentUpdateHandler(svc)))
	mux.Handle("DELETE /admin/admin-content/{id}", mutating("content:delete", adminhandlers.AdminContentDeleteHandler(svc)))
	mux.Handle("POST /admin/admin-content/{id}/publish", mutating("content:publish", adminhandlers.AdminContentPublishHandler(svc)))
	mux.Handle("POST /admin/admin-content/{id}/unpublish", mutating("content:publish", adminhandlers.AdminContentUnpublishHandler(svc)))
	mux.Handle("GET /admin/admin-content/{id}/versions", viewing("content", adminhandlers.AdminContentVersionsHandler(svc)))

	// Users
	mux.Handle("GET /admin/users", viewing("users", adminhandlers.UsersListHandler(svc)))
	mux.Handle("GET /admin/users/{id}", viewing("users", adminhandlers.UserDetailHandler(svc)))
	mux.Handle("POST /admin/users", mutating("users:create", adminhandlers.UserCreateHandler(svc)))
	mux.Handle("POST /admin/users/{id}", mutating("users:update", adminhandlers.UserUpdateHandler(svc)))
	mux.Handle("DELETE /admin/users/{id}", mutating("users:delete", adminhandlers.UserDeleteHandler(svc)))

	// User SSH Keys
	mux.Handle("POST /admin/ssh-keys/{id}", mutating("ssh_keys:create", adminhandlers.UserSSHKeyAddHandler(svc)))
	mux.Handle("DELETE /admin/ssh-keys/{id}/{keyId}", mutating("ssh_keys:delete", adminhandlers.UserSSHKeyDeleteHandler(svc)))

	// User OAuth Connections
	mux.Handle("POST /admin/oauth/{id}", mutating("users:update", adminhandlers.OAuthCreateHandler(svc)))
	mux.Handle("DELETE /admin/oauth/{id}", mutating("users:update", adminhandlers.OAuthDeleteHandler(svc)))

	// Roles
	mux.Handle("GET /admin/users/roles", viewing("roles", adminhandlers.RolesListHandler(svc)))
	mux.Handle("GET /admin/users/roles/new", viewing("roles", adminhandlers.RoleNewFormHandler(svc)))
	mux.Handle("GET /admin/users/roles/{id}", viewing("roles", adminhandlers.RoleDetailHandler(svc)))
	mux.Handle("POST /admin/users/roles", mutating("roles:create", adminhandlers.RoleCreateHandler(svc)))
	mux.Handle("POST /admin/users/roles/{id}", mutating("roles:update", adminhandlers.RoleUpdateHandler(svc)))
	mux.Handle("DELETE /admin/users/roles/{id}", mutating("roles:delete", adminhandlers.RoleDeleteHandler(svc)))

	// Tokens
	mux.Handle("GET /admin/users/tokens", viewing("tokens", adminhandlers.TokensListHandler(svc)))
	mux.Handle("GET /admin/users/tokens/{id}", viewing("tokens", adminhandlers.TokenDetailHandler(svc)))
	mux.Handle("POST /admin/users/tokens", mutating("tokens:create", adminhandlers.TokenCreateHandler(svc)))
	mux.Handle("DELETE /admin/users/tokens/{id}", mutating("tokens:delete", adminhandlers.TokenDeleteHandler(svc)))

	// Sessions
	mux.Handle("GET /admin/sessions", viewing("sessions", adminhandlers.SessionsListHandler(svc)))
	mux.Handle("GET /admin/sessions/{id}", viewing("sessions", adminhandlers.SessionDetailHandler(svc)))
	mux.Handle("DELETE /admin/sessions/{id}", mutating("sessions:delete", adminhandlers.SessionDeleteHandler(svc)))

	// Plugins
	mux.Handle("GET /admin/plugins", viewing("plugins", adminhandlers.PluginsListHandler(svc)))
	mux.Handle("GET /admin/plugins/{name}", viewing("plugins", adminhandlers.PluginDetailHandler(svc)))
	mux.Handle("POST /admin/plugins/{name}/enable", mutating("plugins:admin", adminhandlers.PluginEnableHandler(svc)))
	mux.Handle("POST /admin/plugins/{name}/disable", mutating("plugins:admin", adminhandlers.PluginDisableHandler(svc)))
	mux.Handle("POST /admin/plugins/{name}/reload", mutating("plugins:admin", adminhandlers.PluginReloadHandler(svc)))
	mux.Handle("POST /admin/plugins/routes/approve", mutating("plugins:admin", adminhandlers.PluginApproveRouteHandler(svc)))
	mux.Handle("POST /admin/plugins/routes/revoke", mutating("plugins:admin", adminhandlers.PluginRevokeRouteHandler(svc)))
	mux.Handle("POST /admin/plugins/routes/approve-all", mutating("plugins:admin", adminhandlers.PluginApproveAllRoutesHandler(svc)))
	mux.Handle("POST /admin/plugins/routes/revoke-all", mutating("plugins:admin", adminhandlers.PluginRevokeAllRoutesHandler(svc)))
	mux.Handle("POST /admin/plugins/hooks/approve", mutating("plugins:admin", adminhandlers.PluginApproveHookHandler(svc)))
	mux.Handle("POST /admin/plugins/hooks/revoke", mutating("plugins:admin", adminhandlers.PluginRevokeHookHandler(svc)))
	mux.Handle("POST /admin/plugins/hooks/approve-all", mutating("plugins:admin", adminhandlers.PluginApproveAllHooksHandler(svc)))
	mux.Handle("POST /admin/plugins/hooks/revoke-all", mutating("plugins:admin", adminhandlers.PluginRevokeAllHooksHandler(svc)))

	// Pipelines (read-only viewer, plugins:read permission)
	mux.Handle("GET /admin/pipelines", viewing("plugins", adminhandlers.PipelinesListHandler(svc)))
	mux.Handle("GET /admin/pipelines/{key}", viewing("plugins", adminhandlers.PipelineDetailHandler(svc)))

	// Tables (system operation — mutations require password re-authentication in handler)
	mux.Handle("GET /admin/tables", viewing("tables", adminhandlers.TablesListHandler(svc)))
	mux.Handle("POST /admin/tables", mutating("tables:create", adminhandlers.TableCreateHandler(svc)))
	mux.Handle("POST /admin/tables/update", mutating("tables:update", adminhandlers.TableUpdateHandler(svc)))
	mux.Handle("POST /admin/tables/delete", mutating("tables:delete", adminhandlers.TableDeleteHandler(svc)))

	// Deploy
	mux.Handle("GET /admin/deploy", viewing("deploy", adminhandlers.DeployPageHandler(svc)))
	mux.Handle("POST /admin/deploy/{name}/health", mutating("deploy:read", adminhandlers.DeployHealthHandler(svc)))
	mux.Handle("POST /admin/deploy/{name}/push", mutating("deploy:create", adminhandlers.DeployPushHandler(svc)))
	mux.Handle("POST /admin/deploy/{name}/pull", mutating("deploy:create", adminhandlers.DeployPullHandler(svc)))

	// Import
	mux.Handle("GET /admin/import", viewing("import", adminhandlers.ImportPageHandler()))
	mux.Handle("POST /admin/import", mutating("import:create", adminhandlers.ImportSubmitHandler(svc)))

	// Demo
	mux.Handle("GET /admin/demo", viewing("settings", adminhandlers.DemoHandler()))

	// Search
	mux.Handle("GET /admin/search", adminAuth(csrf(http.HandlerFunc(adminhandlers.AdminSearchHandler(svc)))))

	// Audit
	mux.Handle("GET /admin/audit", adminAuth(csrf(http.HandlerFunc(adminhandlers.AuditLogHandler(svc)))))
	mux.Handle("GET /admin/audit/{eventID}", adminAuth(csrf(http.HandlerFunc(adminhandlers.AuditDetailHandler(svc)))))

	// Settings
	mux.Handle("GET /admin/settings", viewing("config", adminhandlers.SettingsHandler(svc)))
	mux.Handle("POST /admin/settings", mutating("config:update", adminhandlers.SettingsUpdateHandler(svc)))
	mux.Handle("POST /admin/settings/search/rebuild", mutating("search:update", adminhandlers.SearchRebuildHandler(svc)))

	// Locale settings (i18n)
	mux.Handle("GET /admin/settings/locales", viewing("locale", adminhandlers.LocaleSettingsHandler(svc)))
	mux.Handle("GET /admin/settings/locales/{id}/edit", viewing("locale", adminhandlers.LocaleEditDialogHandler(svc)))
	mux.Handle("POST /admin/settings/locales", mutating("locale:create", adminhandlers.LocaleCreateHandler(svc)))
	mux.Handle("PUT /admin/settings/locales/{id}", mutating("locale:update", adminhandlers.LocaleUpdateHandler(svc)))
	mux.Handle("DELETE /admin/settings/locales/{id}", mutating("locale:delete", adminhandlers.LocaleDeleteHandler(svc)))

	// Webhook settings
	mux.Handle("GET /admin/settings/webhooks", viewing("webhook", adminhandlers.WebhookSettingsHandler(svc)))
	mux.Handle("GET /admin/settings/webhooks/{id}", viewing("webhook", adminhandlers.WebhookDetailHandler(svc)))
	mux.Handle("POST /admin/settings/webhooks", mutating("webhook:create", adminhandlers.WebhookCreateHandler(svc)))
	mux.Handle("POST /admin/settings/webhooks/{id}", mutating("webhook:update", adminhandlers.WebhookUpdateHandler(svc)))
	mux.Handle("DELETE /admin/settings/webhooks/{id}", mutating("webhook:delete", adminhandlers.WebhookDeleteHandler(svc)))
	mux.Handle("POST /admin/settings/webhooks/{id}/test", mutating("webhook:update", adminhandlers.WebhookTestHandler(svc)))

	// Backups
	mux.Handle("GET /admin/settings/backups", viewing("backup", adminhandlers.BackupsListHandler(svc)))
	mux.Handle("GET /admin/settings/backups/{id}", viewing("backup", adminhandlers.BackupDetailHandler(svc)))
	mux.Handle("POST /admin/settings/backups", mutating("backup:create", adminhandlers.BackupCreateHandler(svc)))
	mux.Handle("POST /admin/settings/backups/{id}/restore", mutating("backup:update", adminhandlers.BackupRestoreHandler(svc)))
	mux.Handle("DELETE /admin/settings/backups/{id}", mutating("backup:delete", adminhandlers.BackupDeleteHandler(svc)))
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
