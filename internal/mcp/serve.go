package mcp

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// newServer creates an MCPServer with all tools registered.
func newServer(backends *Backends) *server.MCPServer {
	srv := server.NewMCPServer("modula", utility.Version)

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

	return srv
}

// Serve creates an MCP server connected to a remote Modula instance and serves over stdio.
// This mode makes real HTTP calls to the CMS at the given URL.
func Serve(url, apiKey string) error {
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL: url,
		APIKey:  apiKey,
	})
	if err != nil {
		return fmt.Errorf("create SDK client: %w", err)
	}

	return server.ServeStdio(newServer(NewSDKBackends(client)))
}

// ServeDirect creates an MCP server that calls services directly and serves over stdio.
func ServeDirect(svc *service.Registry, ac audited.AuditContext) error {
	return server.ServeStdio(newServer(NewServiceBackends(svc, ac)))
}

// DirectHandler returns an http.Handler that serves the MCP protocol over
// Streamable HTTP, calling services directly without HTTP round-trips.
func DirectHandler(svc *service.Registry, ac audited.AuditContext) http.Handler {
	return server.NewStreamableHTTPServer(
		newServer(NewServiceBackends(svc, ac)),
		server.WithEndpointPath("/mcp"),
	)
}

// RemoteHandler returns an http.Handler that serves the MCP protocol over
// Streamable HTTP, proxying calls to a remote CMS via the Go SDK.
func RemoteHandler(url, apiKey string) (http.Handler, error) {
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL: url,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create SDK client: %w", err)
	}

	return server.NewStreamableHTTPServer(
		newServer(NewSDKBackends(client)),
		server.WithEndpointPath("/mcp"),
	), nil
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
