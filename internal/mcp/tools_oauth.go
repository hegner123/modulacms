package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerOAuthTools(srv *server.MCPServer, backend OAuthBackend) {
	srv.AddTool(
		mcp.NewTool("list_users_oauth",
			mcp.WithDescription("List all user OAuth connections."),
		),
		handleListUsersOAuth(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_user_oauth",
			mcp.WithDescription("Get a single user OAuth connection by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("user OAuth ID (ULID)")),
		),
		handleGetUserOAuth(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_user_oauth",
			mcp.WithDescription("Create a new user OAuth connection."),
			mcp.WithString("user_id", mcp.Description("user ID to associate")),
			mcp.WithString("oauth_provider", mcp.Required(), mcp.Description("OAuth provider name (e.g. 'google', 'github', 'azure')")),
			mcp.WithString("oauth_provider_user_id", mcp.Required(), mcp.Description("user ID from the OAuth provider")),
			mcp.WithString("access_token", mcp.Required(), mcp.Description("OAuth access token")),
			mcp.WithString("refresh_token", mcp.Required(), mcp.Description("OAuth refresh token")),
			mcp.WithString("token_expires_at", mcp.Required(), mcp.Description("Token expiration timestamp")),
		),
		handleCreateUserOAuth(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_user_oauth",
			mcp.WithDescription("Update a user OAuth connection (refresh tokens)."),
			mcp.WithString("id", mcp.Required(), mcp.Description("user OAuth ID (ULID)")),
			mcp.WithString("access_token", mcp.Required(), mcp.Description("New OAuth access token")),
			mcp.WithString("refresh_token", mcp.Required(), mcp.Description("New OAuth refresh token")),
			mcp.WithString("token_expires_at", mcp.Required(), mcp.Description("New token expiration timestamp")),
		),
		handleUpdateUserOAuth(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_user_oauth",
			mcp.WithDescription("Delete a user OAuth connection by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("user OAuth ID (ULID)")),
		),
		handleDeleteUserOAuth(backend),
	)
}

func handleListUsersOAuth(backend OAuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListUsersOAuth(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetUserOAuth(backend OAuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetUserOAuth(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateUserOAuth(backend OAuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		provider, err := req.RequireString("oauth_provider")
		if err != nil {
			return mcp.NewToolResultError("oauth_provider is required"), nil
		}
		providerUserID, err := req.RequireString("oauth_provider_user_id")
		if err != nil {
			return mcp.NewToolResultError("oauth_provider_user_id is required"), nil
		}
		accessToken, err := req.RequireString("access_token")
		if err != nil {
			return mcp.NewToolResultError("access_token is required"), nil
		}
		refreshToken, err := req.RequireString("refresh_token")
		if err != nil {
			return mcp.NewToolResultError("refresh_token is required"), nil
		}
		tokenExpiresAt, err := req.RequireString("token_expires_at")
		if err != nil {
			return mcp.NewToolResultError("token_expires_at is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"user_id":                optionalStrPtr(req, "user_id"),
			"oauth_provider":         provider,
			"oauth_provider_user_id": providerUserID,
			"access_token":           accessToken,
			"refresh_token":          refreshToken,
			"token_expires_at":       tokenExpiresAt,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateUserOAuth(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateUserOAuth(backend OAuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		accessToken, err := req.RequireString("access_token")
		if err != nil {
			return mcp.NewToolResultError("access_token is required"), nil
		}
		refreshToken, err := req.RequireString("refresh_token")
		if err != nil {
			return mcp.NewToolResultError("refresh_token is required"), nil
		}
		tokenExpiresAt, err := req.RequireString("token_expires_at")
		if err != nil {
			return mcp.NewToolResultError("token_expires_at is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"user_oauth_id":    id,
			"access_token":     accessToken,
			"refresh_token":    refreshToken,
			"token_expires_at": tokenExpiresAt,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateUserOAuth(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteUserOAuth(backend OAuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteUserOAuth(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
