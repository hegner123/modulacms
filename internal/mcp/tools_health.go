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

	srv.AddTool(
		mcp.NewTool("get_metrics",
			mcp.WithDescription("Get server metrics snapshot (admin only)."),
		),
		handleGetMetrics(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_environment",
			mcp.WithDescription("Get server environment metadata."),
		),
		handleGetEnvironment(backend),
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

func handleGetMetrics(backend HealthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.GetMetrics(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetEnvironment(backend HealthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.GetEnvironment(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
