package router

import (
	"net/http"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/middleware"
	"golang.org/x/time/rate"
)

func NewModulacmsMux(c config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	// Create rate limiter for auth endpoints: 10 requests per minute per IP
	authLimiter := middleware.NewRateLimiter(rate.Limit(10.0/60.0), 10)

	// Create CORS middleware
	corsMiddleware := middleware.CorsMiddleware(&c)

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	/*
		mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, c)
		})
	*/
	// Auth endpoints with CORS and rate limiting
	mux.Handle("POST /api/v1/auth/login", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, c)
	}))))
	mux.Handle("POST /api/v1/auth/logout", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, c)
	}))))
	mux.Handle("GET /api/v1/auth/me", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MeHandler(w, r, c)
	}))))
	mux.Handle("POST /api/v1/auth/register", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, c)
	}))))
	mux.Handle("POST /api/v1/auth/reset", corsMiddleware(authLimiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ResetPasswordHandler(w, r, c)
	}))))

	// OAuth endpoints with CORS and rate limiting (PUBLIC - no auth required)
	mux.Handle("GET /api/v1/auth/oauth/login", corsMiddleware(authLimiter.Middleware(OauthInitiateHandler(c))))
	mux.Handle("GET /api/v1/auth/oauth/callback", corsMiddleware(authLimiter.Middleware(OauthCallbackHandler(c))))
	mux.HandleFunc("/api/v1/admin/tree/", func(w http.ResponseWriter, r *http.Request) {
		AdminTreeHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/admincontentdatas", func(w http.ResponseWriter, r *http.Request) {
		AdminContentDatasHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/admincontentdatas/", func(w http.ResponseWriter, r *http.Request) {
		AdminContentDataHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/admincontentfields", func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/admincontentfields/", func(w http.ResponseWriter, r *http.Request) {
		AdminContentFieldHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/admindatatypes", func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypesHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/admindatatypes/", func(w http.ResponseWriter, r *http.Request) {
		AdminDatatypeHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/adminfields", func(w http.ResponseWriter, r *http.Request) {
		AdminFieldsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/adminfields/", func(w http.ResponseWriter, r *http.Request) {
		AdminFieldHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/adminroutes", func(w http.ResponseWriter, r *http.Request) {
		AdminRoutesHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/adminroutes/", func(w http.ResponseWriter, r *http.Request) {
		AdminRouteHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/contentdata", func(w http.ResponseWriter, r *http.Request) {
		ContentDatasHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/contentdata/", func(w http.ResponseWriter, r *http.Request) {
		ContentDataHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/contentfields", func(w http.ResponseWriter, r *http.Request) {
		ContentFieldsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/contentfields/", func(w http.ResponseWriter, r *http.Request) {
		ContentFieldHandler(w, r, c)
	})
	mux.HandleFunc("POST /api/v1/content/batch", func(w http.ResponseWriter, r *http.Request) {
		ContentBatchHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/datatype", func(w http.ResponseWriter, r *http.Request) {
		DatatypesHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/datatype/", func(w http.ResponseWriter, r *http.Request) {
		DatatypeHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/datatypefields", func(w http.ResponseWriter, r *http.Request) {
		DatatypeFieldsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/datatypefields/", func(w http.ResponseWriter, r *http.Request) {
		DatatypeFieldHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/fields", func(w http.ResponseWriter, r *http.Request) {
		FieldsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/fields/", func(w http.ResponseWriter, r *http.Request) {
		FieldHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/media", func(w http.ResponseWriter, r *http.Request) {
		MediasHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		MediaHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/mediadimensions", func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/mediadimensions/", func(w http.ResponseWriter, r *http.Request) {
		MediaDimensionHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/mediaupload/", func(w http.ResponseWriter, r *http.Request) {
		MediaUploadHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/routes", func(w http.ResponseWriter, r *http.Request) {
		RoutesHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/routes/", func(w http.ResponseWriter, r *http.Request) {
		RouteHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
		RolesHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/roles/", func(w http.ResponseWriter, r *http.Request) {
		RoleHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		SessionsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/sessions/", func(w http.ResponseWriter, r *http.Request) {
		SessionHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/tables", func(w http.ResponseWriter, r *http.Request) {
		TablesHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/tables/", func(w http.ResponseWriter, r *http.Request) {
		TableHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		TokensHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/tokens/", func(w http.ResponseWriter, r *http.Request) {
		TokenHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/usersoauth", func(w http.ResponseWriter, r *http.Request) {
		UserOauthsHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/usersoauth/", func(w http.ResponseWriter, r *http.Request) {
		UserOauthHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		UsersHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		UserHandler(w, r, c)
	})

	// SSH Key management endpoints (require authentication)
	mux.HandleFunc("POST /api/v1/ssh-keys", func(w http.ResponseWriter, r *http.Request) {
		AddSSHKeyHandler(w, r, c)
	})
	mux.HandleFunc("GET /api/v1/ssh-keys", func(w http.ResponseWriter, r *http.Request) {
		ListSSHKeysHandler(w, r, c)
	})
	mux.HandleFunc("DELETE /api/v1/ssh-keys/", func(w http.ResponseWriter, r *http.Request) {
		DeleteSSHKeyHandler(w, r, c)
	})

	// Import endpoints - accept CMS format and transform to ModulaCMS
	mux.HandleFunc("/api/v1/import/contentful", func(w http.ResponseWriter, r *http.Request) {
		ImportContentfulHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/import/sanity", func(w http.ResponseWriter, r *http.Request) {
		ImportSanityHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/import/strapi", func(w http.ResponseWriter, r *http.Request) {
		ImportStrapiHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/import/wordpress", func(w http.ResponseWriter, r *http.Request) {
		ImportWordPressHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/import/clean", func(w http.ResponseWriter, r *http.Request) {
		ImportCleanHandler(w, r, c)
	})
	mux.HandleFunc("/api/v1/import", func(w http.ResponseWriter, r *http.Request) {
		ImportBulkHandler(w, r, c)
	})

	//mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	//mux.HandleFunc("/auth/google/callback", oauthGoogleCallback)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		SlugHandler(w, r, c)
	})
	return mux

}
