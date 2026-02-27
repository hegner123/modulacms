package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerHealthTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("health",
			mcp.WithDescription("Check overall server health status."),
		),
		handleHealth(client),
	)
}

func handleHealth(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Health.Check(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}
