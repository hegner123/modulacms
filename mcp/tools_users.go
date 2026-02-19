package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerUserTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("whoami",
			mcp.WithDescription("Get the currently authenticated user's profile. Returns user_id, username, name, email, role, and timestamps. Use this to get your user_id for content authoring."),
		),
		handleWhoami(client),
	)

	srv.AddTool(
		mcp.NewTool("list_users",
			mcp.WithDescription("List all users."),
		),
		handleListUsers(client),
	)

	srv.AddTool(
		mcp.NewTool("get_user",
			mcp.WithDescription("Get a single user by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
		),
		handleGetUser(client),
	)

	srv.AddTool(
		mcp.NewTool("create_user",
			mcp.WithDescription("Create a new user. All fields are required."),
			mcp.WithString("username", mcp.Required(), mcp.Description("Username")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Display name")),
			mcp.WithString("email", mcp.Required(), mcp.Description("Email address")),
			mcp.WithString("password", mcp.Required(), mcp.Description("Password")),
			mcp.WithString("role", mcp.Required(), mcp.Description("Role name (e.g. admin, editor, viewer)")),
		),
		handleCreateUser(client),
	)

	srv.AddTool(
		mcp.NewTool("update_user",
			mcp.WithDescription("Update an existing user. This is a full replacement — provide all fields you want to keep. Omitting a field sets it to empty. Password can be omitted to keep the current password."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
			mcp.WithString("username", mcp.Description("Username")),
			mcp.WithString("name", mcp.Description("Display name")),
			mcp.WithString("email", mcp.Description("Email address")),
			mcp.WithString("password", mcp.Description("New password (omit to keep current)")),
			mcp.WithString("role", mcp.Description("Role name")),
		),
		handleUpdateUser(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_user",
			mcp.WithDescription("Delete a user by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
		),
		handleDeleteUser(client),
	)
}

func handleWhoami(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Auth.Me(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleListUsers(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Users.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetUser(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Users.Get(ctx, modulacms.UserID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateUser(client *modulacms.Client) server.ToolHandlerFunc {
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
		role, err := req.RequireString("role")
		if err != nil {
			return mcp.NewToolResultError("role is required"), nil
		}
		params := modulacms.CreateUserParams{
			Username: username,
			Name:     name,
			Email:    modulacms.Email(email),
			Password: password,
			Role:     role,
		}
		result, err := client.Users.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateUser(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params := modulacms.UpdateUserParams{
			UserID:   modulacms.UserID(id),
			Username: req.GetString("username", ""),
			Name:     req.GetString("name", ""),
			Email:    modulacms.Email(req.GetString("email", "")),
			Password: req.GetString("password", ""),
			Role:     req.GetString("role", ""),
		}
		result, err := client.Users.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteUser(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Users.Delete(ctx, modulacms.UserID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
