package mcp

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// newServer creates an MCPServer with all tools registered.
// If cm is non-nil, connection management tools (list_projects,
// switch_project, get_connection) are also registered.
func newServer(backends *Backends, cm *ConnectionManager) *server.MCPServer {
	srv := server.NewMCPServer("modula", utility.Version)

	if cm != nil {
		registerConnectionTools(srv, cm)
	}

	registerContentTools(srv, backends.Content)
	registerSchemaTools(srv, backends.Schema)
	registerMediaTools(srv, backends.Media, backends.MediaFolders)
	registerMediaFolderTools(srv, backends.MediaFolders)
	registerRouteTools(srv, backends.Routes)
	registerUserTools(srv, backends.Users)
	registerRBACTools(srv, backends.RBAC)
	registerConfigTools(srv, backends.Config)
	registerImportTools(srv, backends.Import)
	registerDeployTools(srv, backends.Deploy)
	registerHealthTools(srv, backends.Health)
	registerSessionTools(srv, backends.Sessions)
	registerTokenTools(srv, backends.Tokens)
	registerSSHKeyTools(srv, backends.SSHKeys)
	registerOAuthTools(srv, backends.OAuth)
	registerTableTools(srv, backends.Tables)
	registerPluginTools(srv, backends.Plugins)
	registerAdminContentTools(srv, backends.AdminContent)
	registerAdminSchemaTools(srv, backends.AdminSchema)
	registerAdminRouteTools(srv, backends.AdminRoutes)
	registerAdminMediaTools(srv, backends.AdminMedia, backends.AdminMediaFolders)
	registerAdminMediaFolderTools(srv, backends.AdminMediaFolders)
	registerPublishingTools(srv, backends.Publishing)
	registerVersionTools(srv, backends.Versions)
	registerWebhookTools(srv, backends.Webhooks)
	registerLocaleTools(srv, backends.Locales)
	registerValidationTools(srv, backends.Validations)
	registerSearchTools(srv, backends.Search)
	registerActivityTools(srv, backends.Activity)
	registerAuthTools(srv, backends.Auth)

	return srv
}

// ServeWithRegistry creates an MCP server backed by the project registry.
// It loads ~/.modula/configs.json and exposes tools for listing projects and
// switching the active connection at runtime. If initialProject is non-empty,
// the server connects to that project immediately; otherwise it attempts to
// auto-detect a project from the working directory.
func ServeWithRegistry(initialProject, initialEnv string) error {
	cm, err := NewConnectionManager()
	if err != nil {
		return err
	}

	// Try to connect: explicit project > cwd detection > stay disconnected.
	if initialProject != "" {
		if connErr := cm.SwitchProject(initialProject, initialEnv); connErr != nil {
			return connErr
		}
	} else {
		// Best-effort: connect to whatever project matches the cwd.
		if _, cwdErr := cm.ConnectFromCwd(); cwdErr != nil {
			fmt.Fprintf(os.Stderr, "warning: auto-connect from cwd failed: %v\n", cwdErr)
		}
	}

	backends := NewProxyBackends(cm)
	return server.ServeStdio(newServer(backends, cm))
}

// ServeDirect creates an MCP server that calls services directly and serves over stdio.
func ServeDirect(svc *service.Registry, ac audited.AuditContext) error {
	return server.ServeStdio(newServer(NewServiceBackends(svc, ac), nil))
}

// DirectHandler returns an http.Handler that serves the MCP protocol over
// Streamable HTTP, calling services directly without HTTP round-trips.
func DirectHandler(svc *service.Registry, ac audited.AuditContext) http.Handler {
	return server.NewStreamableHTTPServer(
		newServer(NewServiceBackends(svc, ac), nil),
		server.WithEndpointPath("/mcp"),
	)
}

// APIKeyAuth wraps an http.Handler with Bearer token authentication for the
// MCP endpoint. It extracts the token from the Authorization header and
// performs a constant-time comparison against the configured API key.
func APIKeyAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := auth[len("Bearer "):]
		if subtle.ConstantTimeCompare([]byte(token), []byte(apiKey)) != 1 {
			http.Error(w, "invalid API key", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
