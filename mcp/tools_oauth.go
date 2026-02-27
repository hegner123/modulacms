package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerOAuthTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("list_users_oauth",
			mcp.WithDescription("List all user OAuth connections."),
		),
		handleListUsersOAuth(client),
	)

	srv.AddTool(
		mcp.NewTool("get_user_oauth",
			mcp.WithDescription("Get a single user OAuth connection by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User OAuth ID (ULID)")),
		),
		handleGetUserOAuth(client),
	)

	srv.AddTool(
		mcp.NewTool("create_user_oauth",
			mcp.WithDescription("Create a new user OAuth connection."),
			mcp.WithString("user_id", mcp.Description("User ID to associate")),
			mcp.WithString("oauth_provider", mcp.Required(), mcp.Description("OAuth provider name (e.g. 'google', 'github', 'azure')")),
			mcp.WithString("oauth_provider_user_id", mcp.Required(), mcp.Description("User ID from the OAuth provider")),
			mcp.WithString("access_token", mcp.Required(), mcp.Description("OAuth access token")),
			mcp.WithString("refresh_token", mcp.Required(), mcp.Description("OAuth refresh token")),
			mcp.WithString("token_expires_at", mcp.Required(), mcp.Description("Token expiration timestamp")),
		),
		handleCreateUserOAuth(client),
	)

	srv.AddTool(
		mcp.NewTool("update_user_oauth",
			mcp.WithDescription("Update a user OAuth connection (refresh tokens)."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User OAuth ID (ULID)")),
			mcp.WithString("access_token", mcp.Required(), mcp.Description("New OAuth access token")),
			mcp.WithString("refresh_token", mcp.Required(), mcp.Description("New OAuth refresh token")),
			mcp.WithString("token_expires_at", mcp.Required(), mcp.Description("New token expiration timestamp")),
		),
		handleUpdateUserOAuth(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_user_oauth",
			mcp.WithDescription("Delete a user OAuth connection by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User OAuth ID (ULID)")),
		),
		handleDeleteUserOAuth(client),
	)
}

func handleListUsersOAuth(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.UsersOauth.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetUserOAuth(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.UsersOauth.Get(ctx, modulacms.UserOauthID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateUserOAuth(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.CreateUserOauthParams{
			UserID:              optionalIDPtr[modulacms.UserID](req, "user_id"),
			OauthProvider:       provider,
			OauthProviderUserID: providerUserID,
			AccessToken:         accessToken,
			RefreshToken:        refreshToken,
			TokenExpiresAt:      tokenExpiresAt,
		}
		result, err := client.UsersOauth.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateUserOAuth(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.UpdateUserOauthParams{
			UserOauthID:    modulacms.UserOauthID(id),
			AccessToken:    accessToken,
			RefreshToken:   refreshToken,
			TokenExpiresAt: tokenExpiresAt,
		}
		result, err := client.UsersOauth.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteUserOAuth(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.UsersOauth.Delete(ctx, modulacms.UserOauthID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
