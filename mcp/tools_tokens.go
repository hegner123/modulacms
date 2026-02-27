package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerTokenTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("list_tokens",
			mcp.WithDescription("List all authentication tokens."),
		),
		handleListTokens(client),
	)

	srv.AddTool(
		mcp.NewTool("get_token",
			mcp.WithDescription("Get a single token by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Token ID (ULID)")),
		),
		handleGetToken(client),
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
		handleCreateToken(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_token",
			mcp.WithDescription("Delete a token by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Token ID (ULID)")),
		),
		handleDeleteToken(client),
	)
}

func handleListTokens(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Tokens.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetToken(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Tokens.Get(ctx, modulacms.TokenID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateToken(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.CreateTokenParams{
			UserID:    optionalIDPtr[modulacms.UserID](req, "user_id"),
			TokenType: tokenType,
			Token:     token,
			IssuedAt:  issuedAt,
			ExpiresAt: modulacms.Timestamp(expiresAt),
			Revoked:   req.GetBool("revoked", false),
		}
		result, err := client.Tokens.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteToken(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Tokens.Delete(ctx, modulacms.TokenID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
