package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSearchTools(srv *server.MCPServer, backend SearchBackend) {
	srv.AddTool(
		mcp.NewTool("search_content",
			mcp.WithDescription("Search published content by keyword."),
			mcp.WithString("q", mcp.Required(), mcp.Description("Search query string")),
			mcp.WithNumber("limit", mcp.Description("Max results to return (default 20)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of results to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleSearchContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("rebuild_search_index",
			mcp.WithDescription("Rebuild the full-text search index. This reindexes all published content."),
		),
		handleRebuildSearchIndex(backend),
	)
}

func handleSearchContent(backend SearchBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		q, err := req.RequireString("q")
		if err != nil {
			return mcp.NewToolResultError("q is required"), nil
		}
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.SearchContent(ctx, q, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleRebuildSearchIndex(backend SearchBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.RebuildSearchIndex(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
