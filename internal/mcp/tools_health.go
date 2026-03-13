package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerHealthTools(srv *server.MCPServer, backend HealthBackend) {
	srv.AddTool(
		mcp.NewTool("health",
			mcp.WithDescription("Check overall server health status."),
		),
		handleHealth(backend),
	)
}

func handleHealth(backend HealthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.Health(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
