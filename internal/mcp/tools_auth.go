package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAuthTools(srv *server.MCPServer, backend AuthBackend) {
	srv.AddTool(
		mcp.NewTool("register_user",
			mcp.WithDescription("Register a new user account."),
			mcp.WithString("username", mcp.Required(), mcp.Description("Username")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Display name")),
			mcp.WithString("email", mcp.Required(), mcp.Description("Email address")),
			mcp.WithString("password", mcp.Required(), mcp.Description("Password")),
		),
		handleRegisterUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("request_password_reset",
			mcp.WithDescription("Send a password reset email to the specified address."),
			mcp.WithString("email", mcp.Required(), mcp.Description("Email address to send the reset link to")),
		),
		handleRequestPasswordReset(backend),
	)
}

func handleRegisterUser(backend AuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		username, err := req.RequireString("username")
		if err != nil {
			return mcp.NewToolResultError("username is required"), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		email, err := req.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError("email is required"), nil
		}
		password, err := req.RequireString("password")
		if err != nil {
			return mcp.NewToolResultError("password is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"username": username,
			"name":     name,
			"email":    email,
			"password": password,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.RegisterUser(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleRequestPasswordReset(backend AuthBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, err := req.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError("email is required"), nil
		}
		data, err := backend.RequestPasswordReset(ctx, email)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
