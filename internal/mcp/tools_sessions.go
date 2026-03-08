package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func registerSessionTools(srv *server.MCPServer, client *modula.Client) {
	srv.AddTool(
		mcp.NewTool("list_sessions",
			mcp.WithDescription("List all active sessions."),
		),
		handleListSessions(client),
	)

	srv.AddTool(
		mcp.NewTool("get_session",
			mcp.WithDescription("Get a single session by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Session ID (ULID)")),
		),
		handleGetSession(client),
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
		handleUpdateSession(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_session",
			mcp.WithDescription("Delete a session by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Session ID (ULID)")),
		),
		handleDeleteSession(client),
	)
}

func handleListSessions(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Sessions.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetSession(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Sessions.Get(ctx, modula.SessionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateSession(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params := modula.UpdateSessionParams{
			SessionID:   modula.SessionID(id),
			UserID:      optionalIDPtr[modula.UserID](req, "user_id"),
			ExpiresAt:   modula.Timestamp(req.GetString("expires_at", "")),
			LastAccess:  optionalStrPtr(req, "last_access"),
			IpAddress:   optionalStrPtr(req, "ip_address"),
			UserAgent:   optionalStrPtr(req, "user_agent"),
			SessionData: optionalStrPtr(req, "session_data"),
		}
		result, err := client.Sessions.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteSession(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Sessions.Remove(ctx, modula.SessionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
