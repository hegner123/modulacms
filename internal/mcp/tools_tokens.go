package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTokenTools(srv *server.MCPServer, backend TokenBackend) {
	srv.AddTool(
		mcp.NewTool("list_tokens",
			mcp.WithDescription("List all authentication tokens."),
		),
		handleListTokens(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_token",
			mcp.WithDescription("Get a single token by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Token ID (ULID)")),
		),
		handleGetToken(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_token",
			mcp.WithDescription("Create a new authentication token."),
			mcp.WithString("token_type", mcp.Required(), mcp.Description("Token type (e.g. 'api', 'password_reset')")),
			mcp.WithString("token", mcp.Required(), mcp.Description("Token value")),
			mcp.WithString("issued_at", mcp.Required(), mcp.Description("Issued at timestamp")),
			mcp.WithString("expires_at", mcp.Required(), mcp.Description("Expiration timestamp (RFC3339)")),
			mcp.WithString("user_id", mcp.Description("User ID to associate with the token")),
			mcp.WithBoolean("revoked", mcp.Description("Whether the token is revoked (default false)")),
		),
		handleCreateToken(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_token",
			mcp.WithDescription("Delete a token by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Token ID (ULID)")),
		),
		handleDeleteToken(backend),
	)
}

func handleListTokens(backend TokenBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListTokens(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetToken(backend TokenBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetToken(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateToken(backend TokenBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tokenType, err := req.RequireString("token_type")
		if err != nil {
			return mcp.NewToolResultError("token_type is required"), nil
		}
		token, err := req.RequireString("token")
		if err != nil {
			return mcp.NewToolResultError("token is required"), nil
		}
		issuedAt, err := req.RequireString("issued_at")
		if err != nil {
			return mcp.NewToolResultError("issued_at is required"), nil
		}
		expiresAt, err := req.RequireString("expires_at")
		if err != nil {
			return mcp.NewToolResultError("expires_at is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"user_id":    optionalStrPtr(req, "user_id"),
			"token_type": tokenType,
			"token":      token,
			"issued_at":  issuedAt,
			"expires_at": expiresAt,
			"revoked":    req.GetBool("revoked", false),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateToken(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteToken(backend TokenBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteToken(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
