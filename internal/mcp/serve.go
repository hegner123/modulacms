package mcp

import (
	"net/http"

	"github.com/mark3labs/mcp-go/server"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// newServer creates an MCPServer with all tools registered.
// If cm is non-nil, connection management tools (list_projects,
// switch_project, get_connection) are also registered.
// Variadic opts are applied to the MCPServer after creation.
func newServer(backends *Backends, cm *ConnectionManager, opts ...server.ServerOption) *server.MCPServer {
	srv := server.NewMCPServer("modula", utility.Version, opts...)

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
			utility.DefaultLogger.Warn("auto-connect from cwd failed", cwdErr)
		}
	}

	backends := NewProxyBackends(cm)
	return server.ServeStdio(newServer(backends, cm))
}

// ServeDirect creates an MCP server that calls services directly and serves
// over stdio. Used for local/embedded scenarios where the MCP is a trusted
// local pipe. Permission enforcement is not applied (no HTTP auth context).
// The provided AuditContext is injected into tool calls via
// injectAuditContextMiddleware so that AuditContextFromMCP has a fallback
// identity for audit trails.
func ServeDirect(svc *service.Registry, ac audited.AuditContext) error {
	backends := NewServiceBackends(svc)
	srv := newServer(backends, nil,
		server.WithToolHandlerMiddleware(injectAuditContextMiddleware(ac)),
	)
	return server.ServeStdio(srv)
}

// DirectHandler returns an http.Handler that serves the MCP protocol over
// Streamable HTTP, calling services directly without HTTP round-trips.
// Authentication is handled by the DefaultMiddlewareChain which wraps the
// mux and populates the request context. The PermissionMiddleware checks
// per-tool permissions before executing each tool call.
func DirectHandler(svc *service.Registry) http.Handler {
	backends := NewServiceBackends(svc)
	srv := newServer(backends, nil,
		server.WithToolHandlerMiddleware(PermissionMiddleware()),
	)
	return server.NewStreamableHTTPServer(srv,
		server.WithEndpointPath("/mcp"),
		server.WithHTTPContextFunc(PassthroughHTTPContextFunc()),
	)
}
