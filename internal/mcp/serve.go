package mcp

import (
	"fmt"
	"net/http"

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

// Handler returns an http.Handler that serves the MCP protocol over Streamable HTTP.
// The baseURL should be the public URL of the server (e.g. "http://localhost:8080").
func Handler(baseURL string, apiKey string) (http.Handler, error) {
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("create SDK client: %w", err)
	}

	httpServer := server.NewStreamableHTTPServer(
		newServer(client),
		server.WithEndpointPath("/mcp"),
	)

	return httpServer, nil
}
