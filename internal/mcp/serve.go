package mcp

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/server"

	"github.com/hegner123/modulacms/internal/utility"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// newServer creates an MCPServer with all tools registered.
func newServer(client *modula.Client) *server.MCPServer {
	srv := server.NewMCPServer("modula", utility.Version)

	registerContentTools(srv, client)
	registerSchemaTools(srv, client)
	registerMediaTools(srv, client)
	registerRouteTools(srv, client)
	registerUserTools(srv, client)
	registerRBACTools(srv, client)
	registerConfigTools(srv, client)
	registerImportTools(srv, client)
	registerDeployTools(srv, client)
	registerHealthTools(srv, client)
	registerSessionTools(srv, client)
	registerTokenTools(srv, client)
	registerSSHKeyTools(srv, client)
	registerOAuthTools(srv, client)
	registerTableTools(srv, client)
	registerPluginTools(srv, client)
	registerAdminContentTools(srv, client)
	registerAdminSchemaTools(srv, client)
	registerAdminRouteTools(srv, client)

	return srv
}

// Serve creates an MCP server connected to a Modula instance and serves over stdio.
// This mode makes real HTTP calls to the CMS at the given URL.
func Serve(url, apiKey string) error {
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL: url,
		APIKey:  apiKey,
	})
	if err != nil {
		return fmt.Errorf("create SDK client: %w", err)
	}

	return server.ServeStdio(newServer(client))
}

// DirectHandler returns an http.Handler that serves the MCP protocol over
// Streamable HTTP, routing SDK calls directly to the given in-process handler
// instead of making HTTP requests over the network.
//
// The handler should be the full middleware-wrapped ServeMux so that SDK calls
// go through auth, permissions, and audit middleware normally.
func DirectHandler(handler http.Handler, apiKey string) (http.Handler, error) {
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL: "http://internal",
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Transport: &directTransport{handler: handler},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create direct SDK client: %w", err)
	}

	httpServer := server.NewStreamableHTTPServer(
		newServer(client),
		server.WithEndpointPath("/mcp"),
	)

	return httpServer, nil
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
