package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerActivityTools(srv *server.MCPServer, backend ActivityBackend) {
	srv.AddTool(
		mcp.NewTool("list_recent_activity",
			mcp.WithDescription("List recent CMS activity (audit events)."),
			mcp.WithNumber("limit", mcp.Description("Max events to return (default 20, max 100)"), mcp.DefaultNumber(20)),
		),
		handleListRecentActivity(backend),
	)
}

func handleListRecentActivity(backend ActivityBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		data, err := backend.ListRecentActivity(ctx, limit)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
