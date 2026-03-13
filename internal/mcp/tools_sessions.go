package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSessionTools(srv *server.MCPServer, backend SessionBackend) {
	srv.AddTool(
		mcp.NewTool("list_sessions",
			mcp.WithDescription("List all active sessions."),
		),
		handleListSessions(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_session",
			mcp.WithDescription("Get a single session by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Session ID (ULID)")),
		),
		handleGetSession(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_session",
			mcp.WithDescription("Update an existing session."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Session ID (ULID)")),
			mcp.WithString("user_id", mcp.Description("User ID")),
			mcp.WithString("expires_at", mcp.Description("Expiration timestamp (RFC3339)")),
			mcp.WithString("last_access", mcp.Description("Last access timestamp")),
			mcp.WithString("ip_address", mcp.Description("IP address")),
			mcp.WithString("user_agent", mcp.Description("User agent string")),
			mcp.WithString("session_data", mcp.Description("Session data")),
		),
		handleUpdateSession(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_session",
			mcp.WithDescription("Delete a session by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Session ID (ULID)")),
		),
		handleDeleteSession(backend),
	)
}

func handleListSessions(backend SessionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListSessions(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetSession(backend SessionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetSession(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateSession(backend SessionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"session_id":   id,
			"user_id":      optionalStrPtr(req, "user_id"),
			"expires_at":   req.GetString("expires_at", ""),
			"last_access":  optionalStrPtr(req, "last_access"),
			"ip_address":   optionalStrPtr(req, "ip_address"),
			"user_agent":   optionalStrPtr(req, "user_agent"),
			"session_data": optionalStrPtr(req, "session_data"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateSession(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteSession(backend SessionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteSession(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
